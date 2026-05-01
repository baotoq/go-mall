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

- **`internal/service/`** — implements the generated gRPC/HTTP server interface; handles proto serialization, input validation (UUID, page size cap at 100, sort whitelist), and JWT-guarded route enforcement
- **`internal/biz/`** — domain logic; defines `ProductUsecase` and `CategoryUsecase`, their repo interfaces, and domain models; never imports `data`
- **`internal/data/`** — implements repo interfaces using ent ORM; `data_test.go` spins up a real PostgreSQL via testcontainers as `TestMain`

### Data model

Two ent entities: `Product` and `Category` (one-to-many).

- `Product`: `id` (UUID), `name`, `slug` (unique), `description` (nullable), `price_cents int64`, `currency` (default `"USD"`), `image_url` (nullable), `theme` (enum `light`/`dark`), `stock int`, `category_id` (nullable FK), timestamps
- `Category`: `id` (UUID), `name`, `slug` (unique), `description` (nullable), timestamps
- Index on `(category_id, created_at)` for product listing

Slug uniqueness violations surface as `ErrSlugConflict` (409 Conflict) from the data layer.

### API surface

Defined in `api/catalog/v1/catalog.proto`. Public routes (no JWT):

| Method | Path | RPC |
|--------|------|-----|
| GET | `/v1/products` | ListProducts |
| GET | `/v1/products/{id}` | GetProduct |
| GET | `/v1/categories` | ListCategories |
| GET | `/v1/categories/{id}` | GetCategory |

Protected routes (RS256 JWT required):

| Method | Path | RPC |
|--------|------|-----|
| POST | `/v1/products` | CreateProduct |
| PUT | `/v1/products/{id}` | UpdateProduct |
| DELETE | `/v1/products/{id}` | DeleteProduct |
| POST | `/v1/categories` | CreateCategory |
| PUT | `/v1/categories/{id}` | UpdateCategory |
| DELETE | `/v1/categories/{id}` | DeleteCategory |

Admin-only HTTP handlers registered manually in `server/http.go` (no auth guard — restrict at ingress):
- `POST /v1/admin/seed` — seed demo data
- `DELETE /v1/admin/clean` — delete all data

Error reasons: `PRODUCT_NOT_FOUND`, `CATEGORY_NOT_FOUND`, `INVALID_ARGUMENT`, `CONFLICT`.

### Auth/JWT

JWT middleware uses RS256 via `github.com/golang-jwt/jwt/v5` + `github.com/MicahParks/keyfunc/v3`. JWKS URL is injected at runtime from the Dapr secret store (`KEYCLOAK_JWKS_URL`). Public routes are excluded via a selector that checks the operation against a `publicRoutes` map in both HTTP and gRPC servers.

### ListProducts filters

- Full-text search via `Q`
- Category filter
- Price range (`MinPrice`, `MaxPrice` in cents)
- Sort: `price_asc`, `price_desc`, `created_desc` (default)
- Pagination: `page` + `pageSize` (1–100)

### Dependency injection

Wire wires: `NewData → NewProductRepo + NewCategoryRepo → usecases → NewCatalogService → servers → kratos.App`. Run `make wire` after any provider change.

### Config & secrets

`configs/config.yaml` sets addresses and timeouts. The DB connection string and JWKS URL are injected at runtime via Dapr secret store (keys `CATALOG_DATABASE_CONNECTION_STRING` → `DATABASE_CONNECTION_STRING` fallback, and `KEYCLOAK_JWKS_URL`). Dapr sidecar must be reachable; `main.go` retries 12× at 5 s intervals then panics.

### Testing approach

- `internal/biz/*_test.go` — unit tests with stub repos (in-memory maps); cover all usecase paths including slug conflict and not-found errors
- `internal/service/service_test.go` — validation tests with nop repos; verify UUID parsing, page size clamping, sort validation, required-field checks
- `internal/data/data_test.go` — integration tests with real PostgreSQL via testcontainers (`TestMain` setup); truncate helper between tests
