# Draft: Auth/Authorization Architecture

## Requirements (confirmed)
- Area: Cross-cutting Authentication/Authorization for go-mall microservices
- Scope: Catalog (8001), Cart (8002), Payment (8003), storefront frontend
- Auth provider: Keycloak (OAuth2/OIDC)
- Access control: Simple roles (admin/user)
- Frontend: Next.js storefront needs auth integration

## Technical Decisions
- **Auth Provider**: Keycloak (handles OAuth2, OIDC, user management, roles, sessions)
- **Role Model**: Simple roles — admin vs regular user (Keycloak realm roles)
- **Token Model**: JWT tokens issued by Keycloak, validated by each service

## Research Findings

### Current Auth State: NONE
- No auth middleware in any service (catalog, cart, payment)
- No JWT validation, no session management
- All endpoints are completely open
- Cart's HTTP clients send NO auth headers to catalog/payment

### Existing Patterns to Leverage
- `shared/event/` — well-designed outbox pattern with Dapr (reuse this pattern for shared auth)
- `svc/service_context.go` — each service has dependency injection via ServiceContext
- `logx` structured logging used consistently across all services
- go-zero `rest.Server` supports middleware chains (`server.Use()`)
- Ent ORM in all services (could add user/role entities)

### Duplicated Code to Consolidate
- All 3 `internal/event/adapter.go` files are structurally identical
- Opportunity: consolidate into shared package as pattern reference

### Service Communication
- Cart → Catalog/Payment: plain HTTP, 10s timeout, hardcoded localhost URLs
- No service discovery, no API gateway
- Dapr used for async events, not for service-to-service invocation

### Frontend
- Next.js storefront in `src/storefront/`
- No auth integration currently

### Database
- All services use SQLite in-memory (dev mode)
- Commented MySQL config exists

## Decisions (all confirmed)
- **Auth provider**: Keycloak (OAuth2/OIDC, user management, roles)
- **Role model**: Simple roles (admin/user) stored in Keycloak
- **Token model**: JWT issued by Keycloak, validated via JWKS locally
- **Service-to-service**: Hybrid (user token passthrough + service client credentials)
- **Keycloak deployment**: Docker (dev+prod)
- **Token validation**: JWT signature verification via JWKS (local, no network call per request)
- **Test strategy**: Unit tests with mocked auth
- **Frontend**: DEFERRED — backend-only in this phase

## Scope Boundaries
- **INCLUDE**: Keycloak Docker setup, shared auth middleware package, JWT/JWKS validation, service-to-service auth (client credentials), role-based access (admin/user), Keycloak realm configuration, unit tests
- **EXCLUDE**: Next.js storefront auth (deferred), API gateway, service discovery, user Ent schemas (roles live in Keycloak only)