# Saga

A saga is a sequence of local transactions across services, where each step has a compensating action that semantically undoes it. Use sagas when you need atomic-feeling business outcomes across service boundaries that 2PC/XA can't span — different DBs, async messaging, third-party APIs, or simply database-per-service microservices.

Origin: **Garcia-Molina & Salem, "Sagas," ACM SIGMOD 1987**. The idea predates microservices by 30 years; the original paper targeted long-lived transactions inside a single DBMS. Microservices recycled the name but changed the problem — now the application owns durability of the saga state, not a DBMS. "Long-running transaction" is the wrong mental model. Think **workflow with compensations**.

## Why not 2PC

| Constraint | 2PC / XA | Saga |
|------------|----------|------|
| Heterogeneous stores | Requires XA on every participant | None — each service uses its own DB |
| Failure of coordinator | Locks held until recovery | Forward or backward recovery, no global locks |
| Latency under contention | Head-of-line blocking on locks | None at the data layer |
| Atomicity | True ACID | Semantic atomicity (compensations) |
| Isolation | Yes | **No** — concurrent sagas can interleave |

Sagas trade isolation for availability and decoupling. That trade is the whole game.

## ACD (no I)

A saga is **ACD**: Atomic (semantically), Consistent, Durable — but **not Isolated**. Concurrent sagas produce three anomaly classes:

| Anomaly | Mechanism | Countermeasure |
|---------|-----------|----------------|
| Lost update | Saga B overwrites A's update without seeing it | Commutative ops (`balance += x`); reread+CAS |
| Dirty read | B reads A's tentative state that later compensates | Semantic lock (`status = PENDING`); pessimistic view |
| Fuzzy / non-repeatable | Two reads in one saga return different values | By-value (snapshot data into the saga); version file |

Five Richardson countermeasures: **semantic lock**, **commutative updates**, **pessimistic view** (reorder so risky updates land *after* the pivot), **reread value** (optimistic concurrency), **version file** (replay log). Pick per step, not globally.

## Step taxonomy (Richardson)

Every saga step is exactly one of:

| Type | Property | Phase |
|------|----------|-------|
| **Compensatable** | Has a semantic undo | Before pivot |
| **Pivot** | Go/no-go boundary. Neither retriable nor compensatable. Exactly one per saga. | Middle |
| **Retriable** | Idempotent; cannot fail for business reasons (only transient) | After pivot |

Order is fixed: `compensatable* → pivot → retriable*`. Past the pivot, the saga is forward-only — failure routes to retry/DLQ/on-call, never to compensation. Place irreversible side effects (sent emails, charged cards on a no-refund processor, physical shipment) **as the pivot or after**.

## Recovery

| Failure type | Recovery | Example |
|--------------|----------|---------|
| Transient (timeout, 5xx, broker blip) | **Forward** — retry with exponential backoff + jitter | Network drop on a retriable step |
| Business rule | **Backward** — run `Cn … C1` in reverse | Insufficient inventory, card declined |

Idempotency is non-negotiable in either mode: every activity *will* be retried by the engine, the broker, or your own code.

## Orchestration vs Choreography

```
Orchestration                          Choreography
                                       
  ┌──────────────┐                     ┌──────┐  evt   ┌──────┐
  │ Orchestrator │                     │  A   │ ─────▶ │  B   │
  └──┬──┬──┬──┬──┘                     └──────┘        └──┬───┘
     │  │  │  │                          ▲   evt          │ evt
     ▼  ▼  ▼  ▼                          │                ▼
  [A][B][C][D] participants            [D] ◀────────── [C]
```

| | Choreography (events) | Orchestration (commands) |
|---|---|---|
| Control | Each service reacts to peer events | Central orchestrator issues commands |
| Coupling | Loose | Logically centralized; participants don't know each other |
| Cyclic deps | Easy to introduce | Avoided by design |
| Observability | Hard — flow is emergent | Single state machine, single trace |
| New steps | Confusing once topology grows | Easy to extend |
| SPOF | None | Orchestrator (mitigate with HA / durable engine) |
| Sweet spot | 2–4 participants, linear flows, fan-out | 5+ participants, branches, conditionals, timers |

Hybrid is fine: an orchestrated saga whose leaf steps fan out via pub/sub (Vernon's "process manager"). In Dapr terms: a Workflow that publishes to topics inside an activity.

## Dapr Workflows

GA since **Dapr v1.15** (Feb 2025). The engine is `dapr/durabletask-go` (a fork of Microsoft's Durable Task), embedded inside the sidecar — no separate cluster. v1.17 (Feb 2026) added **workflow versioning** and **state retention policies**. Backend uses the Dapr Scheduler service for reminders/timers and the configured state store for history.

```
[App] ──┐                        ┌──▶ [Activity worker (your Go code)]
        ▼                        │
   [daprd sidecar]               │
        │  workflow engine ──────┘
        │   (durabletask-go)
        ├──▶ State store (Postgres/Redis/SQLite/etc.) — durable history
        └──▶ Scheduler service — durable timers/reminders
```

State store must support **transactions + ETag + actors**: Redis, PostgreSQL, MySQL/MariaDB, MongoDB, SQLite, etcd v2, in-memory. SQL Server and Oracle are not supported for workflows.

### Go SDK — current (May 2026)

The legacy `github.com/dapr/go-sdk/workflow` package is deprecated. Use:

- `github.com/dapr/durabletask-go/workflow` — `Registry`, `WorkflowContext`, `ActivityContext`
- `github.com/dapr/go-sdk/client` — `client.NewWorkflowClient()` for management

```go
import (
    "github.com/dapr/durabletask-go/workflow"
    "github.com/dapr/go-sdk/client"
)

r := workflow.NewRegistry()
r.AddWorkflow(OrderSaga)
r.AddActivity(ReserveStock)
r.AddActivity(ReleaseStock)
r.AddActivity(ChargeCard)
r.AddActivity(RefundCard)
r.AddActivity(ShipOrder)

wfc, _ := client.NewWorkflowClient()
wfc.StartWorker(ctx, r)

id, _ := wfc.ScheduleWorkflow(ctx, "OrderSaga",
    workflow.WithInput(req),
    workflow.WithInstanceID("order-"+req.ID))

md, _ := wfc.WaitForWorkflowCompletion(ctx, id)
```

Lifecycle methods on `wfc`: `FetchWorkflowMetadata`, `RaiseEvent`, `SuspendWorkflow`, `ResumeWorkflow`, `TerminateWorkflow`, `PurgeWorkflow`. CLI equivalents under `dapr workflow ...`.

### Saga in Go — compensation pattern

```go
func OrderSaga(ctx *workflow.WorkflowContext) (any, error) {
    var in OrderInput
    if err := ctx.GetInput(&in); err != nil {
        return nil, err
    }

    var compensations []func()
    runCompensations := func() {
        for i := len(compensations) - 1; i >= 0; i-- {
            compensations[i]()
        }
    }

    var resv ReservationOK
    if err := ctx.CallActivity(ReserveStock,
        workflow.WithActivityInput(in)).Await(&resv); err != nil {
        return nil, err
    }
    compensations = append(compensations, func() {
        _ = ctx.CallActivity(ReleaseStock,
            workflow.WithActivityInput(resv),
            workflow.WithActivityRetryPolicy(retry)).Await(nil)
    })

    var pay PaymentOK
    if err := ctx.CallActivity(ChargeCard,
        workflow.WithActivityInput(in)).Await(&pay); err != nil {
        runCompensations()
        return nil, err
    }
    compensations = append(compensations, func() {
        _ = ctx.CallActivity(RefundCard,
            workflow.WithActivityInput(pay)).Await(nil)
    })

    // ShipOrder is the pivot — past this point, forward-only.
    if err := ctx.CallActivity(ShipOrder,
        workflow.WithActivityInput(in)).Await(nil); err != nil {
        runCompensations()
        return nil, err
    }
    return OrderResult{OK: true}, nil
}
```

Activities are plain Go functions:

```go
func ReserveStock(ctx workflow.ActivityContext) (any, error) {
    var in OrderInput
    if err := ctx.GetInput(&in); err != nil {
        return nil, err
    }
    // Real I/O lives here. Must be idempotent — engine retries on replay.
    return repo.Reserve(ctx.Context(), in.SKU, in.Qty)
}
```

### Determinism rules

Workflow code runs many times during replay. It **must** be deterministic.

| Forbidden in workflow body | Use instead |
|----------------------------|-------------|
| `time.Now()` | `ctx.CurrentTimeUTC()` |
| `rand`, `uuid.New()` | Pass values in via `WithInput`, or generate inside an activity |
| `os.Getenv`, file I/O, network | Activities only |
| `go func()`, channels | Sequential code; use `CallActivity` for parallelism via `Task` futures |
| Map iteration (unsorted) | Sort keys explicitly |
| Parent `context.Context` for control flow | `ctx` provided by the SDK |

Side effects belong in activities. Workflows orchestrate; they don't do.

### Versioning (v1.17+)

Two strategies for evolving a workflow with in-flight instances:

```go
// Named versions — old instances finish on v1, new instances start on v2.
r.AddVersionedWorkflow("OrderSaga", false, OrderSagaV1)
r.AddVersionedWorkflow("OrderSaga", true,  OrderSagaV2) // isLatest=true

// Patching — branch on the patch flag inside the same function.
if ctx.IsPatched("add-fraud-check") {
    if err := ctx.CallActivity(FraudCheck, ...).Await(nil); err != nil {
        runCompensations()
        return nil, err
    }
}
```

Pre-1.17, the only safe options were drain-then-deploy or branching on `IsReplaying()`. Don't rename activities or change input/output schemas without a versioning strategy — old replay history will fail to deserialize.

### History truncation

Every activity completion appends to the workflow history. The full history replays on every event. Past ~10k events, call `ctx.ContinueAsNew(newInput, workflow.WithKeepUnprocessedEvents())` to start a fresh history with rolled-up state. Also call `PurgeWorkflow` (or set a retention policy in v1.17+) after completion — history persists otherwise.

## Choreography via Dapr pub/sub

When a Workflow engine is overkill, use pub/sub + state store. Each service subscribes to upstream events, performs its local step, and publishes either the next event or a failure event:

```
OrderCreated ─▶ stock.Reserve  ─▶ StockReserved ─▶ payments.Charge ─▶ Charged ─▶ ship.Dispatch
                       │                                  │
                       ▼ on fail                          ▼ on fail
                StockFailed ◀─ … ◀─ ChargeFailed ─▶ stock.Release (compensator)
```

```go
// Subscribe to the upstream event, emit the next event or a failure event.
s.AddTopicEventHandler(&common.Subscription{
    PubsubName: "pubsub", Topic: "stock.reserved", Route: "/stock-reserved",
}, func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
    if err := payments.Charge(ctx, decode(e)); err != nil {
        _ = client.PublishEvent(ctx, "pubsub", "payment.failed", failPayload(e))
        return false, nil // ack — failure is handled via the failure event
    }
    return false, client.PublishEvent(ctx, "pubsub", "payment.charged", okPayload(e))
})
```

| Choose pub/sub when | Choose Workflow when |
|---------------------|----------------------|
| Few steps, linear flow | Branches, parallel fan-out |
| Teams own their service end-to-end | Single team owns the flow |
| No human approvals or long timers | Wait for external event / approval / sleep |
| Cross-language services | Single-language is fine |
| Already have strong eventing posture | Want a replayable state machine |

## Library landscape

| Tool | Coordinator | Infra | Status (May 2026) | Use when |
|------|------------|-------|-------------------|----------|
| **dapr/durabletask-go** + go-sdk/client | Orchestration | Sidecar + state store | GA in 1.15, active | Already on Dapr — default |
| **temporal-io/sdk-go** | Orchestration | Temporal cluster | Mature, very active | Need replay tooling, advanced versioning, very high scale |
| **dtm-labs/dtm** | Orchestration | Standalone DTM server + DB | Active (~10.9k★) | Polyglot stack; strong barrier idempotency |
| **restate.dev sdk-go** | Orchestration | Restate broker | Active | Lightweight ops, code-first |
| **itimofeev/go-saga** | In-process orchestration | None | Maintenance mode | Single-service saga, no cross-service coordination |
| **lysu/go-saga** | In-process orchestration | None | **Abandoned (~2017)** | Don't adopt |
| **eventuate-tram-sagas** | Orchestration | JVM only | JVM-only | Skip in Go shops |
| **NATS JetStream + DIY** | Choreography | NATS | Pattern, not a library | Already on NATS, willing to own retries/observability |
| **Outbox + Debezium + Kafka** | Choreography | Postgres + Kafka + Debezium | Pattern | Strong CDC posture, want at-least-once without dual-writes |
| **encore.dev** | Choreography (outbox helper) | Encore platform | Active | Code-as-infra; no orchestration primitive |

**Default for this repo: dapr/durabletask-go.** Reach for Temporal only if Dapr Workflow limits bite (heavy versioning, very large history, advanced query). Use Dapr pub/sub + outbox for choreographed sagas that don't need an orchestrator. Avoid `lysu/go-saga`.

## Pitfalls & production checklist

### Idempotency

Every activity will run more than once — engine replay, broker redelivery, your own retries. The "exactly-once" guarantee in Dapr/Temporal is for *workflow state transitions*, not side effects.

- Pass an idempotency key into every activity (e.g., `ctx.ID() + "/" + stepName`).
- Persist `(idempotency_key, response)` in the **same DB transaction** as the side effect. Redis is a cache, not a system of record — flushes have caused duplicate charges in real incidents.
- TTL must exceed the longest retry window plus broker redelivery (24–72h typical; days for long sagas). Stripe's idempotency-key design is the canonical reference.
- Compensations must also be idempotent — they are retried too.

### Outbox / Inbox

Publishing an event after `tx.Commit()` is the dual-write bug. If the process dies between commit and publish, the saga loses a step or a compensation. **Write the event to an `outbox` table inside the business transaction**; a relay (Debezium CDC tailing the WAL, or a polling worker) publishes from the outbox to the broker. Receivers dedupe via an inbox table keyed by event ID.

A saga without an outbox can permanently lose compensations and leave inventory reserved forever.

### Retries

Exponential backoff with **full jitter**: `sleep = random(0, base * 2^n)`. Without jitter, retries synchronize into thundering herds. Apply a **retry budget** (e.g., 10% of request volume) per downstream so a slow dependency doesn't multiply load on an already-failing service.

```go
retry := &workflow.RetryPolicy{
    MaxAttempts:          5,
    InitialRetryInterval: 500 * time.Millisecond,
    BackoffCoefficient:   2.0,
    MaxRetryInterval:     30 * time.Second,
    RetryTimeout:         5 * time.Minute,
}
ctx.CallActivity(ReserveStock,
    workflow.WithActivityInput(in),
    workflow.WithActivityRetryPolicy(retry))
```

### Compensation that fails

Bound retries → DLQ → human queue → page on-call. Don't silently retry forever. For sagas with hard deadlines, mark `ABANDONED` after T and emit to ops; otherwise zombie state accumulates.

### Observability

- W3C `traceparent` propagates through Dapr automatically (pub/sub, service invocation, workflow). Verify end-to-end — custom middleware can drop it.
- Saga metrics worth dashboarding: success rate, p95/p99 duration, compensation rate (spikes usually mean a downstream regressed), abandoned rate, step-level failure counts.
- Every log line carries `saga_id`, `step_name`, `correlation_id`. The workflow instance ID is your saga ID.
- Replay = re-run the workflow from history against staging. Dapr exposes history via the management API.

### Testing

- Unit-test each compensation in isolation against a fake activity.
- Chaos: kill the sidecar mid-saga, drop random activities, inject 500s.
- Contract tests (Pact) on event payloads catch producer/consumer drift.
- End-to-end with `testcontainers-go` running Postgres + Redis + Dapr.

### Anti-patterns

- **Distributed monolith.** If two services always saga together, they're one bounded context. Merge them.
- **Saga as a transaction.** It isn't — no isolation, only eventual consistency.
- **Saga where a local TX would do.** Two tables in the same DB don't need a saga.
- **Synchronous orchestrator blocking on HTTP for 30s.** Defeats the durability point. Make every step async via activity calls or pub/sub.
- **Non-deterministic workflow body.** `time.Now`, `rand`, goroutines, env vars in the workflow function — instant replay failures.
- **Renaming activities or changing payload schemas without versioning.** Breaks in-flight instances.

## Repo fit — greeter

Stack: go-kratos v2 layered (`service → biz → data`), Wire DI, ent ORM (Postgres), Dapr sidecar with `secretstore` + `pubsub` components.

**Where saga code lives:**

| Layer | Saga responsibility |
|-------|---------------------|
| `internal/service/` | Inbound RPC starts/queries the saga via `wfc.ScheduleWorkflow` / `FetchWorkflowMetadata` |
| `internal/biz/` | Workflow function (`OrderSaga`) and activity functions. Workflow body is pure; activities call `biz` interfaces. Stays free of `data` imports — kratos rule. |
| `internal/data/` | Activity implementations call into `data` repos for DB writes (with idempotency-key dedupe in the same tx) and `EventRepo.Publish` (outbox-backed) |
| `internal/server/` | Workflow worker startup (`wfc.StartWorker(ctx, registry)`) registered as a kratos `transport.Server` so the kratos lifecycle owns it |

**New components to declare in `deploy/k8s/base/infra/dapr/`:**

```yaml
# Workflow needs a transactional state store with actors enabled.
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: workflowstore
spec:
  type: state.postgresql
  version: v2
  metadata:
    - name: connectionString
      secretKeyRef: { name: db-secret, key: dsn }
    - name: actorStateStore
      value: "true"
```

The existing `pubsub` component covers choreography; reuse it. The existing Postgres can host both ent business tables and the workflow history (separate schema recommended).

**Wire wiring:** add a `WorkflowProviderSet` to `cmd/server/wire.go` that produces the `*workflow.Registry`, registers workflows/activities defined in `biz`, and exposes the worker as a kratos `transport.Server`. The kratos `app.New(...)` lifecycle then starts/stops the worker alongside HTTP and gRPC.

**Idempotency:** add an `idempotency_keys` table managed via ent (`key` PK, `response_json`, `created_at`). Every activity that has a side effect upserts `(ctx.ID()+"/"+stepName, resultJSON)` in the same `ent.Tx` as the business write.

**Outbox:** add an `outbox` table; `EventRepo.Publish` writes to `outbox` inside the caller's `ent.Tx`. A small relay loop (or Debezium against the WAL) reads `outbox` rows, calls `client.PublishEvent`, marks them sent. This makes choreographed compensations safe.

**First saga to migrate:** the recursive-buy / multi-step checkout flow currently in `biz` — it crosses inventory, payment, and shipping steps and is the canonical orchestrated-saga shape.

## References

- Garcia-Molina & Salem, "Sagas," ACM SIGMOD 1987
- Chris Richardson — `microservices.io/patterns/data/saga.html`, *Microservices Patterns* ch. 4
- Microsoft Learn — Saga pattern (`learn.microsoft.com/azure/architecture/patterns/saga`)
- AWS Prescriptive Guidance — Saga orchestration / choreography
- Dapr docs — `docs.dapr.io/developing-applications/building-blocks/workflow/`
- Dapr v1.15 GA blog (2025-02), v1.17 versioning blog (2026-02)
- `github.com/dapr/durabletask-go`, `github.com/dapr/go-sdk`, `github.com/dapr/quickstarts/workflows/go`
- Temporal Go samples — `github.com/temporalio/samples-go/blob/main/saga/workflow.go`
- Stripe — Idempotency keys (`stripe.com/blog/idempotency`, `brandur.org/idempotency-keys`)
- Debezium Outbox Event Router (`debezium.io/documentation/reference/stable/transformations/outbox-event-router.html`)
- AWS Builders' Library — Timeouts, retries, backoff with jitter
