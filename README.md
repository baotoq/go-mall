# go-mall

A microservice e-commerce monorepo built with [go-kratos](https://github.com/go-kratos/kratos), [Dapr](https://dapr.io/), and a [Next.js](https://nextjs.org/) BFF frontend.

## Services

| Service | Domain | Port (HTTP) | Port (gRPC) |
|---------|--------|-------------|-------------|
| `catalog` | Products & categories (CRUD, search, pagination) | 8000 | 9000 |
| `cart` | Session-based shopping carts | 8001 | 9001 |
| `order` | Order lifecycle (PENDING → PAID → SHIPPED → DELIVERED/CANCELLED) | 8002 | 9002 |
| `payment` | Payment record ledger | 8004 | 9004 |
| `web` | Next.js BFF + storefront | 3000 | — |

## Tech Stack

- **Go services:** go-kratos, ent ORM, PostgreSQL, Google Wire, Dapr sidecar
- **Frontend:** Next.js (App Router), TypeScript, Tailwind CSS, shadcn/ui
- **Infrastructure:** Kubernetes (Helm), Dapr, Keycloak (JWT), Redis
- **Local dev:** Tilt + Docker Desktop / OrbStack

## Prerequisites

- Go 1.26+
- Node.js 20+
- Docker Desktop or OrbStack
- `kubectl` + `helm`
- Tilt (`brew install tilt-dev/tap/tilt`)
- Dapr CLI (`brew install dapr/tap/dapr-cli`)

## Getting Started

### Install tools

```bash
make init        # installs protoc plugins, wire, ent CLI
```

### Local dev (Tilt)

```bash
make dev         # tilt up --continue (Delve starts immediately)
make debug       # tilt up (Delve waits for attach on :2345)
```

Tilt builds a Linux binary, syncs it into the running container, and port-forwards:

| Resource | Port |
|----------|------|
| HTTP | 8000 |
| gRPC | 9000 |
| Delve | 2345 |
| Postgres | 5432 |
| Redis | 6379 |
| pgAdmin | 5050 |
| Web (Next.js) | 3000 |

### Code generation

```bash
make api         # generate Go/gRPC/HTTP code + OpenAPI from proto files
make config      # generate Go structs from internal conf.proto files
make generate    # run go generate ./... + go mod tidy (wire + ent)
make all         # api + config + generate
```

### Build

```bash
make build       # builds all service binaries → ./bin/
```

### Per-service commands (run from `app/<service>/`)

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

## Architecture

```
api/<service>/v1/        ← protobuf contracts (source of truth)
  └─ *.pb.go, *_grpc.pb.go, *_http.pb.go  (generated)

app/<service>/
  cmd/server/            ← entrypoint + Wire DI wiring
  internal/
    conf/                ← config proto → generated conf.pb.go
    server/              ← HTTP + gRPC server setup
    service/             ← implements generated proto server interface
    biz/                 ← domain logic, repo interfaces (no infra deps)
    data/                ← repo implementations, ent ORM client, Dapr client
      ent/schema/        ← ent schema definitions (source of truth for DB)
      ent/               ← generated ORM code (do not edit by hand)
```

**Dependency rule:** `service → biz → data`. `biz` defines interfaces; `data` implements them.

### Cross-service calls

The `web` Next.js app acts as a BFF orchestrator for the checkout flow:
- `GET http://localhost:8002/v1/carts/:sessionId` (cart)
- `POST http://localhost:8004/v1/orders` (order)

Go services do not call each other directly.

### Authentication

Catalog read endpoints are public. Write endpoints require a JWT validated against Keycloak (`KEYCLOAK_JWKS_URL` injected via Dapr secret store at startup).

### Dapr

Each service expects a Dapr sidecar (gRPC on `DAPR_GRPC_PORT`, default `50001`). Dapr is used for:
- **Secret store** — DB credentials and Keycloak URL injected at startup
- **Pub/Sub** — declared but not yet active

Services retry the sidecar 12× at 5s intervals (60s total) and panic on failure.

## Project Structure

```
api/            ← protobuf definitions
app/            ← service implementations + web frontend
deploy/
  helm/         ← Helm chart (Postgres, Redis, pgAdmin)
  k8s/          ← Kubernetes manifests + Dapr component configs
  keycloak/     ← Keycloak realm config
pkg/            ← shared Go packages
third_party/    ← vendored proto imports (google, validate)
```
