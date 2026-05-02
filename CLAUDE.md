# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Root-level (run from repo root)

```bash
make init        # install protoc plugins, wire, ent CLI
make api         # generate Go/gRPC/HTTP code + OpenAPI from api/**/*.proto
make config      # generate Go structs from internal conf.proto files
make generate    # run go generate ./... + go mod tidy (wire + ent)
make build       # build all service binaries → ./bin/
make all         # api + config + generate
make dev         # tilt up --continue (dev env, Delve starts immediately)
make debug       # tilt up (Delve waits for debugger to attach on :2345)
```

### Per-app (run from app/<service>/, e.g. app/catalog/)

```bash
make wire        # regenerate wire_gen.go
make ent         # regenerate ent ORM from schema
make run         # go run ./cmd/server
make test        # go test -v ./...
```

### Run a single test
```bash
cd app/catalog && go test -v -run TestFunctionName ./internal/...
```

## Services

| Service | Domain | Storage | Events |
|---------|--------|---------|--------|
| `catalog` | Products & categories (CRUD, search, pagination) | ent + Postgres | none (pub/sub configured but unused) |
| `cart` | Session-based shopping carts; lazy creation, idempotent add | ent + Postgres | none |
| `order` | Order lifecycle (PENDING→PAID→SHIPPED→DELIVERED/CANCELLED); items stored as denormalized JSON | ent + Postgres | none |
| `payment` | Payment record ledger; status machine (PENDING→COMPLETED→REFUNDED/FAILED) | ent + Postgres | none |

**Cross-service calls:** The `app/web` Next.js frontend acts as a BFF orchestrator for the UCP checkout layer and makes cross-service HTTP calls: `GET http://localhost:8002/v1/carts/:sessionId` (cart) and `POST http://localhost:8004/v1/orders` (order). Go services themselves do not call each other. Product data is not validated on order creation; cart prices come from the caller and are not checked against catalog.

**catalog-only notes:**
- Read operations (List/Get products and categories) are public; write operations require JWT.
- `POST /v1/admin/seed` and `DELETE /v1/admin/clean` are dev-only HTTP endpoints (no auth guard—restrict at ingress).
- Stock count is informational; it is never decremented by orders.

**cart-only notes:**
- Carts are keyed by `session_id` (not user ID), enabling guest checkout.
- `AddItem` with an existing product increments quantity (unique index on `cart_id+product_id`). `UpdateItem` with quantity ≤ 0 deletes the item.
- Prices are stored at add time and never re-validated against catalog.

**order-only notes:**
- Order items are a JSON column—no normalized `order_items` table.
- Payment linking is post-hoc: create order first, then call `UpdateOrderStatus` with PAID status and a `payment_id` from outside this service.

**payment-only notes:**
- The service is a ledger, not a payment processor. The `provider` field (e.g. "stripe") is informational; no gateway SDK calls are made.
- Only COMPLETED payments can be refunded; PENDING/FAILED are rejected.

## Architecture

This is a **go-kratos** microservice monorepo. Services follow a common pattern; use an existing service (e.g. `catalog`) as reference when adding new ones.

### Layer stack (top → bottom)

```
api/catalog/v1/                ← protobuf contracts (source of truth)
  └─ generated: *.pb.go, *_grpc.pb.go, *_http.pb.go

app/catalog/
  cmd/server/                  ← entrypoint + Wire DI wiring
  internal/
    conf/                      ← config proto → generated conf.pb.go
    server/                    ← HTTP + gRPC server setup
    service/                   ← implements generated proto server interface
    biz/                       ← domain logic, repo interfaces (no infra deps)
    data/                      ← repo implementations, ent ORM client, Dapr client
      ent/schema/              ← ent schema definitions (source of truth for DB)
      ent/                     ← generated ORM code (do not edit by hand)
```

**Dependency rule:** `service → biz → data`. `biz` defines interfaces (`CatalogRepo`, `EventRepo`); `data` implements them. `biz` never imports `data`.

### Dependency injection

Google Wire is used. `wire.go` (build-tag-guarded) declares provider sets; `wire_gen.go` is the generated output. After any change to providers, run `make wire` from `app/<service>/`.

### Config & secrets

Config is loaded from `configs/config.yaml` via kratos `config/file`. **Secrets are injected at runtime by Dapr** (`secretstore` component). `main.go` retries the Dapr sidecar connection up to 12× (60s total) on startup, then overwrites `bc.Data.Database.Source` and `bc.Data.Redis.Addr` from the secret store.

### Dapr integration

The app depends on a Dapr sidecar (gRPC on `DAPR_GRPC_PORT`, default `50001`). Two Dapr components are declared in `deploy/k8s/base/infra/dapr/`:
- `secretstore` — secret injection on startup (only active use across all services)
- `pubsub` — declared but not yet wired in any service

Each service fetches its DB credentials from the secret store using a service-specific key with a shared fallback:
- `CATALOG_DATABASE_CONNECTION_STRING` → `DATABASE_CONNECTION_STRING`
- `CART_DATABASE_CONNECTION_STRING` → `DATABASE_CONNECTION_STRING`
- `ORDER_DATABASE_CONNECTION_STRING` → `DATABASE_CONNECTION_STRING`
- `PAYMENT_DATABASE_CONNECTION_STRING` → `DATABASE_CONNECTION_STRING`

`catalog` also fetches `KEYCLOAK_JWKS_URL` for JWT validation. All services retry the Dapr sidecar 12× at 5s intervals (60s total) and panic on failure—**Dapr sidecar must be running for any service to start**.

### Database

ent ORM with PostgreSQL (`lib/pq` driver). Schema lives in `internal/data/ent/schema/`. After editing schema, run `make ent`. `data.NewData` calls `client.Schema.Create` on startup (auto-migrate).

### Local dev environment (Tilt)

`tilt up` / `make dev` targets Docker Desktop or OrbStack (`allow_k8s_contexts`). The workflow:
1. `compile` local resource builds a Linux binary into `./dist/<service>` on every Go source change
2. Binary is synced into the running container — no image rebuild
3. Delve debugger runs inside the container; VS Code launch config at `.vscode/launch.json` connects to `:2345`
4. Helm chart at `deploy/helm/` provisions Postgres, Redis, pgAdmin in the `go-mall` namespace
5. Dapr is installed via Helm into `dapr-system` namespace

Port forwards: HTTP `8000`, gRPC `9000`, Delve `2345`, Postgres `5432`, Redis `6379`, pgAdmin `5050`.

## Testing

Use TDD: write tests first, confirm they fail for the right reason, then implement the minimal fix and re-run. Do not write maintenance-heavy tests (no exhaustive mocks, no tests that re-assert framework behavior, no tests that break on every refactor). Test behavior, not implementation.

Use `github.com/stretchr/testify/assert` for assertions. Structure every test with AAA comments:
```go
// Arrange
// Act
// Assert
```

### Proto conventions

- API protos: `api/<app>/<domain>/v<N>/<name>.proto` → `make api`
- Error reasons: defined in `error_reason.proto` as an enum; errors use `v1.ErrorReason_XXX.String()` as the reason field
- Internal config proto: `internal/conf/conf.proto` → `make config`
- `third_party/` holds vendored proto imports (google, validate)
