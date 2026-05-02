# Plan: Bind Dapr pub/sub to Kratos HTTP server on port 8000

## Problem
- `daprhttp.NewService(":8002")` / `":8001"` start independent HTTP servers the Dapr sidecar never reaches
- `dapr.io/app-port: "8000"` — sidecar calls port 8000 for `/dapr/subscribe` and event delivery
- `NewServiceWithMux` requires `*chi.Mux`, incompatible with Kratos

## Solution
Register Dapr subscription routes directly on the Kratos HTTP server via `srv.HandleFunc`.
A thin HTTP adapter converts CloudEvent JSON → `*common.TopicEvent` → existing inbox-wrapped handlers.

## Files to Create
1. `pkg/dapr/handler.go` — shared CloudEvent adapter + subscription helpers (both services)

## Files to Modify
### Order service
2. `app/order/internal/server/subscriber.go` — remove daprhttp.NewService; register on Kratos srv
3. `app/order/internal/server/http.go` — accept *OrderSubscriber to trigger registration
4. `app/order/internal/server/server.go` — keep NewOrderSubscriber in ProviderSet
5. `app/order/cmd/server/main.go` — remove sub from kratos.Server(...)
6. `app/order/cmd/server/wire_gen.go` — regenerate (run make wire)

### Payment service
7. `app/payment/internal/server/subscriber.go` — same pattern
8. `app/payment/internal/server/http.go` — accept *PaymentSubscriber
9. `app/payment/internal/server/server.go` — keep NewPaymentSubscriber in ProviderSet
10. `app/payment/cmd/server/main.go` — remove sub from kratos.Server(...)
11. `app/payment/cmd/server/wire_gen.go` — regenerate

### Tests
12. `app/order/internal/server/subscriber_test.go` — update newTestSubscriber
13. `app/payment/internal/server/subscriber_test.go` — update (if exists)

## New Design

### pkg/dapr/handler.go
```
cloudEvent struct { ID, Source, Type, DataContentType, Topic, PubsubName, Data json.RawMessage, DataBase64 string }
TopicHandler(h common.TopicEventHandler) http.HandlerFunc
  - decode CloudEvent body
  - set e.ID, e.Source, e.Type, e.RawData = ce.Data
  - handle data_base64 fallback
  - call h(ctx, &evt)
  - write {"status":"SUCCESS"/"RETRY"/"DROP"}
Subscription struct { PubsubName, Topic, Route string }
WriteSubscriptions(w, []Subscription)
```

### OrderSubscriber (new shape)
```go
type OrderSubscriber struct { wfc, dlq, inbox, log }
// No addr, no svc field

func NewOrderSubscriber(srv *kratoshttp.Server, wfc, dlq, inbox, logger) *OrderSubscriber
  - registers GET /dapr/subscribe
  - registers POST /dapr/events/payment/completed
  - registers POST /dapr/events/payment/failed

// NO Start(), NO Stop() — no longer transport.Server
```

### newApp (order)
```go
// sub *OrderSubscriber kept as parameter to force Wire construction
// but NOT passed to kratos.Server(...)
kratos.Server(hs, gs, ww, ps, rs)  // sub removed
```

## Sequence
1. Create pkg/dapr/handler.go
2. Rewrite order subscriber.go
3. Update order http.go + server.go + main.go
4. Rewrite payment subscriber.go
5. Update payment http.go + server.go + main.go
6. make wire for both services
7. Update subscriber tests
8. go build ./... + go test ./...
