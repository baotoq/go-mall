# Go-Mall Agent Instructions

## Project Overview

**go-mall** is a microservices e-commerce platform built with:
- **Backend**: Go 1.26.1 with go-zero framework (microservices architecture)
- **Frontend**: Next.js with TypeScript, Tailwind CSS, and shadcn/ui

## Architecture

### Services
Each service in `/src/` follows go-zero's **Handler → Logic → Model** pattern:

- **catalog** (formerly product) - Catalog/product management service
  - Port: 8001
  - API Definition: `src/catalog/product.api`
  - Config: `src/catalog/etc/catalog-api.yaml`
  - Database: Ent ORM with SQLite (memory by default)
  - Event Bus: Dapr for outbox messaging pattern
  - Structure:
    ```
    src/catalog/
    ├── product.api          # API specification (goctl)
    ├── product.go           # Main entry point
    ├── ent/                 # Ent ORM (auto-generated)
    │   ├── product.go       # Product entity
    │   └── schema/          # Database schema
    ├── internal/
    │   ├── handler/         # HTTP handlers
    │   ├── logic/           # Business logic
    │   ├── domain/          # Domain models
    │   ├── svc/             # Service context (DI)
    │   ├── event/           # Event handling
    │   ├── config/          # Configuration
    │   └── types/           # Generated types
    ```

### Frontend
- **storefront** - Next.js 16+ e-commerce frontend
  - Responsive design with Tailwind CSS
  - Component library with shadcn/ui patterns
  - API integration with backend services

## Key Conventions

### Go Code
- **Package naming**: lowercase, no underscores (e.g., `catalog`, not `catalog_service`)
- **Type naming**: MixedCaps (e.g., `CreateProductRequest`, `ProductInfo`)
- **Imports**: Use module path (e.g., `catalog/internal/logic/product`)
- **Error handling**: Follow Google/Uber Go style guides (see `.agents/skills/go-error-handling/`)

### go-zero Patterns
- **Handlers**: Accept typed request, return typed response + error
- **Logic**: Business logic separated from HTTP concerns
- **Service Context**: Dependency injection container for database, config, event bus
- **Generated Code**: Not hand-edited (handlers, types from `.api` files via goctl)

### API Design
- **Route group**: defined by `@server(group: ...)` in `.api` files
- **Handlers**: one handler per endpoint, naming: `CreateProduct`, `GetProduct`, etc.
- **Request/Response Types**: auto-generated from `.api` specifications

## Common Tasks

### Adding a new endpoint
1. Update `src/catalog/product.api` with new type + handler directive
2. Run `goctl api go -api product.api -dir .` to regenerate types/routes
3. Implement logic in `internal/logic/product/`
4. Implement handler in `internal/handler/product/`

### Database changes
1. Modify schema in `src/catalog/ent/schema/`
2. Run `cd src/catalog/ent && go generate ./...` to regenerate Ent classes
3. Update logic layer to use new fields

### Testing
- Use table-driven tests (see `.agents/skills/golang-testing/`)
- Testcontainers for integration tests
- Mock Dapr for event testing

## Available Skills

Skills are located in `.agents/skills/` and `.claude/skills/`:
- **go-zero**: go-zero framework patterns and best practices
- **go-error-handling**: Error handling from Google/Uber style guides
- **go-naming**: Go naming conventions
- **golang**: General Go best practices
- **golang-testing**: Testing patterns (table-driven, subtests, benchmarks)
- **vercel-react-best-practices**: React/Next.js optimization
- **next-best-practices**: Next.js 16 file conventions and patterns

## Build & Development

### Local Development
```bash
# Start services with Tilt (requires Tilt & Docker)
tilt up

# Generate code
make generate-ent

# Run catalog service
cd src/catalog && go run product.go -f etc/catalog-api.yaml
```

### Available Tools
- **Makefile**: `generate-ent`, `build`, `tidy`
- **Tiltfile**: Local development with live reload
- **goctl**: Code generation from `.api` files

## Module Names
- **Catalog Service**: module `catalog` (in `src/catalog/go.mod`)
- **Storefront**: module workspace
- Do NOT use old `product` module name

## Troubleshooting

### Import errors after rename
- All `product/` imports should now be `catalog/`
- Run `go mod tidy` in affected services
- Check that all `.go` files have been updated

### Type not found
- Types are generated from `product.api` via goctl
- Update the `.api` file, then regenerate with goctl
- New types appear in `internal/types/`

## Questions?
- Refer to `.agents/skills/` for detailed domain-specific guides
- Check existing handler/logic implementations as examples
- Use skills when implementing new features (e.g., testing, error handling, naming)
