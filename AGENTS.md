# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build all Go services
make build

# Tidy dependencies for all services
make tidy

# Run individual services
make catalog-api      # cd src/catalog && go run product.go -f etc/catalog-api.yaml
make cart-api         # cd src/cart && go run cart.go -f etc/cart-api.yaml
make payment-api      # cd src/payment && go run payment.go -f etc/payment-api.yaml

# Regenerate Ent ORM code (after modifying ent/schema/*.go)
make generate-ent

# Regenerate go-zero handlers/types/routes (after modifying .api files)
# CRITICAL: always use --style snakecase — other styles produce wrong file names
goctl api go -api product.api --dir . --style snakecase

# Run tests in a service
cd src/catalog && go test ./...

# Local dev with live reload (requires Tilt + Docker Desktop/OrbStack)
tilt up

# Frontend
cd src/storefront && npm run dev
cd src/storefront && npm run lint
cd src/storefront && npm run build
```

## Architecture

### Monorepo layout

```
src/                 # Go workspace (go.work)
  catalog/           # Product catalog service — module: catalog (NOT product)
  cart/              # Cart service — module: cart
  payment/           # Payment service — module: payment
  shared/            # Shared libs: auth, event, health, entschema
  storefront/        # Next.js frontend
deploy/k8s/          # Kubernetes manifests (Kustomize)
```

### go-zero Handler → Logic pattern

Each backend service uses the same layered structure:

```
internal/
  handler/<group>/   # HTTP handlers — auto-generated from .api file, do not hand-edit
  logic/<group>/     # Business logic — implement here
  types/             # Request/response types — auto-generated from .api file
  svc/               # ServiceContext: dependency injection (DB client, config, event bus)
  config/            # Config struct matching etc/*.yaml
  domain/            # Domain models (hand-written)
  event/             # Dapr event publishing (outbox pattern)
ent/
  schema/            # Ent entity definitions — hand-edit here
  *.go               # Ent ORM code — auto-generated, do not hand-edit
```

**Source of truth**: `.api` files define types, routes, and handler names. goctl regenerates `handler/`, `types/`, and route registration safely. Business logic in `logic/` is never overwritten.

### Adding an endpoint

1. Add type + handler directive to the `.api` file
2. Run `goctl api go -api <name>.api --dir . --style snakecase`
3. Implement the generated `logic/<group>/<name>_logic.go`

### Database changes

1. Edit `ent/schema/*.go`
2. Run `make generate-ent` (or `cd src/<service>/ent && go generate ./...`)
3. Update logic to use new fields

### Event bus

Services publish domain events via Dapr pub/sub using the outbox pattern. `src/shared/event/` contains the publisher. Redis backs the Dapr state store and pub/sub in all environments.

### Authentication

Keycloak issues JWTs. `src/shared/auth/` validates tokens via JWKS. Each service's `main.go` wires role-based middleware:
- Public: `GET /api/v1/products`, `GET /api/v1/categories`
- Admin: requires `admin` role
- Authenticated: all other routes

### Frontend

Next.js App Router. API routes in `src/app/api/` reverse-proxy to backend services. State management via Zustand. Design system: Apple-inspired binary palette (`#000000` / `#f5f5f7`) with single accent `#0071e3`; see `DESIGN.md`.

## Key conventions

- goctl style is always `snakecase` — forgetting this flag is the most common mistake
- Catalog module is `catalog`, not `product` (was renamed; fix stale imports with `go mod tidy`)
- Imports use module name as root: `catalog/internal/logic/catalog`, `shared/auth`
- API route prefix: `/api/v1` for all services
- Ent schemas use shared mixins from `shared/entschema` (timestamps, IDs)

## Skills

Domain-specific guidance lives in `.claude/skills/` and `.agents/skills/`. Load with the `Skill` tool when working in that area:

- Go: `golang`, `go-error-handling`, `go-naming`, `go-concurrency`, `golang-testing`, `golang-patterns`, `golang-pro`, `go-linting`, `zero-skills`, `testcontainers-go`
- Frontend: `next-best-practices`, `vercel-react-best-practices`, `typescript-expert`, `tailwind-design-system`, `ui-ux-pro-max`
- Other: `documentation-lookup`, `web-design-guidelines`
