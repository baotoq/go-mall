# Order Service

The order service manages the full order lifecycle (PENDING → PAID → SHIPPED → DELIVERED / CANCELLED) and provides a durable saga-based checkout flow powered by Dapr Workflow. Clients can either create orders directly via `CreateOrder` (manual mode) or use the `Checkout` RPC to start an orchestrated saga that coordinates order creation and payment authorization as a single durable transaction — with automatic retries, compensation, and idempotency across process restarts.

## Workflow diagram

```
              ┌──────────────────────────────────────────────────────┐
              │                  order service                       │
              │  ┌──────────┐    ┌──────────────────────────────┐    │
client──Checkout─▶ service │───▶│ OrderSaga workflow           │────┼──┐
              │  └──────────┘    │  CreateOrder                 │    │  │ pub: payment.requested
              │       ▲          │  loop(1..3):                 │    │  │ {workflow_id, order_id,
              │       │ poll     │    PublishPaymentRequested    │    │  │  amount, currency, attempt}
              │  ┌────┴─────┐    │    WaitForExternalEvent      │    │  ▼
              │  │GetCheckout    │    "payment-result-{N}"      │    │ ┌────────────────────┐
              │  │  Status  │    │    (timeout via SDK arg →    │    │ │   pubsub (Redis)   │
              │  └──────────┘    │     task.ErrTaskCanceled)    │    │ └────────────────────┘
              │                  │  MarkPaid (post-pivot retry)  │    │  ▲                │
              │                  │  CancelOrder (compensation)   │    │  │                │ sub:
              │  ┌─────────────────────────────────────────────┐│    │  │ payment.        │ payment.
              │  │ event router (subscriber.go)                ││    │  │ completed/failed│ requested
              │  │ sub: payment.completed, payment.failed       ││    │  │ {workflow_id,   │ {attempt=N}
              │  │ map → "payment-result-{N}" w/ payload         ││    │  │  pay_id, reason,│
              │  │      {Success, PaymentID, ReasonCode}        ││    │  │  attempt=N}     │
              │  │ → RaiseEvent(wfID, "payment-result-{N}", ..) │────┘  │                 │
              │  └─────────────────────────────────────────────┘│      │                 │
              └──────────────────────────────────────────────────┘      │                 │
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

## Service responsibilities

- Order CRUD (existing): `CreateOrder`, `GetOrder`, `ListOrders`, `UpdateOrderStatus`, `CancelOrder`.
- Checkout via Dapr Workflow saga (new): `Checkout` starts a durable workflow; `GetCheckoutStatus` polls it.
- DLQ for orphan payment events: late `payment.completed` / `payment.failed` messages for terminated workflows are written to `workflow_dead_letter_events` instead of being silently dropped.
- PurgeWorkflow cron: runs every 6h to prune terminal workflow history older than 1h from the workflowstore, keeping Postgres disk bounded.

## Configuration

The `saga:` block in `configs/config.yaml` controls the saga subsystem:

```yaml
saga:
  enabled: false
  max_payment_attempts: 3
  per_attempt_timeout: 60s
  payment_initial_delay: 0.5s
  payment_backoff_max: 30s
  mark_paid_retry_max: 5
  mark_paid_budget: 5m
  drain_timeout: 30s
```

When `saga.enabled=false`, the `Checkout` RPC returns `501 Unimplemented` and callers fall back to the manual `CreateOrder` mode. All other saga config fields are read at startup; there are no hard-coded constants in saga code.

## Dev mode quickstart

Assumes Tilt is running (`make dev`) with the order service on port `8000` and payment service on port `8001`. Requires `saga.enabled=true` in `configs/config.yaml` and a Dapr sidecar with `workflowstore` configured.

```bash
# 1. Start a checkout (saga.enabled=true required)
curl -X POST http://localhost:8000/v1/orders/checkout \
  -H 'Content-Type: application/json' \
  -d '{
    "idempotency_key": "11111111-2222-3333-4444-555555555555",
    "user_id": "user-demo-1",
    "session_id": "sess-demo-1",
    "currency": "USD",
    "items": [
      {"product_id": "prod-1", "quantity": 2, "price_cents": 1500}
    ]
  }'
# → {"checkout_id":"...","order_id":"..."}
```

```bash
# 2. Poll status
curl http://localhost:8000/v1/orders/checkout/<checkout_id>
# → {"state":"RUNNING|COMPLETED|FAILED|FAILED_AFTER_PIVOT", "order_id":"...", "payment_id":"...", "attempts":0, "error":""}
```

```bash
# 3. Simulate happy-path payment completion (drives the saga forward)
curl -X POST http://localhost:8001/v1/payments/<payment_id>/complete \
  -H 'Content-Type: application/json' \
  -d '{}'
```

```bash
# 4. Simulate failure (compensates the order)
curl -X POST http://localhost:8001/v1/payments/<payment_id>/fail \
  -H 'Content-Type: application/json' \
  -d '{"reason_code": "insufficient_funds"}'
```

Note: `payment_id` is observable by querying `GET /v1/payments?order_id=<order_id>` on the payment service, or by reading workflow status after the `payment.requested` event lands and the payment service inserts the row.

## Architecture notes

Order saga is implemented as a durable workflow via Dapr Workflow runtime backed by a separate Postgres database (see `deploy/k8s/base/infra/dapr/workflowstore.yaml`). Payment events flow through the `pubsub` Dapr component using the transactional outbox pattern (see `pkg/outbox/`). The subscriber in `internal/server/subscriber.go` merges `payment.completed` and `payment.failed` topics into a single attempt-keyed `payment-result-{N}` workflow event, so the workflow body needs only one `WaitForExternalEvent` call per attempt. The true saga pivot is `payment.completed` landing in the payment service's database — once that commits, `MarkPaid` is forward-only and retried up to 5 times rather than compensated.

## Links

- Plan: `.omc/plans/order-saga-dapr-workflow.md`
- Runbook: `app/order/docs/runbook-saga.md`
- ADR: see §11 of the plan doc
