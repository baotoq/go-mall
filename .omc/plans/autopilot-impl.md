# Implementation Plan: `pkg/outbox` + `order` integration

## Phase 0 — Module-level prerequisites

### 0.1 go.mod changes
No new module deps needed. Verified existing required:
- `entgo.io/ent v0.14.6`
- `github.com/dapr/go-sdk v1.14.2`
- `github.com/google/uuid v1.6.0`
- `github.com/google/wire v0.6.0`
- `github.com/lib/pq v1.12.3`
- `github.com/stretchr/testify v1.11.1`
- `github.com/testcontainers/testcontainers-go v0.42.0` + `.../modules/postgres v0.42.0`
- `github.com/go-kratos/kratos/v2 v2.9.2`

Optional: `github.com/DATA-DOG/go-sqlmock v1.5.2` for unit tests. If not added, skip sqlmock-based tests and rely on testcontainers integration tests.

---

## Phase 1 — Create `pkg/outbox` (no service changes yet)

Create files in this exact order.

### 1.1 `/Users/baotoq/Work/go-mall/pkg/outbox/config.go`

- `type Config struct` with fields: PubsubName, ConsumerID, PollInterval, BatchSize, MaxRetries, RetentionPeriod, BackoffBase, BackoffMax, SweepInterval, PublishTimeout, EnableRelay.
- `func DefaultConfig() Config` — returns fully populated struct with: PubsubName="pubsub", ConsumerID from hostname, PollInterval=5s, BatchSize=100, MaxRetries=5, RetentionPeriod=24h, BackoffBase=1s, BackoffMax=5m, SweepInterval=10m, PublishTimeout=5s, EnableRelay=true.
- `func (c *Config) Validate() error` — error on: PubsubName empty, BatchSize<=0, PollInterval<=0, MaxRetries<0, BackoffBase<=0, BackoffMax<BackoffBase.

### 1.2 `/Users/baotoq/Work/go-mall/pkg/outbox/store.go`

- `const schemaSQL` with DDL for `outbox_messages` and `inbox_states` tables and indexes (exact SQL from spec).
- `func applySchema(ctx context.Context, db *sql.DB) error`
- All SQL query constants:
  - `insertOutboxSQL` — INSERT INTO outbox_messages (id, topic, payload, headers) VALUES ($1,$2,$3,$4)
  - `claimBatchSQL` — SELECT ... FOR UPDATE SKIP LOCKED
  - `markDeliveredSQL`, `markRetrySQL`, `markDeadSQL`
  - `sweepOutboxSQL`, `sweepInboxSQL`
  - `inboxClaimSQL` — INSERT ... ON CONFLICT ... RETURNING processed_at, (xmax=0) AS inserted
  - `inboxMarkDoneSQL`, `inboxRollbackSQL`

### 1.3 `/Users/baotoq/Work/go-mall/pkg/outbox/outbox.go`

- `type TxExecer interface { ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) }`
- `type publisher interface { PublishEvent(ctx, pubsubName, topic string, data any, opts ...client.PublishEventOption) error }` — private, for testability.
- `type PublishOption func(*publishOptions)` + `WithHeaders`, `WithMessageID`
- `type Client struct { db *sql.DB; pub publisher; cfg Config; log *log.Helper; stopCh/doneCh chan struct{}; stop/startOnce sync.Once }`
- `func New(db *sql.DB, daprClient client.Client, cfg Config, logger log.Logger) (*Client, error)` — apply defaults, validate, assign `pub = daprClient`.
- `func (c *Client) Migrate(ctx) error` — calls `applySchema`
- `func (c *Client) Publish(ctx, tx TxExecer, topic string, payload any, opts ...PublishOption) (string, error)`
- `func (c *Client) Start(ctx) error` — guard startOnce; if !EnableRelay return nil; launch `runRelay` + `runSweeper` goroutines
- `func (c *Client) Stop(ctx) error` — guard stopOnce; close stopCh; wait doneCh or ctx.Done

### 1.4 `/Users/baotoq/Work/go-mall/pkg/outbox/relay.go`

- `func (c *Client) runRelay(ctx)` — ticker PollInterval → `processBatch` → exit on stopCh/ctx.Done
- `func (c *Client) processBatch(ctx) error`:
  1. BeginTx
  2. Query claimBatchSQL with BatchSize; scan rows; close cursor
  3. For each row: PublishEvent with PublishTimeout; on success markDelivered; on max-retries markDead; else markRetry with backoff
  4. Commit
- `func (c *Client) runSweeper(ctx)` — ticker SweepInterval → sweep both tables
- Helper `minDuration`, package-level random source for jitter

### 1.5 `/Users/baotoq/Work/go-mall/pkg/outbox/inbox.go`

- `type Message struct { ID, Topic string; Payload json.RawMessage; Headers map[string]string; CreatedAt time.Time }`
- `type Handler func(ctx context.Context, msg Message) error`
- `func (c *Client) Subscribe(topic string, handler Handler) common.TopicEventHandler`
  - Returns `func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)`
  - Inbox claim → duplicate check → call handler → mark done or rollback

### 1.6 `/Users/baotoq/Work/go-mall/pkg/outbox/handler.go`

- `func TypedHandler[T any](fn func(ctx context.Context, evt T) error) Handler`

### 1.7 `/Users/baotoq/Work/go-mall/pkg/outbox/provider.go`

- `var ProviderSet = wire.NewSet(ProvideClient)`
- `func ProvideClient(db *sql.DB, daprClient client.Client, cfg Config, logger log.Logger) (*Client, func(), error)`
- NOTE: ProvideClient does NOT call Migrate or Start. Migration runs in NewData; Start is wired into kratos lifecycle.

**Phase 1 gate:** `go build ./pkg/outbox/...` must succeed.

---

## Phase 2 — `pkg/outbox` tests

### 2.1 `/Users/baotoq/Work/go-mall/pkg/outbox/outbox_integration_test.go`

Uses `testcontainers-go/modules/postgres`. Single `TestMain` boots one PG container. Each test truncates both tables.

Tests:
- `TestMigrate_Idempotent` — applySchema twice, no error
- `TestPublish_AndClaim` — open tx, Publish, commit; relay processBatch with fake publisher; assert processed_at NOT NULL
- `TestPublish_TransactionalAtomicity` — Publish then rollback tx; assert no row
- `TestRelay_RetryOnPublishError` — stub errors once then succeeds; assert retry_count increments, final processed_at NOT NULL
- `TestRelay_DeadLetter` — MaxRetries=2, always-error stub; assert dead state
- `TestRelay_SkipLocked_Concurrent` — two concurrent processBatch calls on 10 rows; assert no double-publish
- `TestSweeper_DeletesProcessed` — insert processed row with processed_at=NOW()-48h, sweep, assert gone
- `TestSubscribe_DedupTrueDuplicate` — same e.ID twice; handler called once
- `TestSubscribe_HandlerErrorRollsBack` — handler errors first call, succeeds second

Fake dapr stub: `type fakePub struct { mu sync.Mutex; calls []string; fn func() error }` implementing `publisher` interface.

**Phase 2 gate:** `cd pkg/outbox && go test -v ./...` passes.

---

## Phase 3 — `order` service: conf + data integration

### 3.1 Edit `/Users/baotoq/Work/go-mall/app/order/internal/conf/conf.proto`

Add `message Outbox { ... }` with fields: pubsub_name(1), consumer_id(2), poll_interval(3, Duration), batch_size(4, int32), max_retries(5, int32), retention_period(6, Duration), backoff_base(7, Duration), backoff_max(8, Duration), sweep_interval(9, Duration), publish_timeout(10, Duration), enable_relay(11, bool).

Add `Outbox outbox = 2;` to `message Data`.

Run `make config` from repo root.

### 3.2 Update `/Users/baotoq/Work/go-mall/app/order/configs/config.yaml`

```yaml
data:
  outbox:
    pubsub_name: pubsub
    poll_interval: 5s
    batch_size: 100
    max_retries: 5
    retention_period: 24h
    backoff_base: 1s
    backoff_max: 5m
    sweep_interval: 10m
    publish_timeout: 5s
    enable_relay: true
```

### 3.3 Create `/Users/baotoq/Work/go-mall/app/order/internal/data/dapr.go`

Copy from `app/catalog/internal/data/dapr.go` with package name `data`. Provides `NewDaprClient() (dapr.Client, func(), error)`.

### 3.4 Refactor `/Users/baotoq/Work/go-mall/app/order/internal/data/data.go`

New shape:
- `ProvideSQLDB(c *conf.Data) (*sql.DB, func(), error)` — opens `*sql.DB`; returns cleanup that closes it.
- `NewData(c *conf.Data, sqlDB *sql.DB, daprClient dapr.Client, outboxClient *outbox.Client, logger log.Logger) (*Data, func(), error)`:
  1. `drv := entsql.OpenDB(c.Database.Driver, sqlDB)`
  2. `client := ent.NewClient(ent.Driver(drv))`
  3. `client.Schema.Create(ctx)` (30s timeout)
  4. `outboxClient.Migrate(ctx)`
  5. Return `&Data{db: client, sqlDB: sqlDB, outbox: outboxClient}`
- `ProvideOutboxConfig(c *conf.Data) outbox.Config` — maps proto fields to Config, falls back to DefaultConfig().
- `ProviderSet = wire.NewSet(NewData, NewOrderRepo, NewDaprClient, ProvideSQLDB, ProvideOutboxConfig, outbox.ProviderSet)`
- Add accessor `(d *Data) Outbox() *outbox.Client`
- Add accessor `(d *Data) SQLDB() *sql.DB`

**Phase 3 gate:** `cd app/order && go build ./internal/...` (partial; wire not yet regenerated).

---

## Phase 4 — Wire + main.go lifecycle in `order`

### 4.1 Edit `/Users/baotoq/Work/go-mall/app/order/cmd/server/main.go`

Update `newApp` to take `*outbox.Client` and register lifecycle:
```go
func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, ob *outbox.Client) *kratos.App {
    return kratos.New(
        ...,
        kratos.BeforeStart(func(ctx context.Context) error { return ob.Start(ctx) }),
        kratos.AfterStop(func(ctx context.Context) error  { return ob.Stop(ctx) }),
    )
}
```

### 4.2 Regenerate wire

```bash
cd app/order && make wire
```

**Phase 4 gate:** `cd app/order && go build ./...` succeeds.

---

## Phase 5 — biz/data: emit `order.created` event

### 5.1 Create `/Users/baotoq/Work/go-mall/app/order/internal/biz/events.go`

```go
package biz

import "time"

const TopicOrderCreated = "order.created"

type OrderCreatedEvent struct {
    OrderID    string    `json:"order_id"`
    UserID     string    `json:"user_id"`
    SessionID  string    `json:"session_id"`
    TotalCents int64     `json:"total_cents"`
    Currency   string    `json:"currency"`
    OccurredAt time.Time `json:"occurred_at"`
}
```

### 5.2 Modify `/Users/baotoq/Work/go-mall/app/order/internal/biz/order.go`

Add to `OrderRepo` interface:
```go
CreateWithEvent(ctx context.Context, o *Order, emit func(ctx context.Context, tx TxExecer, created *Order) error) (*Order, error)
```

Add new types:
```go
type TxExecer interface {
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type OutboxPublisher interface {
    Publish(ctx context.Context, tx TxExecer, topic string, payload any) (string, error)
}
```

Update `OrderUsecase`:
- Add field `outbox OutboxPublisher`
- Update constructor: `func NewOrderUsecase(repo OrderRepo, ob OutboxPublisher) *OrderUsecase`
- Update `Create` to call `repo.CreateWithEvent` with a closure that calls `uc.outbox.Publish`

### 5.3 Implement `CreateWithEvent` in `/Users/baotoq/Work/go-mall/app/order/internal/data/order.go`

```go
func (r *orderRepo) CreateWithEvent(ctx context.Context, o *biz.Order, emit func(ctx context.Context, tx biz.TxExecer, created *biz.Order) error) (*biz.Order, error) {
    sqlTx, err := r.data.sqlDB.BeginTx(ctx, nil)
    if err != nil { return nil, err }
    defer func() { _ = sqlTx.Rollback() }()

    drv := entsql.NewDriver(r.data.db.Dialect(), entsql.Conn{ExecQuerier: sqlTx})
    txClient := ent.NewClient(ent.Driver(drv))

    // same field setters as existing Create
    result, err := txClient.Order.Create(). /* ... */ .Save(ctx)
    if err != nil { return nil, err }
    created := entToOrder(result)

    if err := emit(ctx, sqlTx, created); err != nil { return nil, err }
    if err := sqlTx.Commit(); err != nil { return nil, err }
    return created, nil
}
```

### 5.4 Add outbox adapter in `/Users/baotoq/Work/go-mall/app/order/internal/data/data.go`

```go
type outboxAdapter struct{ c *outbox.Client }
func (a *outboxAdapter) Publish(ctx context.Context, tx biz.TxExecer, topic string, payload any) (string, error) {
    return a.c.Publish(ctx, tx, topic, payload)
}
func NewOutboxPublisher(c *outbox.Client) biz.OutboxPublisher { return &outboxAdapter{c: c} }
```

Add `NewOutboxPublisher` to `ProviderSet`.

### Phase 5 gate

```bash
cd app/order && make wire && go build ./... && go vet ./...
```

---

## Phase 6 — Tests for `order` integration

### 6.1 Update `/Users/baotoq/Work/go-mall/app/order/internal/biz/order_test.go`

- Add `CreateWithEvent` to stub repo (calls `Create` then `emit` with a `fakeTx{}`)
- Add `stubOutbox` implementing `OutboxPublisher` (records calls)
- Add test `TestCreate_PublishesOrderCreatedEvent` — assert 1 call with `topic==TopicOrderCreated`
- Add test `TestCreate_NoPublishOnRepoError` — repo returns error; assert 0 outbox calls

### 6.2 Create `/Users/baotoq/Work/go-mall/app/order/internal/data/order_outbox_test.go`

Testcontainers Postgres integration test:
- `TestCreateWithEvent_AtomicallyEmitsOutbox` — real DB, real outbox; assert outbox_messages row with order id
- `TestCreateWithEvent_RollbackOnEmitError` — emit errors; assert no order row AND no outbox row

**Phase 6 gate:** `cd app/order && go test -v ./...` passes.

---

## Phase 7 — End-to-end validation

1. `go build ./...` from repo root — all services compile.
2. `go vet ./...` — no warnings.
3. `go test ./pkg/outbox/... ./app/order/...` — all pass.
4. Tilt smoke test: `make dev` → create order → verify outbox_messages row → relay drains it.

---

## Files created / modified

**New (pkg/outbox):**
- `pkg/outbox/config.go`
- `pkg/outbox/store.go`
- `pkg/outbox/outbox.go`
- `pkg/outbox/relay.go`
- `pkg/outbox/inbox.go`
- `pkg/outbox/handler.go`
- `pkg/outbox/provider.go`
- `pkg/outbox/outbox_integration_test.go`

**New (order service):**
- `app/order/internal/data/dapr.go`
- `app/order/internal/biz/events.go`
- `app/order/internal/data/order_outbox_test.go`

**Modified:**
- `app/order/internal/conf/conf.proto` (+ regenerated pb.go)
- `app/order/configs/config.yaml`
- `app/order/internal/data/data.go`
- `app/order/internal/data/order.go`
- `app/order/internal/biz/order.go`
- `app/order/cmd/server/main.go`
- `app/order/cmd/server/wire_gen.go` (regenerated)
- `app/order/internal/biz/order_test.go`
