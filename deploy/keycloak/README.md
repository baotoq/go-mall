# Keycloak for go-mall Development

## Quick Start

```bash
docker compose -f deploy/docker-compose.keycloak.yml up -d
```

Keycloak starts in dev mode on **http://localhost:8080**.

Realm `go-mall` is auto-imported on startup from `keycloak/realm-export.json`.

## Default Credentials

| Account | Username | Password |
|---------|----------|----------|
| Admin Console | `admin` | `admin` |
| Test User | `admin@test.com` | `test-password` |
| Test User | `user@test.com` | `test-password` |

**Admin Console:** http://localhost:8080/admin

## Client Credentials

| Client ID | Secret | Type |
|-----------|--------|------|
| catalog | `catalog-secret` | confidential (service accounts) |
| cart | `cart-secret` | confidential (service accounts) |
| payment | `payment-secret` | confidential (service accounts) |
| frontend | `frontend-secret` | public |

## Getting Access Tokens

### Password Grant (for testing/service accounts)

```bash
# Admin user
curl -X POST "http://localhost:8080/realms/go-mall/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=catalog" \
  -d "client_secret=catalog-secret" \
  -d "username=admin@test.com" \
  -d "password=test-password"

# Service account (no username/password)
curl -X POST "http://localhost:8080/realms/go-mall/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=catalog" \
  -d "client_secret=catalog-secret"
```

### Public Client (frontend)

```bash
curl -X POST "http://localhost:8080/realms/go-mall/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=frontend" \
  -d "username=user@test.com" \
  -d "password=test-password"
```

## JWKS Endpoint

Validate JWT tokens using:

```
http://localhost:8080/realms/go-mall/protocol/openid-connect/certs
```

## Health Check

```bash
curl http://localhost:8080/health/ready
```

## Stopping

```bash
docker compose -f deploy/docker-compose.keycloak.yml down
```

## Troubleshooting

**Realm not imported?**
Check logs: `docker compose -f deploy/docker-compose.keycloak.yml logs keycloak`

**Reset realm?**
Remove volume: `docker compose -f deploy/docker-compose.keycloak.yml down -v`

## Next Steps

- Update `etc/catalog-api.yaml` with Keycloak auth settings
- Configure `src/storefront` with frontend client settings