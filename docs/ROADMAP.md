# go-mall Ecommerce Feature Roadmap

Repo-grounded roadmap. Each item lists the concrete service(s), files, and infrastructure the work touches. Phases are ordered by dependency and risk: correctness gaps first, then domain expansion, then growth features, then platform maturity.

## Current state (2026-05-02)

| Area | Status |
|------|--------|
| Services | `catalog`, `cart`, `order`, `payment`, `web` (Next.js) |
| Shared packages | `pkg/id` (UUIDv7), `pkg/money`, `pkg/bootstrap`, `pkg/outbox` (txn outbox + inbox), `pkg/secrets`, `pkg/server` (cors, health) |
| Auth (JWT via Keycloak) | Wired: `catalog` (writes), `payment`. Missing: `cart`, `order` |
| Cross-service calls | None (gRPC/HTTP) |
| Events | `order` publishes `order.created` via outbox; **no consumers wired** |
| Dapr | `secretstore` active; `pubsub` component declared but not consumed |
| Persistence | Postgres + ent ORM per service; auto-migrate on startup |
| Local dev | Tilt + Helm + Delve attach |
| Stock | Informational only — never decremented |
| Cart prices | Trusted from caller, not validated against catalog |
| Payment | Ledger only — no gateway SDK |
| Order items | Denormalized JSON column |

---

## Phase 0 — Correctness Foundations

Block: cannot run real ecommerce flows safely without these.

### 0.1 — Auth on `cart` and `order`
- **Why:** anyone can read/mutate carts and orders by guessing IDs; payment already requires JWT.
- **Touches:** `app/cart/internal/server/{grpc,http}.go`, `app/order/internal/server/{grpc,http}.go`, `app/<svc>/internal/conf/conf.proto` (add JWKS URL like `catalog`), Wire providers, K8s `secretstore` keys (`CART_KEYCLOAK_JWKS_URL`, `ORDER_KEYCLOAK_JWKS_URL` with shared fallback).
- **Pattern reference:** copy from `app/payment/internal/server/*.go`.
- **Acceptance:** all write RPCs reject missing/invalid JWT; read RPCs allow service-token or owner-only access.

### 0.2 — Order → Catalog price validation
- **Why:** `cart` accepts caller-supplied prices; `order` snapshots them. A malicious client can buy at any price.
- **Touches:** new `CatalogClient` interface in `app/order/internal/biz/`, gRPC client wired in `app/order/internal/data/`, `app/order/cmd/server/wire.go`. Reuse `pkg/bootstrap` for client setup.
- **Approach:** on `CreateOrder`, fetch each `Product` by ID from `catalog`, fail order if `price != snapshot.price` (or if product missing/inactive). Use a short-circuit budget (e.g. 500ms total) and surface a domain error reason.
- **Acceptance:** order creation rejects mismatched prices with a typed `ErrorReason_PRICE_MISMATCH`.

### 0.3 — Stock reservation via `order.created`
- **Why:** stock count never moves. Two orders for the last unit both succeed.
- **Touches:** `app/catalog` consumer using `pkg/outbox.Inbox`, new schema `stock_reservation` (ent), new biz method `ReserveStock(ctx, orderID, items)`, idempotent on `(order_id, product_id)`. Emit `stock.reserved` and `stock.rejected` events.
- **Decision:** reservation lives in `catalog` for now; split into a dedicated `inventory` service in Phase 1.6 if write contention warrants it.
- **Acceptance:** consuming `order.created` decrements `stock_count` atomically; insufficient stock publishes `stock.rejected` and the order transitions to CANCELLED.

### 0.4 — Order saga: react to stock + payment events
- **Why:** today `UpdateOrderStatus` is called manually from outside. Real flow must react to `stock.reserved`, `payment.completed`, `payment.failed`.
- **Touches:** `app/order/internal/biz` consumers via `pkg/outbox.Inbox`, new topics `payment.completed` / `payment.failed` from `payment` service, new state machine entry points. Reference `docs/saga.md`.
- **Acceptance:** order auto-transitions PENDING → CONFIRMED (stock.reserved + payment.completed) or CANCELLED (either rejected). Idempotent.

### 0.5 — Wire Dapr pubsub end-to-end
- **Why:** `pubsub.yaml` exists; no service consumes from it. The `outbox` relay pushes to a sink that nobody reads.
- **Touches:** `app/<svc>/internal/data/eventbus.go` (new), `pkg/outbox/relay.go` config, Helm values for broker (Redis Streams or Kafka). Verify CloudEvent envelope alignment with `pkg/outbox`.
- **Acceptance:** integration test publishes via outbox in service A, consumer in service B receives via inbox, dedupe works on retry.

---

## Phase 1 — Domain Expansion

Add the services that real ecommerce needs but aren't represented today.

### 1.1 — `identity` service (users, addresses)
- Currently `cart.session_id` is the only identity anchor; `order.user_id` is a free string.
- Schema: `user`, `address`, `customer_profile`. Hooks to Keycloak (Keycloak owns auth, this owns commerce profile).
- Emits: `user.created`, `address.changed`.

### 1.2 — `payment` gateway adapter
- Today payment is a ledger. Add a pluggable provider interface.
- Touches: `app/payment/internal/biz/provider.go` (interface), `provider_stripe.go` (impl), webhook receiver in `app/payment/internal/server/http.go`.
- Idempotency keys per `payment.id`. Emits `payment.completed` / `payment.failed` (consumed by Phase 0.4).

### 1.3 — `shipping` service
- Schema: `shipment`, `tracking_event`. Consumes `order.confirmed`, calls carrier API stub.
- Emits: `shipment.created`, `shipment.delivered` (drives `order.UpdateOrderStatus → DELIVERED`).

### 1.4 — `tax` service (or `pricing` engine)
- Compute tax + shipping at order time. Even a flat-rate stub now creates the seam for real tax APIs (Avalara/TaxJar) later.
- Order calls `pricing.Quote` synchronously during `CreateOrder`, after Phase 0.2 price validation.

### 1.5 — `notification` service
- Consumes: `order.confirmed`, `order.shipped`, `payment.failed`, `user.created`.
- Channels: email (SES/SMTP), in-app (web), webhook. Templated per locale.

### 1.6 — `inventory` service split (conditional)
- Trigger: stock reservation contention or warehouse expansion.
- Take ownership of `stock_count`, `stock_reservation`, multi-warehouse, replenishment thresholds. `catalog` keeps merchandising data only.

---

## Phase 2 — Commerce Growth Features

### 2.1 — Search
- `meilisearch` or `opensearch`. Indexer consumes `product.created` / `product.updated` from `catalog` (need to add these events to `catalog`).
- Endpoints: typeahead, faceted search, filters (price, category, attribute).

### 2.2 — Promotions / coupons
- New `promotion` service. Schema: `coupon`, `promotion_rule`, `redemption`.
- Cart calls `promotion.Apply(cart)`; result is a `discount_line` rendered server-side.
- Idempotency: redemption row per `(coupon, order_id)`.

### 2.3 — Reviews & ratings
- `review` service. Eligibility check: only users with a DELIVERED order for the SKU may review.
- Pre-publication moderation hook (manual or LLM); aggregate ratings rolled up per product.

### 2.4 — Wishlist / saved-for-later
- Lives near `cart` or as its own small service. Survives session boundary (requires Phase 1.1 user identity).

### 2.5 — Recommendations
- Read-model service. Consumes `order.created`, `product.viewed` (need viewer tracking from `web`).
- Phase A: simple co-purchase. Phase B: vector retrieval if useful.

### 2.6 — Returns / RMA
- New `returns` service. Schema: `return_request`, `return_item`, `refund_link`.
- State machine: REQUESTED → APPROVED → RECEIVED → REFUNDED.
- Drives `payment.RefundPayment` once received.

---

## Phase 3 — Operational Maturity

### 3.1 — Observability stack
- OpenTelemetry tracing: gRPC + HTTP middleware in each service. Helm-installed Tempo/Jaeger.
- Metrics: kratos `metrics.Server` middleware → Prometheus. Outbox lag, inbox dedupe rate, saga-stuck counters.
- Log correlation: trace_id in every log line. Pino-style structured logs.

### 3.2 — API gateway / BFF
- One ingress fronting public REST. Could be kratos-bff or Envoy + Go gateway.
- Owns: rate limiting, request shaping, schema versioning, public auth.

### 3.3 — Admin app
- Routes in `app/web` for: catalog mgmt, order ops, refund button, customer search, coupon issuance, inventory adjustments.
- Reuses existing JWT; admin role from Keycloak.

### 3.4 — Reporting & analytics read-model
- Dedicated read-only service consuming all domain events into a denormalized warehouse table set (Postgres or ClickHouse).
- Powers admin dashboards: GMV, conversion, AOV, refund rate, top products.

### 3.5 — Audit log
- Cross-cutting concern. Either dedicated `audit` service consuming all events, or per-service audit table behind a shared interface in `pkg/audit`.

---

## Phase 4 — Platform & Scale

### 4.1 — Multi-currency / multi-locale
- `pkg/money` already exists — extend with FX rate snapshotting per order. Locale on user profile.

### 4.2 — Feature flags
- LaunchDarkly client or open-source (Unleash, GrowthBook). Used for staged rollouts of any phase ≥ 2.

### 4.3 — Tenant isolation (if multi-merchant)
- Add `tenant_id` to all schemas behind a feature flag; enforce at biz boundary, not just data layer.

### 4.4 — Disaster recovery
- Postgres PITR per service. Outbox replay tooling. Saga rerun command.

### 4.5 — Performance pass
- pprof endpoints behind admin auth.
- Catalog: cache product reads in Redis with explicit invalidation on `product.updated`.
- Order list queries: add covering indexes, paginate with keyset.

---

## Cross-cutting principles

1. **Every new event uses `pkg/outbox`** — no direct broker writes from biz layer. Inbox dedupe required on every consumer.
2. **No biz-layer cross-service calls without a circuit breaker** — wrap all gRPC clients with kratos middleware (timeout, retry, breaker).
3. **All write paths require JWT** by default. Public reads stay anonymous unless they leak inventory or pricing intelligence.
4. **Schema migrations stay in `internal/data/ent/schema/`**. Auto-migrate is acceptable in dev; production needs explicit migration step before Phase 3.1.
5. **Saga compensation is non-optional.** Every state machine transition that touches another service must define and test its compensating action.

---

## Suggested execution order

```
P0.1 (auth) → P0.5 (pubsub) → P0.2 (price validation)
                ↓
       P0.3 (stock) → P0.4 (saga)            ← unblocks real checkout
                ↓
P1.1 (identity) → P1.2 (payment gateway) → P1.4 (tax) → P1.3 (shipping) → P1.5 (notifications)
                ↓
P2.1 (search) ‖ P2.2 (promotions) ‖ P2.3 (reviews)   ← parallelizable
                ↓
P2.4–P2.6, P3.x, P4.x   ← prioritize by business signal
```

`P0` is sequential; from `P1` onward, services are independent enough to parallelize across teams. Phase 3 observability should land before Phase 2 ships to production so growth features have measurement.
