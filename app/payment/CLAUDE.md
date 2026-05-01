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

- **`internal/service/`** — implements the generated gRPC/HTTP server interface; validates inputs (non-empty IDs, UUID parsing, `amount_cents > 0`, required `provider`), converts between proto and biz models
- **`internal/biz/`** — domain logic; defines `PaymentUsecase`, `PaymentRepo` interface, and domain model; enforces the refund state guard; never imports `data`
- **`internal/data/`** — implements `PaymentRepo` using ent ORM

### Data model

Single ent entity: `Payment`.

- Fields: `id` (UUID), `order_id` (string, indexed), `user_id` (string, indexed), `amount_cents int64`, `currency` (default `"USD"`), `status` (string max 20, default `"PENDING"`), `provider` (string), timestamps
- Status is stored as a plain string; the proto enum is mapped in `bizToPayment()`; unknown values fall back to `PAYMENT_STATUS_UNSPECIFIED`

### Status machine

```
PENDING ──► COMPLETED ──► REFUNDED
   └──────────────────────► FAILED
```

- `Create` always sets status to `PENDING`
- Only `COMPLETED` payments can be refunded; any other status returns `ErrPaymentCannotRefund`
- `COMPLETED` and `FAILED` are terminal states set externally (no RPC exposes a direct transition to them)

### API surface

Defined in `api/payment/v1/payment.proto`.

| Method | Path | RPC |
|--------|------|-----|
| POST | `/v1/payments` | CreatePayment |
| GET | `/v1/payments/{id}` | GetPayment |
| GET | `/v1/payments` | ListPayments |
| POST | `/v1/payments/{id}/refund` | RefundPayment |

Error reasons: `PAYMENT_NOT_FOUND`, `INVALID_ARGUMENT`, `PAYMENT_ALREADY_COMPLETED`, `PAYMENT_CANNOT_REFUND`.

### ListPayments filtering

`ListPayments` accepts `user_id`, `order_id`, `page`, `page_size`. In the biz layer: if `order_id` is provided it takes priority (calls `ListByOrder`, ignores pagination); otherwise it queries `ListByUser` with pagination (page < 1 defaults to 1, pageSize < 1 defaults to 20).

### Dependency injection

Wire wires: `NewData → NewPaymentRepo → NewPaymentUsecase → NewPaymentService → servers → kratos.App`. Run `make wire` after any provider change.

### Config & secrets

`configs/config.yaml` sets server addresses and timeouts. The DB connection string is injected at runtime via Dapr secret store (key `PAYMENT_DATABASE_CONNECTION_STRING` → `DATABASE_CONNECTION_STRING` fallback). Dapr sidecar must be reachable; `main.go` retries 12× at 5 s intervals then panics.

### Testing approach

- `internal/biz/payment_test.go` — unit tests with a stub repo; cover `Create` (status forced to PENDING), `Refund` (state guard), `ListPayments` (filter priority)
- `internal/service/payment_test.go` — validation tests with a nop repo; verify UUID parsing, required-field checks, unknown status mapping to UNSPECIFIED
