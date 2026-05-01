# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make wire        # regenerate wire_gen.go after provider changes
make ent         # regenerate ent ORM after schema changes
make run         # go run ./cmd/server
make test        # go test -v ./...
```

Run a single test:
```bash
go test -v -run TestFunctionName ./internal/...
```

Proto generation (run from repo root):
```bash
make api         # regenerates all proto → Go/gRPC/HTTP/OpenAPI
```

## Architecture

Layered clean architecture following the repo-wide Kratos pattern: `service → biz → data`.

### Layers

- **`internal/service/`** — implements the generated gRPC/HTTP server interface; validates inputs (UUID parsing, non-empty `user_id`/`session_id`, non-empty items list), converts between proto and biz models
- **`internal/biz/`** — domain logic; defines `OrderUsecase`, `OrderRepo` interface, domain model `Order` and `OrderItem`; enforces status machine and subtotal calculation; never imports `data`
- **`internal/data/`** — implements `OrderRepo` using ent ORM; items are marshalled to/from JSON on read/write

### Data model

Single ent entity: `Order`.

- Fields: `id` (UUID), `user_id` (string, indexed), `session_id` (string), `items` (JSON `[]OrderItem`), `total_cents int64`, `currency` (default `"USD"`), `status` (string, default `"PENDING"`, indexed), `payment_id` (string, optional), timestamps
- `OrderItem` is a Go struct serialised as a JSON column — there is no `order_items` table
- `SubtotalCents` on each item is computed as `price_cents × quantity` during `Create`; `total_cents` is the sum of all item subtotals

### Status machine

```
PENDING ──► PAID ──► SHIPPED ──► DELIVERED
   └──────────────────────────────────► CANCELLED (only from PENDING)
```

- `Create` sets status to `PENDING`
- `Cancel` only succeeds from `PENDING`; returns `ErrOrderCannotCancel` for PAID/SHIPPED/DELIVERED/CANCELLED
- `UpdateStatus` (via the `UpdateOrderStatus` RPC) transitions to any valid enum value without additional guards — use for PAID→SHIPPED→DELIVERED progression
- `MarkPaid(id, paymentID)` is an internal biz method (not exposed as an RPC) that sets `payment_id` then transitions to `PAID`; rejects if already PAID or CANCELLED

### Payment linking

Payment linking is post-hoc. The public `UpdateOrderStatus` RPC does **not** carry a `payment_id` field — it only transitions the status. To record a payment ID, call the internal `MarkPaid` usecase method directly (e.g., from a future event handler or inter-service call), or extend the proto with an optional `payment_id` field.

### API surface

Defined in `api/order/v1/order.proto`.

| Method | Path | RPC |
|--------|------|-----|
| POST | `/v1/orders` | CreateOrder |
| GET | `/v1/orders/{id}` | GetOrder |
| GET | `/v1/orders` | ListOrders |
| POST | `/v1/orders/{id}/status` | UpdateOrderStatus |
| POST | `/v1/orders/{id}/cancel` | CancelOrder |

`ListOrders` accepts `user_id`, `status` (enum, UNSPECIFIED = no filter), `page`, `page_size` (default 20).

Error reasons: `ORDER_NOT_FOUND`, `INVALID_ARGUMENT`, `ORDER_CANNOT_CANCEL`, `ORDER_ALREADY_PAID`, `ORDER_EMPTY_ITEMS`.

### Dependency injection

Wire wires: `NewData → NewOrderRepo → NewOrderUsecase → NewOrderService → servers → kratos.App`. Run `make wire` after any provider change.

### Config & secrets

`configs/config.yaml` sets server addresses and timeouts. The DB connection string is injected at runtime via Dapr secret store (key `ORDER_DATABASE_CONNECTION_STRING` → `DATABASE_CONNECTION_STRING` fallback). Dapr sidecar must be reachable; `main.go` retries 12× at 5 s intervals then panics.

### Testing approach

- `internal/biz/order_test.go` — unit tests with a stub repo; cover subtotal/total calculation, status transitions, `MarkPaid` idempotency, `UpdateStatus` enum validation, list filtering
- `internal/service/order_test.go` — validation tests with the same stub repo; verify UUID parsing, empty-items rejection, proto↔biz conversion, status mapping
