# Cart & Payment Microservices Development Plan

## 1. Directory Tree
```text
go-mall/
├── src/cart/
│   ├── cart.api
│   ├── cart.go
│   ├── etc/
│   │   └── cart-api.yaml
│   ├── go.mod
│   ├── ent/
│   │   ├── schema/
│   │   │   ├── cartitem.go
│   │   │   ├── outboxmessage.go
│   │   │   └── mixin/
│   │   │       └── id.go
│   │   └── generate.go
│   └── internal/
│       ├── config/
│       │   └── config.go
│       ├── event/
│       │   ├── event.go
│       │   ├── dapr.go
│       │   └── outbox.go
│       ├── handler/
│       │   ├── routes.go
│       │   └── ... (generated handlers)
│       ├── logic/
│       │   └── ... (generated logic)
│       ├── svc/
│       │   └── service_context.go
│       └── types/
│           └── types.go
└── src/payment/
    ├── payment.api
    ├── payment.go
    ├── etc/
    │   └── payment-api.yaml
    ├── go.mod
    ├── ent/
    │   ├── schema/
    │   │   ├── payment.go
    │   │   ├── outboxmessage.go
    │   │   └── mixin/
    │   │       └── id.go
    │   └── generate.go
    └── internal/
        ├── config/
        │   └── config.go
        ├── event/
        │   ├── event.go
        │   ├── dapr.go
        │   └── outbox.go
        ├── handler/
        │   ├── routes.go
        │   └── ... (generated handlers)
        ├── logic/
        │   └── ... (generated logic)
        ├── svc/
        │   └── service_context.go
        └── types/
            └── types.go
```

## 2. API Specifications

### `src/cart/cart.api`
```go
syntax = "v1"

info (
	title:   "Cart API"
	desc:    "Shopping cart management service"
	author:  "go-mall"
	version: "v1"
)

type (
	CartItemRequest {
		ProductId  string `json:"productId"`
		Quantity   int64  `json:"quantity"`
	}
	CartItemResponse {
		Id        string `json:"id"`
		ProductId string `json:"productId"`
		Quantity  int64  `json:"quantity"`
		CreatedAt int64  `json:"createdAt"`
		UpdatedAt int64  `json:"updatedAt"`
	}
	CartItemListResponse {
		Items []CartItemResponse `json:"items"`
		Total int64              `json:"total"`
	}
	DeleteCartItemRequest {
		ProductId string `path:"productId"`
	}
)

@server (
	group:  cart
	prefix: /api/v1/cart
)
service cart-api {
	@doc "Get cart items"
	@handler GetCart
	get /items returns (CartItemListResponse)

	@doc "Add item to cart"
	@handler AddToCart
	post /items (CartItemRequest) returns (CartItemResponse)

	@doc "Update cart item quantity"
	@handler UpdateCartItem
	patch /items (CartItemRequest) returns (CartItemResponse)

	@doc "Delete item from cart"
	@handler DeleteCartItem
	delete /items/:productId (DeleteCartItemRequest)
}
```
*(Note: `X-Session-Id` header is passed automatically by go-zero `httpx.Parse` if mapped via header tag, but often we parse it directly from `r.Header.Get("X-Session-Id")` inside the handler logic as it's required for all routes. We'll explicitly validate it in each handler).*

### `src/payment/payment.api`
```go
syntax = "v1"

info (
	title:   "Payment API"
	desc:    "Payment processing service"
	author:  "go-mall"
	version: "v1"
)

type (
	LineItem {
		ProductId string  `json:"productId"`
		Quantity  int64   `json:"quantity"`
		Price     float64 `json:"price"`
	}
	CreatePaymentRequest {
		TotalAmount float64    `json:"totalAmount"`
		Currency    string     `json:"currency"`
		Items       []LineItem `json:"items"`
	}
	PaymentResponse {
		Id          string  `json:"id"`
		TotalAmount float64 `json:"totalAmount"`
		Currency    string  `json:"currency"`
		Status      string  `json:"status"`
		CreatedAt   int64   `json:"createdAt"`
	}
	GetPaymentRequest {
		Id string `path:"id"`
	}
	GetPaymentsRequest {
		IdempotencyKey string `form:"idempotency_key"`
	}
	GetPaymentsResponse {
		Payments []PaymentResponse `json:"payments"`
	}
)

@server (
	group:  payment
	prefix: /api/v1/payments
)
service payment-api {
	@doc "Create a payment"
	@handler CreatePayment
	post / (CreatePaymentRequest) returns (PaymentResponse)

	@doc "Get payment status"
	@handler GetPayment
	get /:id (GetPaymentRequest) returns (PaymentResponse)

	@doc "List payments by idempotency key"
	@handler GetPayments
	get / (GetPaymentsRequest) returns (GetPaymentsResponse)
}
```

## 3. Ent Schema Definitions

### Cart (`src/cart/ent/schema/cartitem.go`)
**Justification for single `CartItem` entity:**
Using a single `CartItem` entity with a `session_id` field and index simplifies the domain since the cart is tied to anonymous guest sessions. A dedicated `Cart` parent entity provides no additional value and adds unnecessary relational complexity and join overhead.
```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"cart/ent/schema/mixin"
	"github.com/google/uuid"
)

type CartItem struct {
	ent.Schema
}

func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").NotEmpty(),
		field.UUID("product_id", uuid.UUID{}),
		field.Int64("quantity").Positive(),
	}
}

func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
		index.Fields("session_id", "product_id").Unique(),
	}
}

func (CartItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}
```

### Payment (`src/payment/ent/schema/payment.go`)
```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"payment/ent/schema/mixin"
)

type Payment struct {
	ent.Schema
}

func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.String("idempotency_key").NotEmpty(),
		field.Float("total_amount").Positive(),
		field.String("currency").Default("USD"),
		field.String("status").Default("pending"),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("idempotency_key").Unique(),
	}
}

func (Payment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}
```

### Shared `OutboxMessage` (`ent/schema/outboxmessage.go` in BOTH services)
```go
package schema

import (
	"encoding/json"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	// import appropriate mixin package per service
)

type MessageStatus string
const (
	StatusPending    MessageStatus = "pending"
	StatusProcessing MessageStatus = "processing"
	StatusSent       MessageStatus = "sent"
	StatusFailed     MessageStatus = "failed"
)

func (MessageStatus) Values() []string {
	return []string{string(StatusPending), string(StatusProcessing), string(StatusSent), string(StatusFailed)}
}

type OutboxMessage struct {
	ent.Schema
}

func (OutboxMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_name"),
		field.JSON("payload", json.RawMessage{}),
		field.Int32("retry_attempts").Default(0),
		field.Enum("status").GoType(MessageStatus("")),
		field.Time("sent_at").Optional(),
	}
}

func (OutboxMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}
```

## 4. ServiceContext Shape

### Cart ServiceContext
```go
type ServiceContext struct {
	Config     config.Config
	Db         *ent.Client
	Dispatcher event.Dispatcher[event.Event]
}
```

### Payment ServiceContext
```go
type ServiceContext struct {
	Config          config.Config
	Db              *ent.Client
	Dispatcher      event.Dispatcher[event.Event]
	PaymentProvider provider.PaymentProvider
}
```

## 5. Logic-Layer Behavior Spec

### Cart Service
- **Headers**: All handlers extract `X-Session-Id` from HTTP request. If empty/missing, return `httpx.Error` with `401 Unauthorized` or `400 Bad Request`.
- **AddToCart**: Look up `CartItem` by `session_id` + `product_id`. If exists, increment `quantity`. If not, create new.
- **UpdateCartItem**: Find `CartItem` by `session_id` + `product_id`. Set `quantity` to specified value. If `quantity` <= 0, delete the item.
- **DeleteCartItem**: Find and delete `CartItem` by `session_id` + `product_id`. If missing, return successful.
- **GetCart**: Fetch all `CartItem` records for the `session_id`. Calculate total quantity.

### Payment Service
- **Headers**: `CreatePayment` extracts `Idempotency-Key`. Return error if missing.
- **CreatePayment**: 
  1. Check for existing payment by `idempotency_key`. If found, return it immediately (idempotency rule).
  2. Start `ent` transaction.
  3. Create `Payment` record in `pending` status.
  4. Call `PaymentProvider.Charge`.
  5. Update `Payment` status based on `Charge` outcome (`succeeded` or `failed`).
  6. Publish `payment.succeeded` or `payment.failed` event via outbox.
  7. Commit transaction.
- **GetPayment**: Look up payment by `id`. Return 404 if not found.
- **GetPayments**: Look up payments matching the `idempotency_key` query parameter. Return a JSON array of length 0 or 1.

## 6. Payment Provider Design

### Interface (`src/payment/internal/provider/provider.go`)
```go
package provider

import "context"

type ChargeRequest struct {
	Amount   float64
	Currency string
}

type ChargeResponse struct {
	Success       bool
	TransactionId string
	ErrorMessage  string
}

type PaymentProvider interface {
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error)
}
```

### MockProvider Implementation
- Injected into `ServiceContext`.
- Deterministic behavior: If `Amount` ends with `.99` (e.g., $10.99), it fails. Otherwise, it succeeds.
- Returns a generated UUID `TransactionId` on success.

## 7. Idempotency Strategy (Payment)
- Request requires `Idempotency-Key` header.
- Unique index on `idempotency_key` in `ent.Payment` schema.
- Handlers first perform `ent.Client.Payment.Query().Where(payment.IdempotencyKey(key)).Only(ctx)`. If `ent.IsNotFound(err)`, proceed. If found, return the existing payment data without charging again.
- In case of race conditions during creation, the unique index violation (`ent.IsConstraintError`) will catch it, meaning another request with the same key is processing. Return a specific concurrency error or the saved record.

## 8. Concurrency Notes (Cart)
- Since the cart items do not rely on locking inventory, the primary race condition is two concurrent `AddToCart` requests for the same `session_id` and `product_id`.
- Strategy: Use an atomic SQL "upsert" mechanism (using `ent`'s `OnConflict`) OR explicitly use a database transaction inside `AddToCart` and `UpdateCartItem` to lock the rows or rely on the `(session_id, product_id)` unique constraint to safely recover and increment.
- Example using ent upsert:
  ```go
  err := db.CartItem.Create().
      SetSessionID(sessionId).
      SetProductID(productId).
      SetQuantity(req.Quantity).
      OnConflictColumns(cartitem.FieldSessionID, cartitem.FieldProductID).
      AddQuantity(req.Quantity).
      Exec(ctx)
  ```

## 9. Wave-based Task Graph

### Wave 1: Scaffold APIs & Schemas
- Scaffold `src/cart` and `src/payment` directories and `go.mod`.
- Write `cart.api` and `payment.api`.
- Run `goctl api go -api {file} -dir .` for both.
- Define `ent/schema` and `mixin/id.go` for both.
- Run `go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema` for both.

### Wave 2: Service Context & Events
- Mirror `catalog`'s `event/event.go`, `event/dapr.go`, and `event/outbox.go` into both services.
- Define the `MockProvider` in `payment/internal/provider/mock.go`.
- Wire `config.go` and `service_context.go` to match `catalog`.
- Update `cart.go` and `payment.go` main files to initialize SQLite memory DSN (`file:ent?mode=memory&cache=shared&_fk=1`) and Dapr clients.

### Wave 3: Logic Implementation
- **Cart**: Implement `AddToCart`, `GetCart`, `UpdateCartItem`, `DeleteCartItem` parsing `X-Session-Id` header.
- **Payment**: Implement `CreatePayment` handling the `Idempotency-Key` check, transaction, `MockProvider` call, and Outbox event creation.
- **Payment**: Implement `GetPayment`.

### Wave 4: Integration Configs
- Update `Tiltfile` to add `local_resource` entries for `cart-api` and `payment-api` matching `catalog-api`.
- Update `Makefile` to include cart and payment generation & build targets.

### Wave 5: QA & Verification
- Spin up the services via Tilt/Go run.
- Execute sequence of `curl` requests testing Add to Cart and Check Cart.
- Execute sequence of `curl` requests testing Idempotent Payment Success and Failure.

## 10. Verification Commands

#### Wave 1 verification — Scaffolding APIs & Schemas
- Tool: bash
- Steps:
  1. From repo root: `(cd src/cart && goctl api go -api cart.api -dir .)`
  2. From repo root: `(cd src/cart && go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema)`
  3. From repo root: `(cd src/payment && goctl api go -api payment.api -dir .)`
  4. From repo root: `(cd src/payment && go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema)`
- Expected: All commands exit 0. `src/cart/cart.go` and `src/payment/payment.go` exist, and ent directories contain generated files.

#### Wave 2 verification — Service Context & Events
- Tool: bash
- Steps:
  1. From repo root: `(cd src/cart && go build ./...)`
  2. From repo root: `(cd src/payment && go build ./...)`
- Expected: Both commands exit 0; contexts and event dispatchers compile successfully.

#### Wave 3 verification — Logic Implementation
- Tool: bash
- Steps:
  1. From repo root: `(cd src/cart && go build -o /dev/null cart.go)`
  2. From repo root: `(cd src/payment && go build -o /dev/null payment.go)`
- Expected: Both commands exit 0 without compilation errors.

#### Wave 4 verification — Integration Configs
- Tool: bash
- Steps:
  1. From repo root: `tilt alpha tiltfile-result > /tmp/tilt-eval.txt 2>&1 || (tilt args && grep -E "cart-api|payment-api" Tiltfile)`
  2. From repo root: `grep -E "^(cart|payment)-(api|ent):" Makefile`
  3. `(cd src/cart && go build ./...)`
  4. `(cd src/payment && go build ./...)`
- Expected:
  - Step 1 executes without Starlark errors and output confirms `cart-api` and `payment-api` are defined.
  - Step 2 finds the `cart-api`, `payment-api`, `cart-ent`, and `payment-ent` targets in the Makefile.
  - Steps 3 and 4 exit 0, confirming everything compiles properly.

#### Wave 5 verification — QA & Verification
- Tool: bash
- Steps:
  1. **Cart functionality**: `curl -s -X POST http://localhost:8002/api/v1/cart/items -H "X-Session-Id: test-session-123" -H "Content-Type: application/json" -d '{"productId":"0192e21b-1234-7890-abcd-1234567890ab","quantity":2}'`
  2. **Payment Idempotency (First request)**: `curl -s -X POST http://localhost:8003/api/v1/payments/ -H "Idempotency-Key: test-key-001" -H "Content-Type: application/json" -d '{"totalAmount": 100.00, "currency": "USD", "items": []}' > resp1.json && cat resp1.json`
  3. **Payment Idempotency (Repeated request)**: `curl -s -X POST http://localhost:8003/api/v1/payments/ -H "Idempotency-Key: test-key-001" -H "Content-Type: application/json" -d '{"totalAmount": 100.00, "currency": "USD", "items": []}' > resp2.json && cat resp2.json`
  4. **Verify match**: `diff resp1.json resp2.json`
  5. **Verify count**: `curl -s "http://localhost:8003/api/v1/payments/?idempotency_key=test-key-001" > resp_list.json && cat resp_list.json`
- Expected:
  - Step 1 returns HTTP 200 with cart item JSON.
  - Step 2 and Step 3 both return HTTP 200.
  - Both JSON responses have IDENTICAL `id` and `status` values.
  - Step 4 (`diff`) exits 0 (identical responses).
  - Step 5 returns a `payments` array of length exactly 1, and that payment's `id` equals the `id` captured in step 2.

## 11. Out of Scope
- Frontend web interface updates (Next.js changes).
- Postgres/Redis setup; both stick to SQLite in-memory pattern.
- JWT authentication and user management (guest only).
- Cart-to-payment synchronous calls (client passes required payload directly to payment).
- Real payment gateway SDK integrations (Stripe, PayPal, etc.).
- Unit/integration test suites beyond smoke `curl` testing, to maintain velocity constraint.

---

## 📋 DISPLAY INSTRUCTIONS FOR OUTER AGENT

**Outer Agent: You MUST present this development plan using the following format:**

1. **Present the COMPLETE development roadmap** - Do not summarize or abbreviate sections
2. **Preserve ALL task breakdown structures** with checkboxes and formatting intact
3. **show the full risk assessment matrix** with all columns and rows
4. **Display ALL planning templates exactly as generated** - Do not merge sections
5. **Maintain all markdown formatting** including tables, checklists, and code blocks
6. **Present the complete technical specification** without condensing
7. **Show ALL quality gates and validation checklists** in full detail
8. **Display the complete library research section** with all recommendations and evaluations

**Do NOT create an executive summary or overview - present the complete development plan exactly as generated with all detail intact.**
