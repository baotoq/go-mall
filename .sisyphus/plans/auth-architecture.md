# Keycloak Auth Architecture for go-mall

## TL;DR

> **Add cross-cutting authentication/authorization to all 3 go-mall microservices using Keycloak (OAuth2/OIDC) with JWT/JWKS verification.** Create a `shared/auth` package following the established `shared/event` pattern, wire custom go-zero middleware (NOT the built-in `@server(jwt:)` which is HS256-only), configure Keycloak via Docker Compose with realm export, and enforce route-level access control (public browse, admin mutations, user cart, service-to-service internal APIs).

> **Deliverables**:
> - `shared/auth/` package (JWT validator, RBAC middleware, client credentials)
> - Keycloak Docker Compose + realm export JSON
> - Auth middleware wired into catalog, cart, and payment services
> - Route splitting in all 3 `.api` files (public/protected/service-to-service)
> - Cart HTTP clients updated to propagate auth tokens
> - Unit tests with mocked auth (following `outbox_test.go` pattern)
> - Configuration structs updated in all services

> **Estimated Effort**: Large
> **Parallel Execution**: YES - 4 waves
> **Critical Path**: Task 1 → Task 2 → Task 5 → Tasks 8-10 → Task 11 → Task 12 → Final

---

## Context

### Original Request
Design and implement cross-cutting authentication/authorization architecture for the go-mall microservices platform using Keycloak.

### Interview Summary
**Key Discussions**:
- Auth provider: Keycloak (OAuth2/OIDC, handles user management, roles, token lifecycle)
- Role model: Simple roles (admin/user) stored in Keycloak only — no user DB tables
- Token model: JWT tokens issued by Keycloak, validated locally via JWKS public key verification
- Service-to-service: Hybrid — user token passthrough + service client credentials
- Keycloak deployment: Docker (dev+prod)
- Test strategy: Unit tests with mocked auth (no integration tests with real Keycloak in this phase)
- Frontend: DEFERRED — backend-only in this phase
- Route classification: Default split — public browsing, admin mutations, user cart, service-to-service internal APIs
- Keycloak down policy: Cache JWKS + fail-closed (503 when cache expires and Keycloak unreachable)

**Research Findings**:
- 🔴 ZERO auth exists currently — no middleware, no JWT, no sessions in any service
- 🔴 go-zero's built-in `@server(jwt:)` uses HS256 (shared secret) — INCOMPATIBLE with Keycloak RS256 — must use custom middleware
- ✅ `shared/event/` pattern proven and reusable for `shared/auth/` — interfaces + implementations + adapters
- ✅ go-zero `rest.Server` supports `server.Use()` for custom middleware
- ⚠️ All 3 event adapters are structurally identical — good pattern reference
- ⚠️ No Docker Compose or Dockerfiles exist yet — this is the first
- ⚠️ Cart's HTTP clients don't propagate auth headers — must be updated
- ⚠️ Cart uses `sessionId` for anonymous sessions — keeping as-is (defer cart ownership redesign)

### Metis Review
**Identified Gaps** (addressed):
- go-zero JWT incompatibility: RESOLVED — custom middleware with `go-oidc` for RS256/JWKS
- Route classification: RESOLVED — default split (public/admin/user/service-to-service)
- Keycloak unavailability: RESOLVED — cache JWKS + fail-closed
- Cart sessionId model: RESOLVED — keep as-is, don't redesign
- Docker patterns: RESOLVED — first Docker Compose, establish patterns

---

## Work Objectives

### Core Objective
Implement Keycloak-based authentication/authorization across all go-mall microservices with a shared, reusable auth package.

### Concrete Deliverables
- `src/shared/auth/` — JWT validation, RBAC middleware, client credentials (Go package)
- `deploy/docker-compose.keycloak.yml` — Keycloak Docker Compose for dev
- `deploy/keycloak/realm-export.json` — Keycloak realm configuration (realm, clients, roles)
- Updated `src/{catalog,cart,payment}/etc/*-api.yaml` — Keycloak config sections
- Updated `src/{catalog,cart,payment}/internal/config/config.go` — Keycloak config structs
- Updated `src/{catalog,cart,payment}/internal/svc/service_context.go` — Auth dependencies
- Updated `src/{catalog,cart,payment}/*.api` — Route splitting (public/protected/service-to-service)
- Updated `src/{catalog,cart,payment}/internal/handler/` — Regenerated after `.api` changes
- Updated `src/cart/internal/clients/catalog/client.go` — Auth header propagation
- Updated `src/cart/internal/clients/payment/client.go` — Auth header propagation
- `src/shared/auth/validator_test.go` — Unit tests for JWT validation
- `src/shared/auth/middleware_test.go` — Unit tests for RBAC middleware
- `src/shared/auth/client_test.go` — Unit tests for client credentials

### Definition of Done
- [ ] `go test ./shared/auth/...` → PASS in all 3 services
- [ ] `docker compose -f deploy/docker-compose.keycloak.yml up` → Keycloak healthy
- [ ] `curl http://localhost:8001/api/v1/products` → 200 (public, no auth)
- [ ] `curl http://localhost:8889/api/v1/cart/items` → 401 (protected, no token)
- [ ] `curl -H "Authorization: Bearer $ADMIN_TOKEN" -X POST http://localhost:8001/api/v1/products` → 403 for user, allowed for admin
- [ ] Cart checkout propagates user token to catalog/payment services
- [ ] Service-to-service calls use client credentials token

### Must Have
- Shared `auth` package in `src/shared/auth/` following `shared/event/` pattern
- Custom go-zero middleware with `go-oidc/v3` for JWKS verification
- Keycloak realm with `go-mall` realm, 4 clients (catalog, cart, payment, frontend), 2 roles (admin, user)
- Route-level auth: public, user-authenticated, admin-only, service-to-service
- Client credentials flow for cart→catalog and cart→payment calls
- JWKS caching with fail-closed policy
- Unit tests for shared auth package with mocked verifier
- Docker Compose for Keycloak dev environment
- Configuration structs in all 3 services

### Must NOT Have (Guardrails)
- ❌ Do NOT use go-zero's `@server(jwt: Auth)` directive — it uses HS256, incompatible with Keycloak RS256
- ❌ Do NOT create user/role tables in the database — roles live in Keycloak only
- ❌ Do NOT implement token refresh for user tokens — that's frontend's responsibility (deferred)
- ❌ Do NOT build login UI, logout endpoints, or user management APIs
- ❌ Do NOT add CORS middleware — frontend concern, deferred
- ❌ Do NOT modify the `sessionId` model — keep as-is, cart ownership redesign deferred
- ❌ Do NOT implement audit logging beyond standard go-zero request logging
- ❌ Do NOT add an API gateway — out of scope
- ❌ Do NOT add service discovery — out of scope
- ❌ Do NOT over-abstract — start minimal, extend later
- ❌ AI slop: no excessive comments, no over-engineering, no generic names like `data/result/item`

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.
> Acceptance criteria requiring "user manually tests/confirms" are FORBIDDEN.

### Test Decision
- **Infrastructure exists**: NO (no test framework beyond `go test`)
- **Automated tests**: YES (Tests-after)
- **Framework**: Go standard `testing` package + `go test`
- **Test pattern**: Interface-based fakes following `shared/event/outbox_test.go`

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **API/Backend**: Use Bash (curl) — Send requests, assert status + response fields
- **Go packages**: Use Bash (`go test`) — Run unit tests, verify PASS
- **Docker**: Use Bash (`docker compose`) — Start/stop, health checks
- **Config**: Use Bash — Verify file contents, Keycloak connectivity

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately - foundation):
├── Task 1: Keycloak Docker Compose + realm configuration [quick]
├── Task 2: Shared auth package — JWT validator [deep]
├── Task 3: Shared auth package — RBAC middleware [quick]
├── Task 4: Shared auth package — client credentials [quick]
└── Task 5: Configuration structs for all 3 services [quick]

Wave 2 (After Wave 1 - core wiring):
├── Task 6: Catalog service — route splitting + middleware wiring [unspecified-high]
├── Task 7: Cart service — route splitting + middleware wiring [unspecified-high]
├── Task 8: Payment service — route splitting + middleware wiring [unspecified-high]
└── Task 9: Cart HTTP clients — auth header propagation [quick]

Wave 3 (After Wave 2 - testing + polish):
├── Task 10: Unit tests — shared auth package [unspecified-high]
├── Task 11: Integration verification — all services end-to-end [deep]
└── Task 12: Documentation — auth architecture guide [writing]

Wave FINAL (After ALL tasks):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
```

### Dependency Matrix

| Task | Depends On | Blocks |
|------|-----------|--------|
| 1 | - | 2, 3, 4, 11 |
| 2 | 1 (JWKS endpoint config) | 6, 7, 8, 10 |
| 3 | 2 (validator interface) | 6, 7, 8, 10 |
| 4 | 2 (validator for token exchange) | 9, 10 |
| 5 | - | 6, 7, 8 |
| 6 | 2, 3, 5 | 11 |
| 7 | 2, 3, 5, 9 | 11 |
| 8 | 2, 3, 5 | 11 |
| 9 | 4 | 7 |
| 10 | 2, 3, 4 | 11 |
| 11 | 6, 7, 8, 10 | F3 |
| 12 | 11 | - |

### Agent Dispatch Summary

- **Wave 1**: 5 tasks — T1 `quick`, T2 `deep`, T3 `quick`, T4 `quick`, T5 `quick`
- **Wave 2**: 4 tasks — T6 `unspecified-high`, T7 `unspecified-high`, T8 `unspecified-high`, T9 `quick`
- **Wave 3**: 3 tasks — T10 `unspecified-high`, T11 `deep`, T12 `writing`
- **FINAL**: 4 tasks — F1 `oracle`, F2 `unspecified-high`, F3 `unspecified-high`, F4 `deep`

---

## TODOs

---

- [ ] 1. Keycloak Docker Compose + Realm Configuration

  **What to do**:
  - Create `deploy/docker-compose.keycloak.yml` with Keycloak 26.0 container
  - Configure Keycloak in dev mode (`start-dev`) on port 8080
  - Create `deploy/keycloak/realm-export.json` with:
    - Realm name: `go-mall`
    - 4 clients: `catalog`, `cart`, `payment` (confidential, service accounts enabled), `frontend` (public)
    - 2 realm roles: `admin`, `user`
    - Client scopes: `openid`, `profile`, `email`
    - Service account mappings for each backend client
  - Add health check to Docker Compose
  - Add test users: `admin@test.com` (admin role), `user@test.com` (user role), `test-password` for both
  - Create `deploy/keycloak/README.md` with instructions for running and importing the realm

  **Must NOT do**:
  - Do NOT create Dockerfiles for go-mall services (only Keycloak)
  - Do NOT configure MySQL for Keycloak (use embedded H2 in dev mode)
  - Do NOT expose Keycloak admin console outside localhost

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`golang`]
    - `golang`: Go module structure knowledge for understanding project layout

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 2-5, though Task 2 needs JWKS endpoint config from this task)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 2 (needs JWKS endpoint), 11 (needs running Keycloak)
  - **Blocked By**: None (can start immediately)

  **References**:
  **Pattern References**:
  - `src/shared/event/dapr.go` — Pattern for external service configuration (Dapr pubsub name constant)
  - `src/catalog/etc/catalog-api.yaml` — Current config format to extend with Keycloak section
  - `Makefile` — Build/run targets pattern (add `keycloak` targets)

  **API/Type References**:
  - `src/shared/event/store.go` — Interface + enum pattern (`MessageStatus`) to follow for auth types

  **External References**:
  - Keycloak Docker image: `quay.io/keycloak/keycloak:26.0`
  - Keycloak realm export format: https://www.keycloak.org/server/importExport
  - `go-oidc` package: https://github.com/coreos/go-oidc

  **WHY Each Reference Matters**:
  - `dapr.go`: Shows how external service config is embedded in the shared module, follow this pattern for Keycloak endpoints
  - `catalog-api.yaml`: Must add `Keycloak` section to this format, preserving existing fields
  - `Makefile`: Must add targets for `make keycloak-start` and `make keycloak-stop`
  - `store.go`: Follow the interface + enum pattern for auth types (Role, TokenStatus)

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Keycloak container starts and becomes healthy
    Tool: Bash (docker compose)
    Preconditions: Docker is running, ports 8080 and 8081 are free
    Steps:
      1. Run `docker compose -f deploy/docker-compose.keycloak.yml up -d`
      2. Wait 30 seconds for startup
      3. Run `docker compose -f deploy/docker-compose.keycloak.yml ps`
      4. Verify container status contains "healthy" or "running"
      5. Run `curl -s http://localhost:8080/realms/go-mall/.well-known/openid-configuration | jq .issuer`
    Expected Result: Issuer field contains "http://localhost:8080/realms/go-mall"
    Failure Indicators: Container not running, no response from well-known endpoint, missing issuer
    Evidence: .sisyphus/evidence/task-1-keycloak-healthy.txt

  Scenario: Keycloak realm has required clients and roles
    Tool: Bash (curl)
    Preconditions: Keycloak container is running
    Steps:
      1. Obtain admin token: `curl -s -X POST http://localhost:8080/realms/master/protocol/openid-connect/token -d "username=admin&password=admin&grant_type=password&client_id=admin-cli" | jq -r .access_token`
      2. List clients: `curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/admin/realms/go-mall/clients | jq '.[].clientId'`
      3. List roles: `curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/admin/realms/go-mall/roles | jq '.[].name'`
    Expected Result: Clients include "catalog", "cart", "payment", "frontend". Roles include "admin", "user".
    Failure Indicators: Missing clients, missing roles, realm not found
    Evidence: .sisyphus/evidence/task-1-realm-config.txt

  Scenario: Test user can obtain JWT token
    Tool: Bash (curl)
    Preconditions: Keycloak running with test users
    Steps:
      1. `curl -s -X POST http://localhost:8080/realms/go-mall/protocol/openid-connect/token -d "username=user@test.com&password=test-password&grant_type=password&client_id=frontend" | jq .access_token`
      2. Validate token structure by decoding: `echo $TOKEN | cut -d. -f2 | base64 -d 2>/dev/null | jq .`
    Expected Result: JWT contains `sub`, `realm_access.roles` with "user", `iss` matching Keycloak realm URL
    Failure Indicators: No access_token in response, invalid JWT structure, missing roles
    Evidence: .sisyphus/evidence/task-1-token-flow.txt
  ```

  **Commit**: YES (groups with 5)
  - Message: `feat(auth): add Keycloak Docker Compose and realm configuration`
  - Files: `deploy/docker-compose.keycloak.yml`, `deploy/keycloak/realm-export.json`, `deploy/keycloak/README.md`
  - Pre-commit: `docker compose -f deploy/docker-compose.keycloak.yml config` (validate YAML syntax)

---

- [ ] 2. Shared Auth Package — JWT Validator

  **What to do**:
  - Create `src/shared/auth/validator.go` with:
    - `TokenValidator` interface (following `shared/event/store.go` pattern)
    - `KeycloakValidator` struct implementing `TokenValidator`
    - OIDC provider initialization with JWKS discovery and caching
    - JWT validation: verify signature (RS256), issuer, audience, expiry
    - Role extraction: parse `realm_access.roles` from claims
    - Subject extraction: parse `sub` claim
    - Config struct: `KeycloakConfig{RealmURL, ClientID, JWKSCacheTTL}`
    - Lazy initialization with retry for Keycloak connectivity
    - Fail-closed behavior: return 503 when JWKS cache expired and Keycloak unreachable
  - Add `go-oidc/v3` and `golang-jwt/jwt/v5` dependencies to `src/shared/go.mod`
  - Context propagation: store validated claims in `context.Context`

  **Must NOT do**:
  - Do NOT use go-zero's built-in JWT middleware (HS256 incompatible)
  - Do NOT implement token refresh for user tokens
  - Do NOT create user/role database tables
  - Do NOT add login/logout endpoints

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`golang`, `go-error-handling`]
    - `golang`: Production Go patterns, error handling, interface design
    - `go-error-handling`: Wrapping errors with %w, sentinel errors for auth failures

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on Keycloak endpoint config from Task 1)
  - **Parallel Group**: Wave 1 (after Task 1 JWKS endpoint is known)
  - **Blocks**: Tasks 3, 4, 6, 7, 8, 10
  - **Blocked By**: Task 1 (needs Keycloak realm URL for JWKS discovery)

  **References**:
  **Pattern References**:
  - `src/shared/event/event.go` — Interface definition pattern (`Event` interface, `Dispatcher[T]` generic)
  - `src/shared/event/store.go` — Interface + enum pattern (`OutboxStore`, `MessageStatus` enum)
  - `src/shared/event/dapr.go` — External service integration (Dapr client init, config constants)
  - `src/shared/event/outbox.go` — Complex logic with retry pattern (follow for init retry)

  **API/Type References**:
  - `src/catalog/internal/config/config.go` — Config struct pattern (embed `rest.RestConf`)
  - `src/catalog/internal/svc/service_context.go` — ServiceContext pattern for dependency injection

  **Test References**:
  - `src/shared/event/outbox_test.go` — Interface-based fake pattern (`fakeStore`, `recordingDispatcher`)

  **External References**:
  - go-oidc: https://github.com/coreos/go-oidc — OIDC discovery and JWKS verification
  - golang-jwt: https://github.com/golang-jwt/jwt — JWT parsing and claim extraction
  - Keycloak JWKS endpoint: `http://localhost:8080/realms/go-mall/protocol/openid-connect/certs`

  **WHY Each Reference Matters**:
  - `event.go`: Shows generic interface pattern — `Dispatcher[T]` should inspire `TokenValidator` interface design
  - `store.go`: Follow `MessageStatus` enum pattern for auth error types (ErrTokenExpired, ErrInvalidToken, ErrInsufficientRoles)
  - `dapr.go`: Follow the external-service init pattern — lazy initialization, config constants, error wrapping
  - `outbox.go`: The retry logic (MaxRetryAttempts) should guide the Keycloak connectivity retry pattern
  - `config.go`: Must match the `rest.RestConf` embedding pattern for Keycloak config struct
  - `service_context.go`: Must follow this pattern for adding auth validator dependency
  - `outbox_test.go`: Follow the `fakeStore` + `recordingDispatcher` pattern for mock token validator

  **Acceptance Criteria**:

  **If TDD (tests enabled):**
  - [ ] Test file created: `src/shared/auth/validator_test.go`

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: JWT validator parses and validates a correct token
    Tool: Bash (go test)
    Preconditions: `src/shared/go.mod` updated with go-oidc and golang-jwt dependencies
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestKeycloakValidator_ValidToken -v`
      2. Verify test creates fake validator, passes valid JWT mock, and extracts claims
    Expected Result: Test PASS, claims contain correct subject and role
    Failure Indicators: Compilation error, test FAIL, missing claims
    Evidence: .sisyphus/evidence/task-2-validator-valid.txt

  Scenario: JWT validator rejects expired token
    Tool: Bash (go test)
    Preconditions: Same as above
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestKeycloakValidator_ExpiredToken -v`
      2. Verify test passes expired JWT and expects ErrTokenExpired
    Expected Result: Test PASS, ErrTokenExpired returned
    Failure Indicators: Test FAIL, wrong error type, panic
    Evidence: .sisyphus/evidence/task-2-validator-expired.txt

  Scenario: JWT validator rejects token with wrong audience
    Tool: Bash (go test)
    Preconditions: Same as above
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestKeycloakValidator_WrongAudience -v`
    Expected Result: Test PASS, ErrInvalidToken returned for wrong audience
    Failure Indicators: Test FAIL, token accepted incorrectly
    Evidence: .sisyphus/evidence/task-2-validator-audience.txt

  Scenario: Package compiles cleanly
    Tool: Bash (go build)
    Preconditions: All source files created
    Steps:
      1. `cd src/shared && go build ./auth/...`
    Expected Result: Exit code 0, no compilation errors
    Failure Indicators: Build errors, missing imports, type mismatches
    Evidence: .sisyphus/evidence/task-2-build.txt
  ```

  **Commit**: YES (groups with 3, 4)
  - Message: `feat(auth): add shared JWT validator with JWKS verification`
  - Files: `src/shared/auth/validator.go`, `src/shared/go.mod`, `src/shared/go.sum`
  - Pre-commit: `cd src/shared && go build ./auth/...`

---

- [ ] 3. Shared Auth Package — RBAC Middleware

  **What to do**:
  - Create `src/shared/auth/middleware.go` with:
    - `AuthMiddleware` function returning `func(next http.HandlerFunc) http.HandlerFunc` (go-zero middleware signature)
    - `RequireRole(roles ...string)` middleware — checks `realm_access.roles` claim
    - `RequireAuth()` middleware — validates token and injects claims into context
    - `RequireServiceAuth()` middleware — validates service client credentials token
    - Context key constants: `UserIDKey`, `UserRolesKey`
    - Helper functions: `GetUserIDFromContext(ctx)`, `GetUserRolesFromContext(ctx)`
    - Error response format: `{"code": 401, "message": "..."}` for 401, `{"code": 403, "message": "..."}` for 403
  - Middleware must extract `Authorization: Bearer <token>` header
  - Must handle: missing header → 401, invalid token → 401, wrong role → 403, expired → 401

  **Must NOT do**:
  - Do NOT add session-based auth
  - Do NOT implement rate limiting
  - Do NOT add CORS headers
  - Do NOT create route-specific middleware (just role-based)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`golang`, `go-error-handling`]
    - `golang`: Go middleware patterns for go-zero
    - `go-error-handling`: Proper error wrapping for auth failures

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 4, after Task 2)
  - **Parallel Group**: Wave 1 (after Task 2)
  - **Blocks**: Tasks 6, 7, 8, 10
  - **Blocked By**: Task 2 (needs TokenValidator interface)

  **References**:
  **Pattern References**:
  - `src/shared/event/event.go` — Generic interface pattern for `Dispatcher[T]`
  - `src/shared/event/dapr.go` — External service middleware integration

  **API/Type References**:
  - `src/shared/auth/validator.go` (Task 2) — `TokenValidator` interface to use in middleware
  - `src/catalog/product.api` — Current route grouping pattern (`@server(group: catalog)`)

  **External References**:
  - go-zero middleware: https://go-zero.dev/docs/tutorials/http/middleware/
  - go-zero `rest.Server.Use()`: registers global middleware

  **WHY Each Reference Matters**:
  - `event.go`: Interface design pattern for the middleware — follow how `Dispatcher` is defined
  - `dapr.go`: Shows how external dependencies are initialized — middleware initialization should follow similar pattern
  - `validator.go`: The middleware wraps the `TokenValidator` and must conform to its interface
  - `product.api`: Must understand how `@server()` groups work to know where middleware will be registered

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: RequireAuth middleware rejects request without Authorization header
    Tool: Bash (go test)
    Preconditions: Task 2 completed (TokenValidator interface exists)
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestRequireAuth_MissingHeader -v`
      2. Verify middleware returns 401 with `{"code": 401, "message": "missing authorization header"}`
    Expected Result: Test PASS, 401 status, correct error format
    Failure Indicators: Test FAIL, wrong status code, wrong error format
    Evidence: .sisyphus/evidence/task-3-middleware-noheader.txt

  Scenario: RequireRole middleware rejects user without admin role
    Tool: Bash (go test)
    Preconditions: Same
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestRequireRole_InsufficientRole -v`
      2. Verify middleware returns 403 with `{"code": 403, "message": "insufficient permissions"}`
    Expected Result: Test PASS, 403 status, correct error format
    Failure Indicators: Test FAIL, wrong status code, user With role accessing admin endpoint
    Evidence: .sisyphus/evidence/task-3-middleware-role.txt

  Scenario: RequireAuth middleware injects claims into context
    Tool: Bash (go test)
    Preconditions: Same
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestRequireAuth_ContextInjection -v`
      2. Verify next handler receives `UserIDKey` and `UserRolesKey` in context
    Expected Result: Test PASS, context contains user ID and roles
    Failure Indicators: Test FAIL, missing context values
    Evidence: .sisyphus/evidence/task-3-middleware-context.txt

  Scenario: RequireServiceAuth validates client credentials token
    Tool: Bash (go test)
    Preconditions: Task 4 completed (client credentials)
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestRequireServiceAuth_ValidToken -v`
    Expected Result: Test PASS, service token accepted
    Failure Indicators: Test FAIL, service token rejected
    Evidence: .sisyphus/evidence/task-3-middleware-service.txt
  ```

  **Commit**: YES (groups with 2, 4)
  - Message: `feat(auth): add RBAC middleware with role-based access control`
  - Files: `src/shared/auth/middleware.go`
  - Pre-commit: `cd src/shared && go build ./auth/...`

---

- [ ] 4. Shared Auth Package — Client Credentials Flow

  **What to do**:
  - Create `src/shared/auth/client.go` with:
    - `ServiceClient` struct for machine-to-machine authentication
    - `NewServiceClient(config ServiceClientConfig)` constructor
    - Token acquisition: POST to Keycloak `/protocol/openid-connect/token` with `grant_type=client_credentials`
    - Token caching: store token with expiry, refresh 30s before expiration
    - Thread-safe token cache using `sync.RWMutex`
    - `GetToken(ctx context.Context) (string, error)` method
  - Service client config: `ServiceClientConfig{RealmURL, ClientID, ClientSecret}`

  **Must NOT do**:
  - Do NOT implement token refresh for user tokens (only client credentials)
  - Do NOT cache tokens across services (each service has its own client instance)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`golang`, `go-concurrency`]
    - `golang`: Go struct and interface design
    - `go-concurrency`: Thread-safe token cache with `sync.RWMutex`

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 3)
  - **Parallel Group**: Wave 1 (after Task 2)
  - **Blocks**: Task 9 (cart HTTP clients), Task 10 (tests)
  - **Blocked By**: Task 2 (needs validator interface for token validation)

  **References**:
  **Pattern References**:
  - `src/shared/event/dapr.go` — External service client initialization pattern
  - `src/cart/internal/clients/catalog/client.go` — HTTP client pattern (timeout, error handling)

  **API/Type References**:
  - `src/shared/auth/validator.go` (Task 2) — TokenValidator interface
  - Keycloak token endpoint: `POST /realms/go-mall/protocol/openid-connect/token`

  **External References**:
  - Keycloak client credentials: https://www.keycloak.org/docs/latest/securing_apps/#_client_credentials
  - `sync.RWMutex` for thread-safe cache: https://pkg.go.dev/sync#RWMutex

  **WHY Each Reference Matters**:
  - `dapr.go`: Shows how to initialize an external client (Dapr) — follow for Keycloak HTTP client
  - `catalog/client.go`: Current HTTP client pattern (10s timeout, error wrapping) — match this pattern
  - `validator.go`: Client credentials tokens need the same validation as user tokens

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: ServiceClient acquires and caches client credentials token
    Tool: Bash (go test)
    Preconditions: Task 2 completed
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestServiceClient_GetToken -v`
      2. Verify token is cached and second call returns cached token
    Expected Result: Test PASS, token acquired and cached, second call uses cache
    Failure Indicators: Test FAIL, token not cached, new request on each call
    Evidence: .sisyphus/evidence/task-4-client-token.txt

  Scenario: ServiceClient refreshes token before expiry
    Tool: Bash (go test)
    Preconditions: Same
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestServiceClient_TokenRefresh -v`
      2. Set token expiry to near-expiry, verify new token is requested
    Expected Result: Test PASS, new token acquired when within 30s of expiry
    Failure Indicators: Test FAIL, stale token used
    Evidence: .sisyphus/evidence/task-4-client-refresh.txt

  Scenario: ServiceClient handles Keycloak unavailability
    Tool: Bash (go test)
    Preconditions: Same
    Steps:
      1. `cd src/shared && go test ./auth/ -run TestServiceClient_KeycloakDown -v`
      2. Simulate Keycloak unreachable, verify error returned
    Expected Result: Test PASS, clear error message about Keycloak unavailability
    Failure Indicators: Test FAIL, panic, nil error
    Evidence: .sisyphus/evidence/task-4-client-down.txt
  ```

  **Commit**: YES (groups with 2, 3)
  - Message: `feat(auth): add client credentials flow with token caching`
  - Files: `src/shared/auth/client.go`
  - Pre-commit: `cd src/shared && go build ./auth/...`

---

- [ ] 5. Configuration Structs for All 3 Services

  **What to do**:
  - Add `KeycloakConfig` struct to `src/shared/auth/config.go`:
    ```go
    type KeycloakConfig struct {
        RealmURL      string
        ClientID      string
        ClientSecret  string // empty for public clients
        JWKSCacheTTL  time.Duration
    }
    ```
  - Update `src/catalog/internal/config/config.go` to embed `KeycloakConfig`
  - Update `src/cart/internal/config/config.go` to embed `KeycloakConfig`
  - Update `src/payment/internal/config/config.go` to embed `KeycloakConfig`
  - Update all 3 YAML config files to add Keycloak section:
    ```yaml
    Keycloak:
      RealmURL: "http://keycloak:8080/realms/go-mall"
      ClientID: "catalog"  # or "cart", "payment"
      ClientSecret: "xxx"  # from Keycloak client config
      JWKSCacheTTL: 24h
    ```
  - Update all 3 `ServiceContext` structs to add `Validator` and `ServiceClient` fields
  - Update all 3 `main.go` files to initialize auth dependencies

  **Must NOT do**:
  - Do NOT modify Ent schemas or database models
  - Do NOT change the existing config format — only ADD the Keycloak section
  - Do NOT hardcode Keycloak URLs in Go code — they belong in YAML config

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`golang`, `go-naming`]
    - `golang`: Go struct embedding patterns
    - `go-naming`: Consistent naming for config fields

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 2-4, no code dependency)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 6, 7, 8 (need config for wiring)
  - **Blocked By**: None (can start immediately)

  **References**:
  **Pattern References**:
  - `src/catalog/internal/config/config.go` — Current config struct pattern (embed `rest.RestConf`)
  - `src/cart/internal/config/config.go` — Adding custom fields (`CatalogBaseURL`, `PaymentBaseURL`)
  - `src/payment/internal/config/config.go` — Mirrors catalog config pattern

  **API/Type References**:
  - `src/catalog/etc/catalog-api.yaml` — Current YAML format to extend
  - `src/cart/etc/cart-api.yaml` — Current YAML format (has service URLs)
  - `src/payment/etc/payment-api.yaml` — Current YAML format

  **Test References**:
  - `src/shared/event/outbox_test.go` — Test pattern for shared packages

  **External References**:
  - go-zero config: https://go-zero.dev/docs/tutorials/http/config/

  **WHY Each Reference Matters**:
  - `config.go` files: Must follow the exact pattern of embedding `rest.RestConf` and adding custom fields
  - YAML files: Must preserve existing format and add Keycloak section in same style
  - `cart config`: Shows how to add service-specific config (URLs) — follow this pattern for Keycloak URLs

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: All 3 services compile with new config
    Tool: Bash (go build)
    Preconditions: Config structs and YAML files updated
    Steps:
      1. `cd src/catalog && go build ./...`
      2. `cd src/cart && go build ./...`
      3. `cd src/payment && go build ./...`
    Expected Result: All 3 build without errors
    Failure Indicators: Compilation error, missing imports, type mismatch
    Evidence: .sisyphus/evidence/task-5-build.txt

  Scenario: ServiceContext initializes auth dependencies
    Tool: Bash (go test)
    Preconditions: ServiceContext updated
    Steps:
      1. `cd src/catalog && go test ./internal/svc/ -v`
      2. Verify ServiceContext has Validator and ServiceClient fields
    Expected Result: Tests pass (or compilation succeeds if no existing tests)
    Failure Indicators: Nil pointer, missing field, init error
    Evidence: .sisyphus/evidence/task-5-svccontext.txt

  Scenario: YAML config loads successfully
    Tool: Bash (go test)
    Preconditions: YAML files updated
    Steps:
      1. Write a small test that loads each YAML with `conf.MustLoad`
      2. Verify `KeycloakConfig` fields are populated
    Expected Result: Config loads without error, Keycloak fields populated
    Failure Indicators: Parse error, missing field, default values used
    Evidence: .sisyphus/evidence/task-5-config.txt
  ```

  **Commit**: YES (with Task 1)
  - Message: `feat(auth): add Keycloak config structs and YAML to all services`
  - Files: `src/shared/auth/config.go`, `src/catalog/internal/config/config.go`, `src/cart/internal/config/config.go`, `src/payment/internal/config/config.go`, `src/catalog/etc/catalog-api.yaml`, `src/cart/etc/cart-api.yaml`, `src/payment/etc/payment-api.yaml`, `src/catalog/internal/svc/service_context.go`, `src/cart/internal/svc/service_context.go`, `src/payment/internal/svc/service_context.go`, `src/catalog/main.go`, `src/cart/cart.go`, `src/payment/payment.go`
  - Pre-commit: `cd src/catalog && go build ./... && cd ../cart && go build ./... && cd ../payment && go build ./...`

---

- [ ] 6. Catalog Service — Route Splitting + Middleware Wiring

  **What to do**:
  - Split `src/catalog/product.api` into multiple `@server()` groups:
    - **Public group** (no auth): `GET /products`, `GET /products/:id`, `GET /products/by-slug/:slug`, `GET /categories`, `GET /categories/:id`
    - **Admin group** (admin role): `POST /products`, `PUT /products/:id`, `DELETE /products/:id`, `POST /products/:id/increase-stock`, `POST /categories`
    - **Service group** (client credentials): `POST /reservations`, `POST /reservations/:id/confirm`, `POST /reservations/:id/cancel`
  - Do NOT add `jwt: Auth` directive (HS256 incompatible) — use custom middleware
  - Regenerate handlers: `goctl api go -api product.api --dir . --style snakecase`
  - Wire middleware in `src/catalog/main.go`:
    - Public routes: no middleware
    - Admin routes: `RequireRole("admin")`
    - Service routes: `RequireServiceAuth()`
  - Update ServiceContext to include Validator, init in `main.go`

  **Must NOT do**:
  - Do NOT use `@server(jwt: Auth)` — it generates HS256 middleware
  - Do NOT modify business logic in `internal/logic/` — only middleware wiring
  - Do NOT change product/reservation data models

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang`, `zero-skills`]
    - `golang`: Go middleware patterns
    - `zero-skills`: go-zero `.api` file format, `goctl` code generation, route group patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 7, 8)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 11 (integration verification)
  - **Blocked By**: Tasks 2, 3, 5 (needs validator, middleware, config)

  **References**:
  **Pattern References**:
  - `src/catalog/product.api` — Current route grouping (single `@server(group: catalog)`)
  - `src/catalog/internal/handler/` — Generated handler structure

  **API/Type References**:
  - `src/shared/auth/middleware.go` (Task 3) — `RequireAuth()`, `RequireRole()`, `RequireServiceAuth()`
  - `src/shared/auth/validator.go` (Task 2) — `TokenValidator` interface
  - `src/catalog/internal/config/config.go` (Task 5) — `KeycloakConfig` struct

  **Test References**:
  - `src/shared/event/outbox_test.go` — Interface-based test pattern

  **External References**:
  - go-zero .api format: https://go-zero.dev/docs/tutorials/api/grammar/
  - goctl route groups: https://go-zero.dev/docs/tutorials/api/route/

  **WHY Each Reference Matters**:
  - `product.api`: Current single-group route definition that must be split into 3 groups
  - `handler/`: Will be regenerated by goctl — understand structure to know what changes
  - `middleware.go`: The middleware functions to wire into routes
  - `config.go`: Must read KeycloakConfig to initialize validator

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Public routes accessible without auth
    Tool: Bash (curl)
    Preconditions: Catalog service running with auth middleware wired
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" http://localhost:8001/api/v1/products`
      2. `curl -s -o /dev/null -w "%{http_code}" http://localhost:8001/api/v1/categories`
    Expected Result: 200 for both public routes without Authorization header
    Failure Indicators: 401 or 403 response
    Evidence: .sisyphus/evidence/task-6-public-routes.txt

  Scenario: Admin routes require admin role
    Tool: Bash (curl)
    Preconditions: Service running, test tokens available
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8001/api/v1/products -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" -d '{}'`
      2. `curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8001/api/v1/products -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"name":"test","slug":"test","description":"test","image_url":"http://test","price":10,"total_stock":100}'`
    Expected Result: Step 1 returns 403 (user role on admin endpoint), Step 2 returns 201 or 4xx (admin role allowed)
    Failure Indicators: 401 for admin token (token not validated), 200/201 for user token (insufficient role check)
    Evidence: .sisyphus/evidence/task-6-admin-routes.txt

  Scenario: Service routes accept client credentials token
    Tool: Bash (curl)
    Preconditions: Service running, service token available
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8001/api/v1/reservations -H "Authorization: Bearer $SERVICE_TOKEN" -H "Content-Type: application/json" -d '{}'`
    Expected Result: 200 or 4xx (accepted auth, business logic may fail with invalid data)
    Failure Indicators: 401 (service token not recognized)
    Evidence: .sisyphus/evidence/task-6-service-routes.txt
  ```

  **Commit**: YES
  - Message: `feat(auth): wire catalog service auth middleware with route splitting`
  - Files: `src/catalog/product.api`, `src/catalog/internal/handler/` (regenerated), `src/catalog/main.go`, `src/catalog/internal/svc/service_context.go`
  - Pre-commit: `cd src/catalog && go build ./...`

---

- [ ] 7. Cart Service — Route Splitting + Middleware Wiring

  **What to do**:
  - Split `src/cart/cart.api` into route groups:
    - **User group** (authenticated user): `GET /items`, `POST /items`, `PATCH /items`, `DELETE /items/:productId`, `POST /checkout`
  - All cart routes require user authentication (no public access to cart data)
  - Regenerate handlers: `goctl api go -api cart.api --dir . --style snakecase`
  - Wire `RequireAuth()` middleware on all user routes in `src/cart/cart.go`
  - Update ServiceContext and `main.go`

  **Must NOT do**:
  - Do NOT add anonymous cart access (defer cart ownership redesign)
  - Do NOT modify cart business logic
  - Do NOT change cart Ent schemas

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang`, `zero-skills`]
    - `golang`: Go middleware wiring
    - `zero-skills`: go-zero .api format and goctl

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 6, 8)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 11 (integration verification)
  - **Blocked By**: Tasks 2, 3, 5, 9 (needs validator, middleware, config, AND updated HTTP clients)

  **References**:
  **Pattern References**:
  - `src/cart/cart.api` — Current route definition
  - `src/cart/internal/clients/catalog/client.go` — HTTP client to update (Task 9 dependency)

  **API/Type References**:
  - `src/shared/auth/middleware.go` (Task 3) — `RequireAuth()` middleware
  - `src/cart/internal/config/config.go` (Task 5) — Updated `KeycloakConfig`

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Cart routes require authentication
    Tool: Bash (curl)
    Preconditions: Cart service running with auth middleware
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" http://localhost:8889/api/v1/cart/items`
      2. `curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $USER_TOKEN" http://localhost:8889/api/v1/cart/items`
    Expected Result: Step 1 returns 401 (no auth), Step 2 returns 200 (with auth)
    Failure Indicators: 200 without auth, 401 with valid token
    Evidence: .sisyphus/evidence/task-7-auth-required.txt

  Scenario: Cart checkout propagates auth to downstream services
    Tool: Bash (curl)
    Preconditions: All 3 services running, Keycloak running
    Steps:
      1. Add item to cart with auth: `curl -s -H "Authorization: Bearer $USER_TOKEN" -X POST http://localhost:8889/api/v1/cart/items -H "Content-Type: application/json" -d '{"productId":"...","quantity":1}'`
      2. Checkout: `curl -s -H "Authorization: Bearer $USER_TOKEN" -X POST http://localhost:8889/api/v1/cart/checkout`
      3. Verify catalog and payment services receive propagated auth headers (check logs)
    Expected Result: Checkout completes, downstream services receive valid auth
    Failure Indicators: 401 from downstream services, missing auth headers
    Evidence: .sisyphus/evidence/task-7-checkout-propagation.txt
  ```

  **Commit**: YES
  - Message: `feat(auth): wire cart service auth middleware`
  - Files: `src/cart/cart.api`, `src/cart/internal/handler/` (regenerated), `src/cart/cart.go`, `src/cart/internal/svc/service_context.go`
  - Pre-commit: `cd src/cart && go build ./...`

---

- [ ] 8. Payment Service — Route Splitting + Middleware Wiring

  **What to do**:
  - Split `src/payment/payment.api` into route groups:
    - **User group** (authenticated user): `GET /:id` (view own payment)
    - **Service group** (client credentials): `POST /` (create payment), `GET /` (list/query payments)
  - Regenerate handlers: `goctl api go -api payment.api --dir . --style snakecase`
  - Wire middleware in `src/payment/payment.go`

  **Must NOT do**:
  - Do NOT add public payment access
  - Do NOT modify payment business logic
  - Do NOT add payment user-scoping logic (defer — current phase just adds auth middleware)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang`, `zero-skills`]

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 6, 7)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 11 (integration verification)
  - **Blocked By**: Tasks 2, 3, 5 (needs validator, middleware, config)

  **References**:
  **Pattern References**:
  - `src/payment/payment.api` — Current route definition

  **API/Type References**:
  - `src/shared/auth/middleware.go` (Task 3) — `RequireAuth()`, `RequireServiceAuth()`
  - `src/payment/internal/config/config.go` (Task 5) — Updated config

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Payment creation requires service credentials
    Tool: Bash (curl)
    Preconditions: Payment service running with auth middleware
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8890/api/v1/payments -H "Content-Type: application/json" -d '{}'`
      2. `curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8890/api/v1/payments -H "Authorization: Bearer $SERVICE_TOKEN" -H "Content-Type: application/json" -d '{"idempotency_key":"test-1","total_amount":100,"currency":"USD"}'`
    Expected Result: Step 1 returns 401 (no auth), Step 2 returns 201 or 4xx (business logic may fail)
    Failure Indicators: 201 without auth, 401 with service token
    Evidence: .sisyphus/evidence/task-8-service-auth.txt

  Scenario: Payment view requires user auth
    Tool: Bash (curl)
    Preconditions: Same
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" http://localhost:8890/api/v1/payments/nonexistent`
      2. `curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $USER_TOKEN" http://localhost:8890/api/v1/payments/nonexistent`
    Expected Result: Step 1 returns 401, Step 2 returns 404 (auth passed, payment not found)
    Failure Indicators: 200 without auth
    Evidence: .sisyphus/evidence/task-8-user-auth.txt
  ```

  **Commit**: YES
  - Message: `feat(auth): wire payment service auth middleware`
  - Files: `src/payment/payment.api`, `src/payment/internal/handler/` (regenerated), `src/payment/payment.go`, `src/payment/internal/svc/service_context.go`
  - Pre-commit: `cd src/payment && go build ./...`

---

- [ ] 9. Cart HTTP Clients — Auth Header Propagation

  **What to do**:
  - Update `src/cart/internal/clients/catalog/client.go`:
    - Accept `auth.TokenSource` or extract token from `context.Context`
    - Inject `Authorization: Bearer <token>` header in `do()` method
    - For reservation endpoints (service-to-service): use `ServiceClient.GetToken()` to get client credentials token
    - For user-propagated calls: forward user's token from request context
  - Update `src/cart/internal/clients/payment/client.go`:
    - Same pattern: accept token source, inject auth header
    - Use client credentials for payment creation (service-to-service)
  - Update `src/cart/internal/svc/service_context.go`:
    - Store `ServiceClient` reference for client credentials
    - Pass auth context to HTTP clients

  **Must NOT do**:
  - Do NOT add auth logic in the business logic layer — keep it in clients
  - Do NOT modify the `do()` method signature beyond adding auth header
  - Do NOT create a shared HTTP client package — keep changes minimal and local

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`golang`, `go-concurrency`]
    - `golang`: Context propagation patterns
    - `go-concurrency`: Thread-safe token caching (already in ServiceClient from Task 4)

  **Parallelization**:
  - **Can Run In Parallel**: NO (needs Task 4's ServiceClient)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 7 (cart service wiring needs updated clients)
  - **Blocked By**: Task 4 (needs ServiceClient for client credentials)

  **References**:
  **Pattern References**:
  - `src/cart/internal/clients/catalog/client.go` — Current HTTP client with `do()` method
  - `src/cart/internal/clients/payment/client.go` — Same pattern, has `Idempotency-Key` header (follow for auth header)

  **API/Type References**:
  - `src/shared/auth/client.go` (Task 4) — `ServiceClient.GetToken()` for client credentials
  - `src/shared/auth/middleware.go` (Task 3) — Context keys for user token extraction

  **Why Each Reference Matters**:
  - `catalog/client.go`: Must add auth header injection to the `do()` method without changing its signature
  - `payment/client.go`: Same pattern, already has `Idempotency-Key` — follow for auth header addition
  - `client.go` (Task 4): Source of client credentials tokens for service-to-service calls

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Cart client propagates auth header to catalog
    Tool: Bash (go test)
    Preconditions: Task 4 completed (ServiceClient available)
    Steps:
      1. `cd src/cart && go test ./internal/clients/catalog/ -run TestClient_AuthHeader -v`
      2. Verify HTTP request includes `Authorization: Bearer` header
    Expected Result: Test PASS, auth header present in outgoing request
    Failure Indicators: Test FAIL, missing auth header, wrong token type
    Evidence: .sisyphus/evidence/task-9-catalog-auth.txt

  Scenario: Cart client uses service credentials for reservation calls
    Tool: Bash (go test)
    Preconditions: Same
    Steps:
      1. `cd src/cart && go test ./internal/clients/catalog/ -run TestClient_ServiceAuth -v`
      2. Verify service-to-service calls use client credentials token
    Expected Result: Test PASS, client credentials token used for reservation endpoints
    Failure Indicators: Test FAIL, user token used instead of service token
    Evidence: .sisyphus/evidence/task-9-service-creds.txt

  Scenario: Cart client propagates auth header to payment
    Tool: Bash (go test)
    Preconditions: Same
    Steps:
      1. `cd src/cart && go test ./internal/clients/payment/ -run TestClient_AuthHeader -v`
    Expected Result: Test PASS, auth header present in payment client requests
    Failure Indicators: Test FAIL, missing auth header
    Evidence: .sisyphus/evidence/task-9-payment-auth.txt
  ```

  **Commit**: YES
  - Message: `feat(auth): propagate auth headers in cart HTTP clients`
  - Files: `src/cart/internal/clients/catalog/client.go`, `src/cart/internal/clients/payment/client.go`, `src/cart/internal/svc/service_context.go`
  - Pre-commit: `cd src/cart && go build ./...`

---

- [ ] 10. Unit Tests — Shared Auth Package

  **What to do**:
  - Create `src/shared/auth/validator_test.go`:
    - `TestKeycloakValidator_ValidToken` — valid JWT returns claims
    - `TestKeycloakValidator_ExpiredToken` — returns ErrTokenExpired
    - `TestKeycloakValidator_WrongAudience` — returns ErrInvalidToken
    - `TestKeycloakValidator_MalformedToken` — returns ErrInvalidToken
    - `TestKeycloakValidator_WrongSigningKey` — returns ErrInvalidToken
    - `TestKeycloakValidator_KeycloakDown` — returns ErrKeycloakUnavailable (cached JWKS works, then fails)
  - Create `src/shared/auth/middleware_test.go`:
    - `TestRequireAuth_MissingHeader` — returns 401
    - `TestRequireAuth_InvalidToken` — returns 401
    - `TestRequireRole_InsufficientRole` — returns 403 (user on admin endpoint)
    - `TestRequireRole_AdminAccess` — passes through for admin
    - `TestRequireRole_UserAccess` — passes through for user endpoints
    - `TestRequireServiceAuth_ValidToken` — passes through for service token
    - `TestRequireServiceAuth_InvalidToken` — returns 401
    - `TestRequireAuth_ContextInjection` — verifies user ID and roles in context
  - Create `src/shared/auth/client_test.go`:
    - `TestServiceClient_GetToken` — acquires and caches client credentials token
    - `TestServiceClient_TokenRefresh` — refreshes before expiry
    - `TestServiceClient_KeycloakDown` — returns error when Keycloak unreachable
    - `TestServiceClient_ConcurrentAccess` — thread-safe token cache under concurrent access
  - All tests use `FakeTokenValidator` and `FakeKeycloakServer` following `outbox_test.go` pattern

  **Must NOT do**:
  - Do NOT require a running Keycloak instance for unit tests (use fakes)
  - Do NOT add integration tests (those would be in Task 11)
  - Do NOT use real network calls in unit tests

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang`, `golang-testing`]
    - `golang`: Interface design for testability
    - `golang-testing`: Table-driven tests, test helpers, fakes

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 6-8)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 11 (integration verification needs all tests passing)
  - **Blocked By**: Tasks 2, 3, 4 (needs all shared auth code)

  **References**:
  **Pattern References**:
  - `src/shared/event/outbox_test.go` — Interface-based fake pattern (`fakeStore`, `recordingDispatcher`)
  - `src/shared/event/store.go` — Interface pattern to follow for fakes

  **Test References**:
  - `src/shared/auth/validator.go` (Task 2) — `TokenValidator` interface to fake
  - `src/shared/auth/middleware.go` (Task 3) — Middleware functions to test
  - `src/shared/auth/client.go` (Task 4) — `ServiceClient` to test

  **Why Each Reference Matters**:
  - `outbox_test.go`: THE pattern to follow for testing shared packages — interface-based fakes, no real dependencies
  - `store.go`: Shows the `OutboxStore` interface that was faked — same pattern for `TokenValidator`

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: All shared auth unit tests pass
    Tool: Bash (go test)
    Preconditions: All auth package code implemented
    Steps:
      1. `cd src/shared && go test ./auth/... -v -count=1`
    Expected Result: All tests PASS, 0 failures, coverage > 70%
    Failure Indicators: Any test FAIL, build error, race condition
    Evidence: .sisyphus/evidence/task-10-unit-tests.txt

  Scenario: No race conditions in concurrent token access
    Tool: Bash (go test -race)
    Preconditions: Same
    Steps:
      1. `cd src/shared && go test ./auth/... -race -count=1`
    Expected Result: All tests PASS, no race detector warnings
    Failure Indicators: Race condition detected, test FAIL
    Evidence: .sisyphus/evidence/task-10-race-tests.txt
  ```

  **Commit**: YES
  - Message: `test(auth): add unit tests for shared auth package`
  - Files: `src/shared/auth/validator_test.go`, `src/shared/auth/middleware_test.go`, `src/shared/auth/client_test.go`
  - Pre-commit: `cd src/shared && go test ./auth/... -v`

---

- [ ] 11. Integration Verification — All Services End-to-End

  **What to do**:
  - Ensure all 3 services start with Keycloak configuration
  - Verify the full auth flow:
    1. Start Keycloak Docker Compose
    2. Import realm configuration
    3. Start all 3 go-mall services
    4. Test public endpoints (no auth needed)
    5. Test protected endpoints with admin token (full access)
    6. Test protected endpoints with user token (limited access)
    7. Test service-to-service with client credentials
    8. Test expired token → 401
    9. Test invalid token → 401
    10. Test wrong role → 403
  - Update `Makefile` with auth-related targets:
    - `make keycloak-start` — start Keycloak Docker
    - `make keycloak-stop` — stop Keycloak Docker
    - `make keycloak-seed` — create test users and roles
    - `make test-auth` — run full auth integration verification

  **Must NOT do**:
  - Do NOT add integration test files to the project (this is manual verification via curl)
  - Do NOT modify test infrastructure (just Makefile targets)
  - Do NOT create automated CI/CD pipeline (defer to future)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`golang`, `zero-skills`]
    - `golang`: Service startup verification
    - `zero-skills`: go-zero service configuration

  **Parallelization**:
  - **Can Run In Parallel**: NO (verifies all services together)
  - **Parallel Group**: Wave 3 (after all Wave 2 tasks)
  - **Blocks**: Final verification tasks (F1-F4)
  - **Blocked By**: Tasks 6, 7, 8, 10 (all services wired, all tests passing)

  **References**:
  **Pattern References**:
  - `Makefile` — Current build/run targets
  - `deploy/docker-compose.keycloak.yml` (Task 1) — Keycloak Docker Compose
  - `deploy/keycloak/realm-export.json` (Task 1) — Realm configuration

  **API/Type References**:
  - `src/catalog/product.api` — Public and protected route groups (Task 6)
  - `src/cart/cart.api` — User-authenticated routes (Task 7)
  - `src/payment/payment.api` — Service-authenticated routes (Task 8)

  **Why Each Reference Matters**:
  - `Makefile`: Must add `keycloak-*` targets following the existing target pattern
  - Docker Compose + realm: Must be running before integration verification
  - `.api` files: Define which routes are public/protected — verification must test each group

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Full auth flow — public endpoints accessible without auth
    Tool: Bash (curl)
    Preconditions: Keycloak running, all 3 services running
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" http://localhost:8001/api/v1/products`
      2. `curl -s -o /dev/null -w "%{http_code}" http://localhost:8001/api/v1/categories`
    Expected Result: 200 for both
    Failure Indicators: 401/403 (auth middleware blocking public routes)
    Evidence: .sisyphus/evidence/task-11-public-flow.txt

  Scenario: Full auth flow — admin can create products
    Tool: Bash (curl)
    Preconditions: Admin token obtained from Keycloak
    Steps:
      1. Get admin token: `ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/realms/go-mall/protocol/openid-connect/token -d "username=admin@test.com&password=test-password&grant_type=password&client_id=frontend" | jq -r .access_token)`
      2. Create product: `curl -s -w "%{http_code}" -X POST http://localhost:8001/api/v1/products -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"name":"Test Product","slug":"test-product","description":"Test","image_url":"http://test","price":10,"total_stock":100}'`
    Expected Result: 201 status code
    Failure Indicators: 403 (admin role not recognized), 401 (token invalid)
    Evidence: .sisyphus/evidence/task-11-admin-flow.txt

  Scenario: Full auth flow — user cannot create products (403)
    Tool: Bash (curl)
    Preconditions: User token obtained from Keycloak
    Steps:
      1. Get user token: `USER_TOKEN=$(curl -s -X POST http://localhost:8080/realms/go-mall/protocol/openid-connect/token -d "username=user@test.com&password=test-password&grant_type=password&client_id=frontend" | jq -r .access_token)`
      2. Attempt create product: `curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8001/api/v1/products -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" -d '{}'`
    Expected Result: 403 status code
    Failure Indicators: 201 (user accessing admin endpoint), 401 (token not validated)
    Evidence: .sisyphus/evidence/task-11-user-denied.txt

  Scenario: Full auth flow — expired/invalid token returns 401
    Tool: Bash (curl)
    Preconditions: Services running
    Steps:
      1. `curl -s -o /dev/null -w "%{http_code}" http://localhost:8889/api/v1/cart/items -H "Authorization: Bearer invalid.token.here"`
    Expected Result: 401 status code with `{"code": 401, "message": "..."}`
    Failure Indicators: 200 (invalid token accepted), 500 (server error)
    Evidence: .sisyphus/evidence/task-11-invalid-token.txt

  Scenario: Service-to-service auth — client credentials flow
    Tool: Bash (curl)
    Preconditions: All services running, service client configured
    Steps:
      1. Get service token: `SERVICE_TOKEN=$(curl -s -X POST http://localhost:8080/realms/go-mall/protocol/openid-connect/token -d "grant_type=client_credentials&client_id=catalog&client_secret=<secret>" | jq -r .access_token)`
      2. Create reservation: `curl -s -w "%{http_code}" -X POST http://localhost:8001/api/v1/reservations -H "Authorization: Bearer $SERVICE_TOKEN" -H "Content-Type: application/json" -d '{"session_id":"test","items":[]}'`
    Expected Result: Auth passes (200 or business logic response)
    Failure Indicators: 401 (service token not accepted)
    Evidence: .sisyphus/evidence/task-11-s2s-flow.txt

  Scenario: All services start successfully with Keycloak config
    Tool: Bash
    Preconditions: Keycloak running
    Steps:
      1. `cd src/catalog && go build ./...`
      2. `cd src/cart && go build ./...`
      3. `cd src/payment && go build ./...`
      4. Start each service and verify no startup errors
    Expected Result: All 3 services build and start without auth configuration errors
    Failure Indicators: Build error, startup panic, Keycloak connection error
    Evidence: .sisyphus/evidence/task-11-services-start.txt
  ```

  **Commit**: YES (with Task 12)
  - Message: `feat(auth): add integration verification and Makefile auth targets`
  - Files: `Makefile`
  - Pre-commit: `make build`

---

- [ ] 12. Documentation — Auth Architecture Guide

  **What to do**:
  - Create `docs/auth-architecture.md` with:
    - Architecture overview diagram (text-based)
    - Auth flow: user → Keycloak → JWT → service → JWKS verify
    - Service-to-service flow: service → Keycloak client credentials → JWT → target service
    - Route auth matrix (which routes are public/admin/user/service-to-service)
    - Configuration reference (KeycloakConfig struct fields, YAML examples)
    - Local development setup (Docker Compose, test users, getting tokens)
    - How to add a new service to auth
    - How to add a new protected route
    - Common errors and solutions (expired token, wrong audience, Keycloak down)

  **Must NOT do**:
  - Do NOT create a README (the architecture doc is sufficient)
  - Do NOT add API documentation (go-zero generates this)
  - Do NOT write a runbook (production operations guide, defer)

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []
    - No special skills needed — documentation task

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 11)
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: None (can start once auth architecture is finalized in earlier tasks)

  **References**:
  **Pattern References**:
  - `AGENTS.md` — Project documentation style and conventions
  - `copilot-instructions.md` — More detailed project overview (if exists)

  **API/Type References**:
  - All `.api` files — Route definitions for auth matrix
  - `src/shared/auth/` — All auth package interfaces for documentation

  **Why Each Reference Matters**:
  - `AGENTS.md`: Follow this style for the auth architecture doc
  - `.api` files: Must document the exact route-to-auth mapping
  - Auth package: Must document the public API for other developers

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Auth architecture doc covers all required sections
    Tool: Bash (grep)
    Preconditions: Doc written
    Steps:
      1. `grep -c "Architecture Overview" docs/auth-architecture.md`
      2. `grep -c "Route Auth Matrix" docs/auth-architecture.md`
      3. `grep -c "Service-to-Service" docs/auth-architecture.md`
      4. `grep -c "Local Development" docs/auth-architecture.md`
      5. `grep -c "How to Add" docs/auth-architecture.md`
    Expected Result: Each grep returns 1 or more (section exists)
    Failure Indicators: Any grep returns 0 (missing section)
    Evidence: .sisyphus/evidence/task-12-doc-coverage.txt

  Scenario: Documentation matches actual implementation
    Tool: Bash (diff)
    Preconditions: All previous tasks completed
    Steps:
      1. Verify route auth matrix matches `.api` files
      2. Verify configuration keys match `KeycloakConfig` struct
      3. Verify Keycloak URLs match Docker Compose config
    Expected Result: No mismatches between doc and code
    Failure Indicators: Documented routes differ from .api files, config keys don't match
    Evidence: .sisyphus/evidence/task-12-doc-accuracy.txt
  ```

  **Commit**: YES (with Task 11)
  - Message: `docs(auth): add auth architecture guide`
  - Files: `docs/auth-architecture.md`
  - Pre-commit: None

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go build ./...` in each service + `go vet ./...` + `go test ./shared/auth/...`. Review all changed files for: `interface{}` instead of `any`, empty error catches, `fmt.Println` in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names (data/result/item/temp).
  Output: `Build [PASS/FAIL] | Vet [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration (features working together, not isolation). Test edge cases: expired token, invalid token, missing header, wrong role. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance: no `@server(jwt:)` in .api files, no user DB tables, no login UI, no CORS middleware, no rate limiting. Detect cross-task contamination: Task N touching Task M's files.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **1+5**: `feat(auth): add Keycloak Docker Compose, realm config, and service config structs` — `deploy/`, `src/shared/auth/config.go`, `src/*/internal/config/`, `src/*/etc/`, `src/*/internal/svc/`, `src/*/main.go`
- **2+3+4**: `feat(auth): add shared auth package (JWT validator, RBAC middleware, client credentials)` — `src/shared/auth/`
- **6**: `feat(auth): wire catalog service auth middleware with route splitting` — `src/catalog/`
- **7**: `feat(auth): wire cart service auth middleware` — `src/cart/`
- **8**: `feat(auth): wire payment service auth middleware` — `src/payment/`
- **9**: `feat(auth): propagate auth headers in cart HTTP clients` — `src/cart/internal/clients/`
- **10**: `test(auth): add unit tests for shared auth package` — `src/shared/auth/*_test.go`
- **11+12**: `feat(auth): add integration verification, Makefile targets, and architecture docs` — `Makefile`, `docs/auth-architecture.md`

---

## Success Criteria

### Verification Commands
```bash
# All services build
cd src/catalog && go build ./... && cd ../cart && go build ./... && cd ../payment && go build ./... && cd ../shared && go build ./...

# Shared auth tests pass
cd src/shared && go test ./auth/... -v -count=1

# No race conditions
cd src/shared && go test ./auth/... -race -count=1

# Keycloak starts
docker compose -f deploy/docker-compose.keycloak.yml up -d
curl -s http://localhost:8080/realms/go-mall/.well-known/openid-configuration | jq .issuer

# Public endpoints (no auth)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8001/api/v1/products  # Expected: 200

# Protected endpoints (no auth → 401)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8889/api/v1/cart/items  # Expected: 401

# Admin endpoints (user token → 403)
curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $USER_TOKEN" -X POST http://localhost:8001/api/v1/products  # Expected: 403
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent (no `@server(jwt:)`, no user DB tables, no login UI, no CORS, no rate limiting, no sessionId redesign)
- [ ] All unit tests pass (`go test ./auth/...`)
- [ ] No race conditions detected (`go test -race ./auth/...`)
- [ ] Keycloak starts and serves JWKS endpoint
- [ ] Public routes return 200 without auth
- [ ] Protected routes return 401 without auth
- [ ] Admin-only routes return 403 with user token
- [ ] Service-to-service routes accept client credentials token
- [ ] Auth architecture documentation complete