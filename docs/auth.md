# go-mall Authentication & Authorization

## Overview

Services use OAuth2/OIDC via Keycloak. JWTs are signed with RS256. Each service validates tokens independently using JWKS discovery. Roles live in Keycloak only — no user/role tables in application databases.

- **Protocol**: OpenID Connect
- **Token format**: JWT (RS256)
- **Identity provider**: Keycloak 26.0
- **Middleware**: Custom go-zero middleware in `shared/auth/`

## Keycloak Setup

```bash
cd deploy && docker compose -f docker-compose.keycloak.yml up -d
```

- Admin console: http://localhost:8080/admin (admin / admin)
- Realm: `go-mall`
- JWKS endpoint: http://localhost:8080/realms/go-mall/protocol/openid-connect/certs

### Clients

| Client | Type | Service |
|--------|------|---------|
| catalog | confidential | catalog-api |
| cart | confidential | cart-api |
| payment | confidential | payment-api |
| frontend | public | Next.js storefront |

### Roles

- `admin` — product/category management
- `user` — shopping cart, checkout

### Test Users

| User | Roles | Password |
|------|-------|----------|
| admin-user | admin | test-password |
| test-user | user | test-password |

## Shared Auth Package

`src/shared/auth/` provides validator, middleware, and client credentials flow.

### TokenValidator

```go
validator := auth.NewKeycloakValidator(auth.KeycloakConfig{
    RealmURL:     "http://localhost:8080/realms/go-mall",
    ClientID:     "catalog",
    ClientSecret: "...",
})
claims, err := validator.Validate(ctx, rawJWT)
```

Lazy-initialises OIDC provider on first call. Caches JWKS with 24h TTL.

### Middleware

```go
server.Use(auth.RequireAuth(validator))     // any authenticated user
server.Use(auth.RequireRole(validator, "admin")) // admin only
server.Use(auth.RequireServiceAuth(validator))   // service-to-service
```

### ServiceClient (Client Credentials)

```go
client := auth.NewServiceClient(auth.ServiceClientConfig{
    KeycloakConfig: cfg.Keycloak,
})
token, err := client.GetToken(ctx)
```

Caches tokens and refreshes 30s before expiry.

## Service Auth Wiring

### Catalog (port 8001)

| Route | Auth |
|-------|------|
| GET /api/v1/products/* | public |
| GET /api/v1/categories/* | public |
| POST /api/v1/products | admin |
| PUT /api/v1/products/:id | admin |
| DELETE /api/v1/products/:id | admin |
| POST /api/v1/categories | admin |
| POST /api/v1/products/:id/increase-stock | admin |
| POST /api/v1/reservations | protected |
| POST /api/v1/reservations/:id/confirm | protected |
| POST /api/v1/reservations/:id/cancel | protected |

### Cart (port 8889)

| Route | Auth |
|-------|------|
| GET /api/v1/cart/items | public |
| PATCH /api/v1/cart/items | protected |
| DELETE /api/v1/cart/items/:productId | protected |
| POST /api/v1/cart/checkout | protected |

### Payment (port 8890)

| Route | Auth |
|-------|------|
| All routes | service-to-service (RequireServiceAuth) |

Payment is only called by cart service. Every request must carry a valid service token.

## Service-to-Service Auth

Cart service acquires a token via `ServiceClient` and propagates it to catalog and payment clients:

```go
if token, err := ctx.ServiceClient.GetToken(context.Background()); err == nil {
    ctx.CatalogClient.SetAuthToken(token)
    ctx.PaymentClient.SetAuthToken(token)
}
```

The clients inject `Authorization: Bearer <token>` into every outgoing Dapr HTTP request.

## Testing

### Get a user token

```bash
curl -X POST http://localhost:8080/realms/go-mall/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=frontend" \
  -d "username=test-user" \
  -d "password=test-password"
```

### Get a service token

```bash
curl -X POST http://localhost:8080/realms/go-mall/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=cart" \
  -d "client_secret=<cart-secret>"
```

### Test a public route

```bash
curl http://localhost:8001/api/v1/products
```

### Test a protected route

```bash
# Without token — expect 401
curl -X POST http://localhost:8889/api/v1/cart/checkout

# With token — expect 200
curl -X POST http://localhost:8889/api/v1/cart/checkout \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Test an admin route

```bash
# User token — expect 403
curl -X POST http://localhost:8001/api/v1/products \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d '{"name":"X","slug":"x","price":1,"totalStock":1}'

# Admin token — expect 200
curl -X POST http://localhost:8001/api/v1/products \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"X","slug":"x","price":1,"totalStock":1}'
```

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| 503 on all requests | Keycloak unreachable or JWKS cache expired | Check Keycloak container health, restart service to re-discover OIDC provider |
| 401 on valid token | Token signed by wrong realm / expired | Verify `iss` claim matches realm URL, check token expiry |
| 403 on admin route | User lacks `admin` role | Add role in Keycloak console or use admin test user |
| Cart → catalog 401 | Service token not propagated | Check `ServiceClient.GetToken` works, verify `SetAuthToken` called before requests |
| "oidc discovery failed" on startup | Keycloak not ready | Start Keycloak before services; validator lazy-init retries on first request |
