# Shared Healthz Handler — Design Spec

**Date:** 2026-05-01
**Status:** Approved

## Problem

All four services (catalog, cart, order, payment) contain an identical `/healthz` handler inline in their `internal/server/http.go`. Any change (e.g., adding a response body) must be applied in four places.

## Solution

Extract the handler into a shared `pkg/server` package so all services reference one implementation.

## Package

- **Path:** `gomall/pkg/server`
- **File:** `pkg/server/health.go`
- **Package name:** `server`

```go
package server

import "net/http"

func Healthz(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("I'm alive"))
}
```

## Response

- HTTP status: `200 OK`
- Body: `I'm alive`

## Integration

Each service's `internal/server/http.go` imports with an alias (to avoid collision with the local `package server` declaration):

```go
import pkgserver "gomall/pkg/server"

srv.HandleFunc("/healthz", pkgserver.Healthz)
```

## Files Changed

| Action | File |
|--------|------|
| Create | `pkg/server/health.go` |
| Update | `app/catalog/internal/server/http.go` |
| Update | `app/cart/internal/server/http.go` |
| Update | `app/order/internal/server/http.go` |
| Update | `app/payment/internal/server/http.go` |

## Out of Scope

- Dependency health checks (DB, Redis) — not needed; 200 always means "process is alive"
- Shared server options builder — deferred, would require a shared conf type
