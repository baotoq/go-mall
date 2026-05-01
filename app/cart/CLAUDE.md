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

- **`internal/service/`** — implements the generated gRPC/HTTP server interface; handles proto serialization and input validation (UUID format, quantity bounds)
- **`internal/biz/`** — domain logic; defines `CartUsecase`, `CartRepo` interface, and domain models (`Cart`, `CartItem`); never imports `data`
- **`internal/data/`** — implements `CartRepo` using ent ORM; `FindOrCreateBySession` is the entry point for all cart access

### Data model

Two ent entities: `Cart` (keyed by `session_id`, unique) and `CartItem` (composite unique index on `cart_id + product_id`).

- Prices stored as `price_cents int64`; cart total is computed on the fly from items
- `AddItem` upserts: existing product increments quantity, new product inserts
- `UpdateItem` with quantity ≤ 0 deletes the item
- `image_url` is nullable; `currency` defaults to `"USD"`

### API surface

Defined in `api/cart/v1/cart.proto`. HTTP routes:

| Method | Path | RPC |
|--------|------|-----|
| GET | `/v1/carts/{session_id}` | GetCart |
| POST | `/v1/carts/{session_id}/items` | AddItem |
| PUT | `/v1/carts/{session_id}/items/{product_id}` | UpdateItem |
| DELETE | `/v1/carts/{session_id}/items/{product_id}` | RemoveItem |
| DELETE | `/v1/carts/{session_id}` | ClearCart |

Error reasons are defined in `api/cart/v1/error_reason.proto` (`CART_NOT_FOUND`, `ITEM_NOT_FOUND`, `INVALID_ARGUMENT`).

### Dependency injection

Wire wires: `NewData → NewCartRepo → NewCartUsecase → NewCartService → servers → kratos.App`. After any provider change run `make wire`.

### Config & secrets

`configs/config.yaml` sets server addresses and timeouts. The DB connection string is injected at runtime via Dapr secret store (key `CART_DATABASE_CONNECTION_STRING` falling back to `DATABASE_CONNECTION_STRING`). The Dapr sidecar must be reachable on startup; `main.go` retries 12× at 5 s intervals then panics.

### Testing approach

- `internal/biz/cart_test.go` — unit tests with a stub `CartRepo`; covers all use-case paths and total calculation
- `internal/service/cart_test.go` — validation tests with a nop repo; verifies UUID format checks, empty session, and quantity bounds
