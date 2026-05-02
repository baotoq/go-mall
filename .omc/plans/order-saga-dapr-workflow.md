# Plan: Order Saga via Dapr Workflow

**Status:** Draft — v0.3 (Consensus review iteration 1 — RALPLAN-DR deliberate mode)
**Date:** 2026-05-02
**Author:** /plan interview → RALPLAN-DR deliberate mode
**Scope:** Order + Payment only
**Risk:** High (distributed transactions, money flow, first cross-service flow in repo)

---

## 1. Requirements Summary

Introduce an orchestrated saga that coordinates **order creation → payment authorization → order-paid transition** as a single durable workflow, replacing today's ad-hoc client-side orchestration where the caller manually invokes `CreateOrder` then later updates status with a `payment_id`.

**Why this matters:**
- Today services don't talk to each other (`CLAUDE.md:23` "Cross-service calls: none"); payment linking is post-hoc and fragile.
- `docs/saga.md:375` already names the multi-step checkout flow as the canonical first saga to implement.
- The existing transactional outbox (`pkg/outbox/`) and `OrderCreatedEvent` on `order.created` (`app/order/internal/biz/events.go`, `app/order/internal/data/order.go:71`) provide a concrete head-start.
- Dapr Workflow runtime is documented but unused; this exercises the documented stack.

**Non-goals (this iteration):**
- Inventory/stock reservation (catalog stock untouched).
- Real payment-gateway integration (`provider` field stays informational).
- Shipping service.
- Notifications consumer.
- Cart `Checkout` RPC (new `Checkout` lives on the order service only).

---

## 2. Decisions Recap

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Saga scope | Order + Payment only | Smallest viable; proves pattern; minimal blast radius. |
| 2 | Trigger | New `order.Checkout` RPC | Explicit entry point; `CreateOrder` stays as "manual mode" escape hatch. |
| 3 | Activity invocation | Pub/sub + `WaitForExternalEvent` | Decoupled; reuses declared `pubsub` Dapr component; no cross-service imports. |
| 4 | Compensation | Retry with backoff (3 attempts, 500ms initial, 2× backoff, 30s max delay, 60s per-attempt timeout), then cancel order | **Two distinct budgets:** payment loop ~181.5s (3×60s + backoffs); MarkPaid post-pivot retry: 5 attempts, 5min RetryTimeout. |
| 5 | Idempotency | Client-provided `idempotency_key` as workflow instance ID | Stripe-style; `docs/saga.md:282-283`. |
| 6 | Workflow location | `app/order/internal/biz/saga.go`; worker in `app/order/cmd/server/` | Matches `docs/saga.md:340-348`. |
| 7 | Status retrieval | New `order.GetCheckoutStatus` RPC via `wfc.FetchWorkflowMetadata` | Clean separation; no `GetOrder` pollution. |

**Retry budget constants (single source of truth):**
```go
const (
    maxPaymentAttempts  = 3
    perAttemptTimeout   = 60 * time.Second
    paymentInitialDelay = 500 * time.Millisecond
    paymentBackoffMax   = 30 * time.Second
    markPaidRetryMax    = 5
    markPaidBudget      = 5 * time.Minute
)
// Payment loop worst-case: 3×60s + 0.5s + 1s backoffs = 181.5s ≈ 3.1min
// AC6 test bound: within 3.5min (includes scheduling overhead)
```

---

## 3. RALPLAN-DR Summary (v0.3 — post-consensus revision)

### Principles (5)

1. **Determinism in workflow body; side effects in activities.** Workflow body forbids all runtime-evaluated time calls (`time.Now`, `time.Since`, `time.Until`, `time.After`, `time.NewTicker`), `rand.*`, goroutines, `os.Getenv`, raw map iteration. Duration *constants* (`time.Second`, etc.) are allowed. Error reasons stored as `ReasonCode` enum strings, not `err.Error()`, to avoid replay non-determinism. (`docs/saga.md:187-200`)

2. **Idempotency at every boundary.** Activities pass deterministic `WithMessageID(workflowID+":pay-req:"+attempt)` to the outbox. Outbox `messageID` column has a UNIQUE constraint; `INSERT ... ON CONFLICT DO NOTHING` ensures double-replay produces exactly one outbox row. (`docs/saga.md:281-283`)

3. **Outbox before pub/sub.** All cross-service event publishes flow through `pkg/outbox/`. `PublishPaymentRequestedActivity` opens an `ent.Tx` solely to wrap the outbox insert and idempotency-key upsert; both commit atomically.

4. **Compensatable → true Pivot → Post-pivot Retriable.** `CreateOrder` and `PublishPaymentRequested` are compensatable (order can be cancelled). The true **pivot** is `payment.completed` landing in the payment service's DB (via `CompletePayment`): once that commits, money semantics apply. `MarkPaid` is **post-pivot retriable** (forward-only, retried 5×; explicitly not compensatable). (`docs/saga.md:31-41`)

5. **Additive API surface.** Existing RPCs unchanged. `PaymentUsecase.Create` extended via `CreatePaymentInput` struct (not positional args) to avoid breaking callers.

### Decision Drivers

1. **Correctness under partial failure.** Saga state must be durable; replays must not double-act.
2. **Minimal change to existing services.** Wire existing skeletons rather than rewrite.
3. **Operability.** Workflow ID == saga ID, traceparent propagated, every event tagged with `workflow_instance_id`.

### Viable Options Considered

#### Option A: Pub/sub + `WaitForExternalEvent` *(SELECTED)*
**Pros:** Most decoupled; reuses `pubsub` component; matches `docs/saga.md:75`.
**Cons:** Requires attempt-keyed event names (`payment-completed-{N}`) to prevent cross-attempt event routing race; largest implementation surface; late events for terminated workflows require DLQ rather than silent discard.
**Decision:** Accepted over Option E because the codebase has no sync gRPC between services; establishing async-first avoids baking in a coupling that is harder to remove.

#### Option B: Direct gRPC from activity
**Pros:** Compile-time safety; one fewer subscriber.
**Cons:** Breaks "zero cross-service imports" (`CLAUDE.md:23`); couples deployment lifecycle.
**Invalidation:** Creates a stronger coupling than pub/sub (shared `go.mod` dependency + synchronized deploys).

#### Option C: Dapr service invocation
**Pros:** Sync semantics with Dapr resilience.
**Cons:** Sync call inside an activity doesn't model the real "external payment processor" shape.
**Invalidation:** Doesn't represent the async payment-gateway shape the codebase grows toward.

#### Option D: Pure choreography
**Cons:** Abandons the feature premise; loses replay, durable timers, central observability.

#### Option E: gRPC for control flow + pub/sub for terminal notifications *(not selected)*
**Pros:** Eliminates cross-attempt event routing race; simpler workflow body.
**Cons:** Introduces compile-time cross-service dependency; couples deployment.
**Invalidation:** `CLAUDE.md:23` invariant; async-first direction.

---

## 4. Architecture & Data Flow

### 4.1 Event topology

```
              ┌──────────────────────────────────────────────────────┐
              │                  order service                       │
              │  ┌──────────┐    ┌──────────────────────────────┐    │
client──Checkout─▶ service │───▶│ OrderSaga workflow           │────┼──┐
              │  └──────────┘    │  CreateOrder                 │    │  │ pub: payment.requested
              │       ▲          │  loop(1..3):                 │    │  │ {workflow_id, order_id,
              │       │ poll     │    PublishPaymentRequested    │    │  │  amount, currency, attempt}
              │  ┌────┴─────┐    │    WaitForExternalEvent      │    │  ▼
              │  │GetCheckout    │    "payment-completed-{N}"    │    │ ┌────────────────────┐
              │  │  Status  │    │    "payment-failed-{N}"       │    │ │   pubsub (Redis)   │
              │  └──────────┘    │  MarkPaid (post-pivot retry)  │    │ └────────────────────┘
              │                  │  CancelOrder (compensation)   │    │  ▲                │
              │  ┌─────────────────────────────────────────────┐│    │  │                │ sub:
              │  │ event router (subscriber.go)                ││    │  │ payment.        │ payment.
              │  │ sub: payment.completed, payment.failed       ││    │  │ completed/failed│ requested
              │  │ → RaiseEvent(wfID, "pay-{done|fail}-{N}", ..)│────┘  │ {workflow_id,   │ {attempt=N}
              │  └─────────────────────────────────────────────┘│      │  pay_id, reason}│
              └──────────────────────────────────────────────────┘      │                │
                                                                         │                ▼
                               ┌──────────────────────────────────────────────────────────────┐
                               │                  payment service                             │
                               │  sub: payment.requested{attempt=N}                          │
                               │     → CreatePayment(PENDING, workflow_instance_id, attempt)  │
                               │     → UNIQUE(workflow_instance_id, attempt)                  │
                               │  RPCs (dev/test driver):                                     │
                               │     CompletePayment(id) → pub payment.completed{attempt=N}  │
                               │     FailPayment(id) → pub payment.failed{attempt=N}         │
                               └──────────────────────────────────────────────────────────────┘
```

### 4.2 Workflow body (pseudocode — deterministic)

```go
// app/order/internal/biz/saga.go
func OrderSaga(ctx *workflow.WorkflowContext) (any, error) {
    var in CheckoutInput
    if err := ctx.GetInput(&in); err != nil { return nil, err }
    workflowID := ctx.InstanceID()

    var orderID string
    if err := ctx.CallActivity(CreateOrderActivity,
        workflow.WithActivityInput(in)).Await(&orderID); err != nil {
        return CheckoutResult{State: "FAILED", Reason: "order_create_failed"}, nil
    }

    type po struct{ paymentID string; success bool; reasonCode string }
    var outcome po
    delay := paymentInitialDelay // time.Duration constant — deterministic

    for attempt := 1; attempt <= maxPaymentAttempts; attempt++ {
        if err := ctx.CallActivity(PublishPaymentRequestedActivity,
            workflow.WithActivityInput(PaymentRequestedInput{
                WorkflowID: workflowID, OrderID: orderID,
                Amount: in.TotalCents, Currency: in.Currency, Attempt: attempt,
            })).Await(nil); err != nil {
            outcome.reasonCode = "publish_failed"
        } else {
            // Attempt-keyed events prevent cross-attempt routing race.
            completedEvt := fmt.Sprintf("payment-completed-%d", attempt)
            failedEvt    := fmt.Sprintf("payment-failed-%d",    attempt)
            completed := ctx.WaitForExternalEvent(completedEvt, perAttemptTimeout)
            failed    := ctx.WaitForExternalEvent(failedEvt,    perAttemptTimeout)
            timer     := ctx.CreateTimer(perAttemptTimeout)
            // workflow.Any: confirm exact API name during Step 0.1 research.
            winner, _ := workflow.Any(ctx, completed, failed, timer).Await(nil)
            switch winner {
            case completed:
                var p PaymentOutcome
                if completed.Await(&p) == nil {
                    outcome = po{paymentID: p.PaymentID, success: true}
                }
            case failed:
                var p PaymentOutcome
                _ = failed.Await(&p)
                outcome.reasonCode = p.ReasonCode // enum, not err.Error()
            default:
                outcome.reasonCode = "timeout"
            }
        }
        if outcome.success { break }
        if attempt < maxPaymentAttempts {
            if err := ctx.CreateTimer(delay).Await(nil); err != nil { return nil, err }
            if delay < paymentBackoffMax { delay *= 2 }
        }
    }

    if !outcome.success {
        _ = ctx.CallActivity(CancelOrderActivity,
            workflow.WithActivityInput(CancelInput{OrderID: orderID, Reason: outcome.reasonCode})).Await(nil)
        return CheckoutResult{State: "FAILED", OrderID: orderID, Reason: "payment_exhausted", LastError: outcome.reasonCode}, nil
    }

    // Post-pivot retriable: payment.completed in payment DB is the irreversible pivot.
    retry := &workflow.RetryPolicy{
        MaxAttempts: markPaidRetryMax, InitialRetryInterval: paymentInitialDelay,
        BackoffCoefficient: 2.0, MaxRetryInterval: paymentBackoffMax,
        RetryTimeout: markPaidBudget,
    }
    if err := ctx.CallActivity(MarkPaidActivity,
        workflow.WithActivityInput(MarkPaidInput{OrderID: orderID, PaymentID: outcome.paymentID}),
        workflow.WithActivityRetryPolicy(retry)).Await(nil); err != nil {
        return CheckoutResult{State: "FAILED_AFTER_PIVOT", OrderID: orderID, PaymentID: outcome.paymentID}, nil
    }
    return CheckoutResult{State: "COMPLETED", OrderID: orderID, PaymentID: outcome.paymentID}, nil
}
```

> `workflow.Any` / `Task.Any` exact API must be verified in Step 0.1 before implementation. The pseudocode shape is correct per `docs/saga.md:128-172`.

### 4.3 Activity contract

| Activity | Side effect | Idempotency key |
|----------|-------------|-----------------|
| `CreateOrderActivity` | Insert order PENDING + outbox `order.created` + idempotency-key upsert — all in same `ent.Tx` | `workflow_instance_id` |
| `PublishPaymentRequestedActivity` | `ent.Tx`: outbox row with `WithMessageID(workflowID+":pay-req:"+attempt)` + idempotency-key upsert | `workflowID+":pay-req:"+attempt`; outbox UNIQUE on `messageID` → `INSERT ON CONFLICT DO NOTHING` |
| `CancelOrderActivity` | Update order CANCELLED; treat `ErrOrderCannotCancel` (already CANCELLED) as success | `workflowID+":cancel"` |
| `MarkPaidActivity` | `repo.MarkPaid(order_id, payment_id)` — **returns existing row (not error) if already PAID with same `payment_id`; rejects only if PAID with a different `payment_id`** (see Step 3.5) | `(order_id, payment_id)` pair |

---

## 5. Acceptance Criteria (v0.3)

| # | Criterion | Evidence |
|---|-----------|---------|
| AC1 | `POST /v1/orders/checkout` with valid fields returns `200 {checkout_id, order_id}` and creates order row `status=PENDING` | Integration: HTTP call → DB assertion |
| AC2a | Same `idempotency_key`, in-flight, same process: returns identical `checkout_id`; one order row | Integration: sequential calls |
| AC2b | Same `idempotency_key` after **service restart**: returns identical `checkout_id` (cross-process durability) | Integration: call → restart service → call again |
| AC2c | Same `idempotency_key` after workflow **purged**: returns `CHECKOUT_DUPLICATE_KEY` (purged = response expired; client must use a new key) | Integration: complete → purge → retry same key |
| AC2d | Same `idempotency_key` + different `user_id`: returns `CHECKOUT_DUPLICATE_KEY` | Integration: two calls, different users |
| AC3 | `GET /v1/orders/checkout/{id}` returns `{state, order_id, payment_id, attempts, error}` from `wfc.FetchWorkflowMetadata` | Integration: poll freshly-scheduled workflow |
| AC4 | After `payment.completed` (with matching `workflow_instance_id`) arrives, order PENDING→PAID within 5s; payment row uniquely keyed by `(workflow_instance_id, attempt)` | Integration: `CompletePayment` → DB + status assertion |
| AC5 | After 3 `payment.failed` events (one per attempt), order PENDING→CANCELLED, workflow `state=FAILED, reason=payment_exhausted` | Integration: `FailPayment` ×3 |
| AC6 | No payment event within `perAttemptTimeout × maxAttempts + backoffs` (~3.1min): order CANCELLED, workflow FAILED. **Test bound: within 3.5min.** | Integration: no `CompletePayment`; assert within 3.5min |
| AC7 | Workflow body has zero forbidden non-deterministic calls. Verified by custom `golangci-lint` analyzer (not grep) covering: `time.Now/Since/Until/After/NewTicker/NewTimer`, `rand.*`, `os.Getenv`, `go func()`, raw map-range, channel ops outside activities. Applied to `app/order/internal/biz/` package. | Custom analyzer + CI gate |
| AC8 | All cross-service event publishes go through `pkg/outbox/` — no direct `client.PublishEvent` in `app/order/internal/biz/` or `app/order/internal/data/` | `grep -rn 'PublishEvent' app/order/internal/biz/ app/order/internal/data/` matches only outbox internals |
| AC9 | `workflowstore` Dapr component with `actorStateStore=true` present in `deploy/k8s/base/infra/dapr/` | YAML present + `kubectl get components -n go-mall workflowstore` |
| AC9b | `workflowstore` uses a **separate database** from ent business DB; backed by secret key `WORKFLOWSTORE_DATABASE_CONNECTION_STRING` | Helm chart diff + `kubectl get secret` |
| AC10 | *(Deferred to follow-up iteration — requires Step 4.6 tracing injection to be shipped first)* W3C `traceparent` propagates across the full saga path | — |
| AC11 | `make build` green for all services; `make test` passes `app/order/` and `app/payment/`; `make e2e-saga` happy-path passes | CI |
| AC12 | Late `payment.completed-{N}` for terminated workflow: inserted into `workflow_dead_letter_events` table, `saga_orphan_payment_total` incremented, no DB mutation, no panic. Subscriber distinguishes "workflow terminated" (`RaiseEvent` terminal error → DLQ) from transient error (inbox NACK → redelivery). | Integration: trigger cancel, publish late event |

---

## 6. Implementation Steps

### Step 0 — Research (before coding)

- **0.1** Confirm `workflow.Any` / `Task.Any` API name in `durabletask-go v0.11.3`. Fallback: use the two-`WaitForExternalEvent` + timer pattern from `docs/saga.md:161-172`.
- **0.2** Confirm `client.ScheduleWorkflow` error type for duplicate instance ID (typed error for `errors.Is` vs string-only).
- **0.3** Confirm `wfc.StopWorker(ctx)` or equivalent drain API in `go-sdk`. Document fallback if absent.
- **0.4** Confirm Dapr `state.postgresql v2` supports `retentionPeriod` metadata.

### Step 1 — Dependencies and infra

- **1.1** Promote `github.com/dapr/durabletask-go v0.11.3` from **indirect to direct** in `go.mod` (already present at `go.mod:42` — do NOT re-add).
- **1.2** Create `deploy/k8s/base/infra/dapr/workflowstore.yaml`:
  ```yaml
  apiVersion: dapr.io/v1alpha1
  kind: Component
  metadata:
    name: workflowstore
    namespace: go-mall
  spec:
    type: state.postgresql
    version: v2
    metadata:
      - name: connectionString
        secretKeyRef: { name: workflow-db-secret, key: WORKFLOWSTORE_DATABASE_CONNECTION_STRING }
      - name: actorStateStore
        value: "true"
      - name: retentionPeriod   # verify per Step 0.4; omit if unsupported
        value: "168h"
  ```
- **1.3** Add `workflowstore.yaml` to `deploy/k8s/base/infra/dapr/kustomization.yaml`.
- **1.4** Provision **separate** Postgres database for workflow state in `deploy/helm/` — add `workflow` DB to `initdbScripts` in `values.yaml`. Add `workflow-db-secret` Secret with key `WORKFLOWSTORE_DATABASE_CONNECTION_STRING` fetched from Dapr secret store (fallback: `DATABASE_CONNECTION_STRING`).
- **1.5** Update `app/order/cmd/server/main.go` to load `WORKFLOWSTORE_DATABASE_CONNECTION_STRING` via `secrets.Parse` (same pattern as `main.go:84`).
- **1.6** Add DLQ ent schema `workflow_dead_letter_events` to order service: fields `(id, topic, payload_json, workflow_instance_id, reason, created_at)`. Run `make ent` from `app/order/`.
- **1.7** Add `saga:` block to `app/order/internal/conf/conf.proto` and `configs/config.yaml`:
  ```yaml
  saga:
    enabled: false
    max_payment_attempts: 3
    per_attempt_timeout: "60s"
    payment_initial_delay: "500ms"
    payment_backoff_max: "30s"
    mark_paid_retry_max: 5
    mark_paid_budget: "5m"
    drain_timeout: "30s"
  ```
  Run `make config` from `app/order/`. Saga code reads from `conf.Saga`; no hard-coded constants.

### Step 2 — Proto contracts

- **2.1** Edit `api/order/v1/order.proto` — add to `Order` service:
  ```proto
  rpc Checkout(CheckoutRequest) returns (CheckoutResponse) {
    option (google.api.http) = { post: "/v1/orders/checkout" body: "*" };
  }
  rpc GetCheckoutStatus(GetCheckoutStatusRequest) returns (GetCheckoutStatusResponse) {
    option (google.api.http) = { get: "/v1/orders/checkout/{checkout_id}" };
  }
  ```
  Messages: `CheckoutRequest{idempotency_key, user_id, session_id, currency, items[]}`, `CheckoutResponse{checkout_id, order_id}`, `GetCheckoutStatusRequest{checkout_id}`, `GetCheckoutStatusResponse{state, order_id, payment_id, attempts, error}`. Reuse existing `OrderItem` message.
- **2.2** Edit `api/payment/v1/payment.proto` — add:
  ```proto
  rpc CompletePayment(CompletePaymentRequest) returns (Payment) {
    option (google.api.http) = { post: "/v1/payments/{id}/complete" body: "*" };
  }
  rpc FailPayment(FailPaymentRequest) returns (Payment) {
    option (google.api.http) = { post: "/v1/payments/{id}/fail" body: "*" };
  }
  ```
  Add error reasons: `PAYMENT_INVALID_TRANSITION`, `CHECKOUT_NOT_FOUND`, `CHECKOUT_DUPLICATE_KEY`.
- **2.3** Run `make api` from repo root; commit generated files.

### Step 3 — Order biz: workflow + activities

- **3.1** Create `app/order/internal/biz/saga.go` per §4.2. Duration constants in a `const` block (no `time.*` runtime calls). Import `fmt` only (no `rand`, `os`, raw maps).
- **3.2** `CreateOrderActivity` — wraps a new `OrderUsecase.CreateForCheckout(ctx, in, txCallbacks...)` that threads the idempotency-key upsert into the existing `CreateWithEvent` tx callback chain (so idempotency key + `OrderCreatedEvent` commit atomically). `OrderCreatedEvent` gains `WorkflowInstanceID string` field (additive JSON).
- **3.3** `PublishPaymentRequestedActivity` — opens `ent.Tx` explicitly; upserts idempotency key; calls `outbox.Publish(ctx, tx, "payment.requested", payload, outbox.WithMessageID(workflowID+":pay-req:"+attempt))`. Zero-rows-affected on conflict = success.
- **3.4** `CancelOrderActivity` — wraps `OrderUsecase.Cancel` (`app/order/internal/biz/order.go:124-136`); treats `ErrOrderCannotCancel` (already CANCELLED) as success.
- **3.5** **`MarkPaid` idempotency fix** — modify `app/order/internal/biz/order.go:138-150`: return existing order row (not error) if already `PAID` with the **same** `payment_id`; return new error `ErrPaymentConflict` if PAID with a *different* `payment_id`. This prevents `FAILED_AFTER_PIVOT` on any worker crash+replay after the DB commit.
- **3.6** `CheckoutUsecase` in `checkout.go`:
  - `Schedule`: call `wfc.ScheduleWorkflow` with `WithInstanceID(req.IdempotencyKey)`. Check error type (Step 0.2) — "already exists" → return existing checkout_id; validate `(idempotency_key, user_id)` match → reject with `CHECKOUT_DUPLICATE_KEY` on mismatch.
  - `Status`: `wfc.FetchWorkflowMetadata` → map Dapr runtime states + deserialise `CheckoutResult` output.
- **3.7** Update `app/order/internal/biz/biz.go` `ProviderSet` to include `NewCheckoutUsecase`, workflow registry provider, `wfc` client provider.

### Step 4 — Order data + event router

- **4.1** Add `app/order/internal/data/ent/schema/idempotencykey.go`: fields `(key PK, response_json, created_at)` with cleanup cron (TTL 7 days — matching workflowstore retention).
- **4.2** Verify outbox `messageID` column has UNIQUE constraint; add migration if absent. Update `insertOutboxSQL` → `INSERT ... ON CONFLICT (messageID) DO NOTHING`. Confirm `pkg/outbox/outbox.go:38-41` exposes `WithMessageID` option; add if absent.
- **4.3** Create `app/order/internal/server/subscriber.go`:
  - Subscribe to `payment.completed` and `payment.failed` on `pubsub` component.
  - Parse `attempt` from payload; compose event name (`"payment-completed-{N}"` / `"payment-failed-{N}"`).
  - Call `wfc.RaiseEvent(ctx, instanceID, eventName, payload)`.
  - On "workflow terminated/not found" error → insert into `workflow_dead_letter_events` + increment `saga_orphan_payment_total` + log with `workflow_instance_id`.
  - On transient error → inbox NACK → redelivery.
  - Idempotent via `pkg/outbox/inbox.go`.
- **4.4** Register subscriber as kratos `transport.Server` in `app/order/internal/server/server.go`.
- **4.5** Tracing — inject W3C `traceparent` into outbox `WithHeaders` inside activities using `go.opentelemetry.io/otel/propagation`. Required for AC10 (deferred but must be scaffolded now).
- **4.6** Register Prometheus counters/histograms: `saga_duration_seconds`, `saga_attempts_total{outcome}`, `saga_compensations_total`, `saga_orphan_payment_total`, `outbox_pending_gauge`. Add to a `metrics.go` file in `app/order/internal/biz/`.

### Step 5 — Order service layer

- **5.1** Implement `Checkout` and `GetCheckoutStatus` in `app/order/internal/service/order.go` (or new `checkout.go`). If `conf.Saga.Enabled == false`, `Checkout` returns `codes.Unimplemented` (fallback: use `CreateOrder` "manual mode"). Validate: `idempotency_key` non-empty UUID, `items` non-empty, amounts > 0.
- **5.2** Map `wfc` workflow states → `GetCheckoutStatusResponse.State` enum.

### Step 6 — Order server wiring

- **6.1** Wire `workflow.Registry` provider; register `OrderSaga` + four activities.
- **6.2** Wire `wfc, _ := client.NewWorkflowClient()` provider (retry per R6 pattern).
- **6.3** `WorkflowWorker` kratos transport adapter with **drain semantics**:
  ```go
  type WorkflowWorker struct {
      wfc    *client.WorkflowClient
      reg    *workflow.Registry
      cancel context.CancelFunc
  }
  func (w *WorkflowWorker) Start(ctx context.Context) error {
      c, cancel := context.WithCancel(context.Background())
      w.cancel = cancel
      return w.wfc.StartWorker(c, w.reg)
  }
  func (w *WorkflowWorker) Stop(ctx context.Context) error {
      // If StopWorker(drainCtx) API confirmed in Step 0.3, call it here with conf.Saga.DrainTimeout.
      w.cancel()
      return nil
  }
  ```
  **Shutdown ordering** in `app.New(...)`: `transport.Server(httpServer, grpcServer, workflowWorker, outboxServer)` — HTTP/gRPC stops first (no new requests accepted), then workflow worker drains, then outbox relay stops.
- **6.4** Run `make wire` from `app/order/`.
- **6.5** **PurgeWorkflow cron** — runs every 6h; queries terminal workflows older than 1h via a local `completed_workflows` tracking table; calls `wfc.PurgeWorkflow(ctx, instanceID)` for each. Wire into app lifecycle.

### Step 7 — Payment service

- **7.1** Add fields to `app/payment/internal/data/ent/schema/payment.go`: `workflow_instance_id string` (nullable, indexed) + `attempt int32` (default 0). Add UNIQUE constraint on `(workflow_instance_id, attempt)`.
- **7.2** Run `make ent` from `app/payment/`.
- **7.3** Add `app/payment/internal/server/subscriber.go`: subscribe to `payment.requested`; idempotent on `(workflow_instance_id, attempt)`; call `PaymentUsecase.CreateFromWorkflow(CreatePaymentInput{WorkflowInstanceID, Attempt, ...})` (request struct — no positional arg change).
- **7.4** Implement `CompletePayment` / `FailPayment` RPCs in `app/payment/internal/service/`:
  - State machine: PENDING → COMPLETED or PENDING → FAILED only. Reject all other transitions with `PAYMENT_INVALID_TRANSITION`.
  - On COMPLETED: outbox publish `payment.completed` with payload `{workflow_instance_id, attempt, payment_id, order_id}`.
  - On FAILED: outbox publish `payment.failed` with payload `{workflow_instance_id, attempt, order_id, reason_code}`.
  - Subscriber in order service reads `attempt` and routes to `payment-completed-{N}` / `payment-failed-{N}` event.
- **7.5** Add unit tests for state machine guards.

### Step 8 — Tests (see §7 for detail)

### Step 9 — Documentation and operations

- **9.1** Update `CLAUDE.md`: remove "Cross-service calls: none" under order/payment; replace with saga description.
- **9.2** Create `app/order/README.md`: workflow-of-life diagram + dev-mode curl guide (happy path via `localhost:8000` / `localhost:8001`).
- **9.3** Add `tilt-saga-demo` Tilt resource (`local_resource` in `Tiltfile`) for scripted happy-path curl sequence.
- **9.4** Create `app/order/docs/runbook-saga.md`:
  - **Orphan payment** (`saga_orphan_payment_total > 0`): query `workflow_dead_letter_events`; issue `FailPayment` or `RefundPayment` manually.
  - **Workflow stuck RUNNING**: `dapr workflow get -i {id}` / `dapr workflow terminate -i {id}`.
  - **State store unreachable**: flip `saga.enabled=false` via config rollout; services fall back to `CreateOrder` manual mode.
  - **Outbox backed up**: identify stuck rows in `outbox_messages`; `kubectl rollout restart deploy/order`.
- **9.5** Create `app/order/internal/biz/reconcile.go`: nightly cron joining `payments (COMPLETED)` ↔ `orders (PAID, payment_id)`. Report drift rows; emit `saga_reconciliation_drift_total`; alert on first occurrence.
- **9.6** Capacity model in runbook: `workflow_history_row_size × peak_checkouts_per_day × retention_days` with 3× headroom. Add Postgres disk-usage alerts at 60/75/85% before production deploy.
- **9.7** Deploy order (canary rollout): (a) payment service first (additive — new column nullable, RPCs idle); (b) order with `saga.enabled=false`; (c) verify build + tests; (d) flip `saga.enabled=true` for 5% canary; (e) monitor `saga_compensations_total`, `saga_orphan_payment_total` for 24h; (f) ramp 25% → 50% → 100%.

---

## 7. Verification Steps (Expanded Test Plan — v0.3)

### 7.1 Unit tests

| File | What it tests | ACs |
|------|---------------|-----|
| `app/order/internal/biz/saga_test.go` | Replay determinism via `durabletask-go` test harness; forbidden-call AST walker | AC7 |
| `app/order/internal/biz/activities_test.go` | Each activity with mocked repo + outbox; idempotency-key upsert in same `ent.Tx` | AC2a, AC4, AC5, AC8 |
| `app/order/internal/biz/order_markpaid_test.go` | `MarkPaid`: same `payment_id` → existing row (no error); different `payment_id` → `ErrPaymentConflict` | AC4, M7 |
| `app/order/internal/biz/checkout_test.go` | `Schedule`: duplicate key same user → existing; duplicate key different user → `CHECKOUT_DUPLICATE_KEY`; post-restart cross-process | AC2a-d |
| `app/order/internal/server/subscriber_test.go` | Inbox dedup; "workflow terminated" → DLQ; missing `workflow_instance_id` → reject + metric | AC8, AC12 |
| `app/payment/internal/service/payment_test.go` | State machine: PENDING→COMPLETED only; PENDING→FAILED only; reject all other transitions | AC4, AC5 |
| `app/payment/internal/server/subscriber_test.go` | `payment.requested` idempotent on `(workflow_instance_id, attempt)` | AC2a |

### 7.2 Integration tests (testcontainers-go)

| Scenario | Steps | Pass condition | ACs |
|----------|-------|----------------|-----|
| Happy path | Checkout → wait `payment.requested` → `CompletePayment` → poll status | Order PAID, workflow COMPLETED, one workflow run | AC1, AC4, AC11 |
| Idempotent checkout (same process) | Two sequential Checkout calls, same key | One order row, identical response | AC2a |
| Cross-process idempotency | Checkout → restart service → Checkout same key | Identical `checkout_id` returned | AC2b |
| Post-purge duplicate key | Checkout → complete → purge workflow → retry same key | `CHECKOUT_DUPLICATE_KEY` | AC2c |
| Different-user duplicate key | Checkout(key=K,user=A) → Checkout(key=K,user=B) | `CHECKOUT_DUPLICATE_KEY` on second | AC2d |
| Payment failure ×3 | Checkout → `FailPayment` ×3 | Order CANCELLED, workflow FAILED `payment_exhausted` | AC5 |
| Timeout exhaustion | Checkout → no payment events → wait | Order CANCELLED within 3.5min | AC6 |
| Late completion (AC12) | Cancel via timeout → publish late `payment.completed-1` | DLQ entry; no DB mutation | AC12 |
| Cross-attempt event leak | Checkout → fail attempt 1 → inject late `payment.completed-1` while attempt 2 in flight | Workflow accepts only `payment-completed-2`; no double-PAID | M2, R1 |
| MarkPaid idempotency | Workflow reaches pivot → crash worker → restart → replay MarkPaid | Order stays PAID (not FAILED_AFTER_PIVOT) | AC4, M7 |
| Outbox lag injection | Inject 90s delay in outbox relay | No spurious cancellation | R5 |
| Sidecar dies mid-RaiseEvent | Kill sidecar after `RaiseEvent` dispatched | Inbox NACK → redelivery on restart → workflow receives event | R7 |
| Two-replica dedup | Two order-service replicas receive competing events | Inbox dedup; workflow receives each event once | R7 |
| Status during RUNNING | Poll `GetCheckoutStatus` before payment events | `state=RUNNING`, `attempts=1`, `order_id` set | AC3 |

Runner: `make test-integration-saga`; expected runtime < 10min.

### 7.3 End-to-end (Tilt + curl)

| Scenario | Pass condition |
|----------|----------------|
| Happy path via real ingress | `GetCheckoutStatus` reaches `COMPLETED` with `order_id`+`payment_id` |
| Worker restart mid-saga | Workflow resumes; completes successfully |
| Sidecar restart mid-saga | Same |

Runner: `make e2e-saga`; nightly CI + local opt-in.

### 7.4 Observability tests

Require Step 4.5 (traceparent injection) and Step 4.6 (metric registration) to be shipped.

| Concern | Test | Tooling |
|---------|------|---------|
| Trace continuity (deferred — AC10) | Single `trace_id` HTTP→workflow→pub/sub→handler→MarkPaid | `e2e/saga/trace_test.go` with OTEL collector |
| Metrics emitted | Assert counters exist after run | Prometheus scrape assertion |
| Structured logs | Every saga log line carries `workflow_instance_id`, `order_id`, `attempt` | Captured log grep |

### 7.5 Static checks (CI gates)

- **AC7**: custom `golangci-lint` analyzer; forbidden pattern list in §5 AC7; applied to `app/order/internal/biz/` package.
- **AC8**: `grep -rn 'client\.PublishEvent' app/order app/payment` matches only `pkg/outbox/` internals.
- **AC9**: yq assertion `workflowstore.yaml` contains `actorStateStore: "true"`.
- **AC9b**: yq assertion `workflowstore.yaml` references `workflow-db-secret` (not `db-secret`).
- Outbox `messageID` UNIQUE constraint present in schema migration.
- `make build` green for all four services.

---

## 8. Risks and Mitigations (v0.3)

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|------------|--------|------------|
| R1 | Cross-attempt event routing — late attempt-N event satisfies attempt-N+1 wait | **HIGH** | **HIGH** | Attempt-keyed event names (M2 fix). DLQ for post-termination events (Step 1.6 + 4.3). Reconciliation cron (Step 9.5). Alert on `saga_orphan_payment_total > 0`. |
| R2 | Workflow body non-determinism breaks replay | Medium | High | Custom `golangci-lint` analyzer (AC7). Mandatory replay tests. |
| R3 | `actorStateStore=true` misconfigured → workflows wedge | Medium | High | yq CI assertion (AC9). `make tilt-doctor` target. |
| R4 | Idempotency-key collision / UUID reuse post-purge | Low | High | `(idempotency_key, user_id)` mismatch guard. Post-purge → `CHECKOUT_DUPLICATE_KEY` (use new key). |
| R5 | Outbox `BackoffMax=5min` exceeds `perAttemptTimeout=60s` → spurious cancellation | **HIGH** | Medium | Reduce outbox `BackoffMax` for saga-related topics; alert on `outbox_pending > 100` for > 5min. |
| R6 | Worker registers before Dapr sidecar reachable | Medium | Medium | Reuse existing 12×5s retry pattern (`main.go:74-99`). |
| R7 | Subscriber + outbox relay double-handling inbound events | Low | Medium | Inbox dedup (`pkg/outbox/inbox.go`). Two-replica integration test. |
| R8 | Workflow history grows unbounded | Low | **Medium** | PurgeWorkflow cron (Step 6.5) + Dapr `retentionPeriod` (Step 1.2) + separate DB (Step 1.4) + disk alerts (Step 9.6). |
| R9 | `Checkout` vs `CreateOrder` confusion | Low | Low | `CLAUDE.md` update (Step 9.1). Feature flag `saga.enabled` (Step 1.7). |
| R10 | First cross-service flow stresses untested infra | Medium | Medium | Feature flag (Step 1.7). Canary 5%→100% rollout (Step 9.7). |
| R11 | Workflowstore DSN shares rotation cycle with business DB | Low | High | Separate secret key (Step 1.4 + 1.5). Key-rotation procedure in runbook (Step 9.4). |
| R12 | Outbox retention (24h default) < workflow lifetime → `payment.requested` row swept before saga terminates | Low | Medium | Spec outbox retention ≥ max workflow age + grace (14d for saga-related topics); add per-topic config override. |
| R13 | No DLQ for `RaiseEvent` failures on terminated workflows — money-related events silently dropped | Medium | High | `workflow_dead_letter_events` table + DLQ logic in subscriber (Step 1.6 + 4.3). |
| R14 | Worker context-cancel propagation: kratos `Start(ctx)` passes process context; drain requires independent cancel | Low | Low | `WorkflowWorker` uses `context.WithCancel(context.Background())` for drain ctx, cancelled in `Stop(ctx)` (Step 6.3). |

---

## 9. Pre-mortem (v0.3)

### Scenario 1 — "Silent revenue leak"

**Headline:** Finance flags ~0.3% of orders CANCELLED with COMPLETED payments attached. $42k/month drift.

**How it happened:** Regional broker outage → outbox lag > 60s → per-attempt timeout fires before `payment.completed` from attempt 1 arrives → workflow retries → attempt 2 also completes → attempt-1 payment orphaned (COMPLETED, no matching PAID order). Dashboard tracked `saga_compensations_total` but `saga_orphan_payment_total` alert was never wired.

**v0.3 fixes:** Attempt-keyed event names prevent attempt-1 event from satisfying attempt-2 wait. DLQ (Step 1.6 + 4.3) captures orphan events. Reconciliation cron (Step 9.5) detects drift. `saga_orphan_payment_total > 0` is a hard alert. Residual risk (R5): outbox lag still causes spurious retries — see item 8 in §10 for v2 restructure.

### Scenario 2 — "Disk full at launch night"

**Headline:** Postgres hits 100% three weeks post-launch; all writes fail for 47 minutes.

**How it happened:** No Dapr retention policy. Workflow state on the same Postgres as ent business tables. ~30KB per saga × 200k checkouts/day = 6GB/day. No disk alert.

**v0.3 fixes:** Step 1.2 adds `retentionPeriod: 168h`. Step 1.4 provisions separate workflow database. Step 6.5 adds PurgeWorkflow cron. Step 9.6 adds capacity model + disk alerts before production deploy. R8 impact updated to Medium.

### Scenario 3 — "Day-1 checkout returns 409"

**Headline:** Mobile retry library reuses idempotency key on transient errors. Every retry returns `CHECKOUT_DUPLICATE_KEY`. Conversion −18% for 90 minutes.

**How it happened:** `Schedule` returned 409 (Stripe-style was supposed to return existing `checkout_id`). Cross-process case not tested. Staging single-replica masked the bug.

**v0.3 fixes:** AC2a-d now explicitly cover same-key-same-process, cross-process, post-purge, and different-user cases. `Schedule` implementation must pass all four. Integration tests for AC2b and AC2c added (§7.2). Pre-launch contract test with mobile/web teams + canary rollout (Step 9.7).

### Scenario 4 — "`FAILED_AFTER_PIVOT` on every worker restart" *(NEW)*

**Headline:** Post-launch monitoring: persistent `FAILED_AFTER_PIVOT` for a fraction of orders. Payments COMPLETED on the ledger; orders never flip to PAID.

**How it happened:** `MarkPaid` at `app/order/internal/biz/order.go:138-150` returns `ErrOrderAlreadyPaid` (HTTP 400) if the row is already PAID. Any worker crash between `MarkPaidActivity`'s DB commit and Dapr writing the history record causes replay. On replay, order is already PAID → `MarkPaid` returns error → activity fails → workflow retries until `markPaidBudget` exhausts → `FAILED_AFTER_PIVOT`.

**v0.3 fix:** Step 3.5 modifies `MarkPaid` to return the existing row (not error) if already PAID with the **same** `payment_id`. Only rejects if PAID with a different `payment_id`. `MarkPaidActivity` is idempotent under replay — no `FAILED_AFTER_PIVOT`. Integration test added (§7.2 "MarkPaid idempotency").

### Cross-cutting takeaways (v0.3 — all four folded in)

| Takeaway | Status |
|----------|--------|
| Step 9.4/9.5 (reconciliation cron + retention + capacity) | ✅ Steps 9.4, 9.5, 9.6 added |
| AC2 split for cross-process/post-restart | ✅ AC2a-d in §5; integration tests in §7.2 |
| R1 → HIGH/HIGH; R8 impact → Medium | ✅ §8 updated |
| Pre-launch checklist (alerts wired) | ✅ Alerts in §8 R1/R13; disk alerts Step 9.6; reconciliation cron Step 9.5 |

---

## 10. Open Questions / Future Work

1. **Auto-refund on late payment success** — revisit if Scenario 1 residual drift exceeds threshold in Q1 post-launch.
2. **Inventory reservation** — next saga step; adds `ReserveStock` activity before `PublishPaymentRequested`.
3. **Notification consumer** — `payment.completed` / `payment.failed` now published; notifications service can subscribe.
4. **Cart `Checkout` RPC** — cart-side endpoint that hydrates items and forwards to `order.Checkout`.
5. **Workflow versioning** — adopt `AddVersionedWorkflow` (Dapr v1.17+) before the second saga lands.
6. **Replay-against-staging tooling** — `dapr workflow ...` CLI for incident response.
7. **Pub/sub topic naming convention** — `payment.completed` (single topic, attempt in payload) confirmed as the standard; document in `deploy/k8s/base/infra/dapr/pubsub.yaml` subscription declarations.
8. **v2 restructure — Synthesis A** (publish-once + cancel round-trip): publish `payment.requested` once per workflow; on overall deadline → publish `payment.cancel.requested`; wait for `payment.cancelled` → cancel order. Eliminates the outbox-lag + spurious-retry window (R5). Deferred to v2.

---

## 11. ADR (v0.3)

**Decision:** Implement an orchestrated saga between order and payment services using Dapr Workflow, triggered by a new `Checkout` RPC, communicating via pub/sub with attempt-keyed `WaitForExternalEvent` for inbound results, and using a client-provided idempotency key as the workflow instance ID. Workflow state on a separate Postgres database with 7-day retention and PurgeWorkflow cron.

**Drivers:** correctness under partial failure (no double-charge); minimal change to existing services; operability.

**Alternatives considered:** direct gRPC (rejected: cross-import coupling), Dapr service invocation (rejected: sync model doesn't fit payment processor shape), pure choreography (rejected: no replay/durable timers), gRPC-control + pub/sub-notifications hybrid (rejected: compile-time cross-service dependency).

**Consequences:** introduces workflow state store dependency; adds three new RPCs, two subscribers, one DLQ table; attempt-keyed event names are a protocol contract between order and payment services; future sagas inherit this scaffolding.

**Follow-ups:** §10 items 1-8. Synthesis A restructure (item 8) targeted for v2 to close R5.

---

## 12. Plan changelog

- **v0.1 (2026-05-02)** — Initial draft after 7-question interview.
- **v0.2 (2026-05-02)** — Deliberate-mode augmentation: expanded test plan (unit/integration/e2e/observability) + pre-mortem with 3 failure scenarios.
- **v0.3 (2026-05-02)** — Consensus review iteration 1. Architect REVISE (4 blockers, 2 HIGH structural) + Critic REVISE (10 independent additional items). Changes:
  - Retry budget split into two named constant sets; AC6 test bound fixed to 3.5min.
  - Cross-attempt event routing race fixed via attempt-keyed event names (`payment-completed-{N}`).
  - DLQ table `workflow_dead_letter_events` (Step 1.6) + subscriber DLQ logic (Step 4.3).
  - Workflowstore: separate DB + secret `WORKFLOWSTORE_DATABASE_CONNECTION_STRING` + `retentionPeriod` + PurgeWorkflow cron (Step 6.5).
  - Drain-aware `WorkflowWorker` with explicit shutdown ordering (Step 6.3).
  - Pivot relabeled: `payment.completed` is the true pivot; `MarkPaid` is post-pivot retriable.
  - `MarkPaid` made idempotent on same `payment_id` (Step 3.5) — prevents `FAILED_AFTER_PIVOT` on replay.
  - Outbox idempotency: deterministic `WithMessageID` + UNIQUE constraint (Step 4.2 + Principle 2).
  - AC2 split into AC2a-d (in-flight, cross-process, post-purge, different-user).
  - AC7 upgraded to custom `golangci-lint` analyzer; AC10 deferred; AC11 enhanced with e2e gate.
  - `saga:` config block in `conf.proto` / `config.yaml` (Step 1.7) — no hard-coded constants.
  - All §9 cross-cutting takeaways folded into §5/§6/§8.
  - R1 → HIGH/HIGH; R5 → HIGH/Medium; R8 impact → Medium. Added R11-R14.
  - Added: Scenario 4 (MarkPaid replay bug), Step 0 research checklist, Steps 4.5-4.6, Steps 9.4-9.7, 8 missing integration test scenarios, `OrderCreatedEvent.WorkflowInstanceID`, deploy-order section (Step 9.7).
  - Option E added to alternatives. `durabletask-go` promote-not-add clarified (Step 1.1).
