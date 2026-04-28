# go-mall Agent Instructions

## Project Overview

**go-mall** is a microservices e-commerce platform:
- **Backend**: Go 1.26.1 with go-zero framework
- **Frontend**: Next.js with TypeScript, Tailwind CSS, shadcn/ui
- **Monorepo**: Go workspace at `src/` directory

## Services (ports)

| Service | Port | Entry Point |
|---------|------|-------------|
| catalog | 8001 | `src/catalog/product.go -f etc/catalog-api.yaml` |
| cart | 8002 | `src/cart/cart.go -f etc/cart-api.yaml` |
| payment | 8003 | `src/payment/payment.go -f etc/payment-api.yaml` |

## CRITICAL: goctl Style

**Always use `snake_case` style** for goctl commands:
```bash
goctl api go -api product.api --dir . --style snakecase
```

Other styles (`go_zero`, `goZero`) produce wrong naming. This is the most commonly missed convention.

## Code Generation

### API services (from .api files)
```bash
# 1. Edit .api file
# 2. Regenerate types/handlers/routes (safe to re-run)
goctl api go -api product.api --dir . --style snakecase

# 3. Implement business logic in internal/logic/
# 4. Verify: cd src/catalog && go build ./...
```

### Ent ORM (from schema)
```bash
# After modifying ent/schema/*.go
make generate-ent
# or: cd src/catalog/ent && go generate ./...
```

## Module Names

- **Catalog service**: module `catalog` (NOT `product`)
- **Cart service**: module `cart`
- **Payment service**: module `payment`
- **Shared**: module `shared`
- **Storefront**: Next.js module (workspace)

## Common Commands

```bash
# Build all services
make build

# Tidy dependencies
make tidy

# Run a service locally
cd src/catalog && go run product.go -f etc/catalog-api.yaml

# Local dev with live reload (requires Tilt + Docker)
tilt up
```

## Architecture

### go-zero Service Structure
```
src/catalog/
├── product.api          # API spec (goctl source)
├── product.go           # Entry point
├── etc/*.yaml           # Config
├── ent/                 # Ent ORM (auto-generated)
│   └── schema/          # Database schemas
└── internal/
    ├── handler/         # HTTP handlers (generated)
    ├── logic/           # Business logic (implement here)
    ├── types/           # Generated types
    └── config/          # Config struct
```

### Frontend Structure
```
src/storefront/
├── src/app/             # Next.js app router pages
├── src/components/      # React components (shadcn/ui)
└── src/lib/             # API clients, stores
```

## Key Conventions

### Go
- Package naming: lowercase, no underscores
- Type naming: MixedCaps (`CreateProductRequest`)
- Imports: use module path (e.g., `catalog/internal/logic`)

### API Routes
- Routes defined via `@server(group: ...)` in `.api` files
- Prefix: `/api/v1` (catalog), `/api/v1/cart`, `/api/v1/payments`

### Database
- Ent ORM with SQLite by default (memory mode)
- Commented MySQL config in `etc/catalog-api.yaml`
- Dapr SDK integrated for event bus / outbox pattern

## Skills Available

Project skills in `.agents/skills/` and `.claude/skills/`:

**Go**: `golang`, `go-error-handling`, `go-naming`, `go-concurrency`, `golang-testing`, `golang-patterns`, `golang-pro`, `go-linting`, `zero-skills`, `testcontainers-go`

**Frontend**: `next-best-practices`, `vercel-react-best-practices`, `typescript-expert`, `tailwind-design-system`, `ui-ux-pro-max`

**Other**: `documentation-lookup`, `web-design-guidelines`

Load skills explicitly when working in their domain.

## Design System

Frontend follows Apple-inspired design documented in `DESIGN.md`:
- Binary color scheme: black (`#000000`) / light gray (`#f5f5f7`)
- Single accent: Apple Blue (`#0071e3`)
- Typography: SF Pro Display/Text (system fonts)
- Pill-shaped CTAs with 980px radius

## Troubleshooting

### Import errors after renaming
- All `product/` imports should be `catalog/`
- Run `go mod tidy` in affected services

### goctl produces wrong file names
- You forgot `--style snakecase`
- Re-run with correct style flag

### Types not found
- Types generated from `.api` via goctl
- Update `.api` file, then regenerate
- New types appear in `internal/types/`

## Other Important Files

- `copilot-instructions.md` - More detailed project overview
- `DESIGN.md` - Design system reference
- `.mcp.json` - MCP server configs (context7, go-zero, shadcn, stitch)
