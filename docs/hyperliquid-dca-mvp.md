# Hyperliquid Spot DCA Bot — MVP Design

## Goal

Automate recurring spot buys on Hyperliquid at a configured cadence using a ladder of limit orders (recursive-buy / DCA). The bot runs unattended, survives restarts without duplicating orders, and exposes a status endpoint for observability.

**Out of scope (MVP):** UI, alerting, sells/take-profit, multi-exchange, backtesting runner, on-chain signal multipliers (designed in `dca-strategy.md`, wired in a later phase).

---

## Architecture

```
configs/config.yaml
        │
        ▼
[Scheduler Server]  ──tick──▶  [RecursiveBuyUsecase]
  (time.Ticker)                        │           │
                                       ▼           ▼
                              [biz.Exchange]  [StrategyStateRepo]
                                    │                │
                              [data/hyperliquid]  [data/redis]
                                    │
                         Hyperliquid REST /exchange
```

Fits inside the existing `go-kratos v2` layout:

| Layer | Package | Role |
|-------|---------|------|
| **server** | `internal/server/scheduler.go` | Kratos `transport.Server`; owns ticker goroutine per strategy |
| **biz** | `internal/biz/recursive_buy.go` | Ladder math, nonce, cloid, order dispatch |
| **biz interfaces** | `internal/biz/exchange.go` | `Exchange`, `StrategyStateRepo` contracts |
| **data** | `internal/data/hyperliquid/` | Implements `Exchange` via go-hyperliquid SDK |
| **data** | `internal/data/strategy_state.go` | Implements `StrategyStateRepo` in Redis |
| **service** | `internal/service/strategy.go` | gRPC `StrategyService.Status` |
| **api** | `api/trading/v1/strategy.proto` | Proto definition |

---

## Exchange Interface (`biz.Exchange`)

```go
type Exchange interface {
    MidPx(ctx context.Context, asset Asset) (decimal.Decimal, error)
    SpotBalance(ctx context.Context, asset Asset) (decimal.Decimal, error)
    PlaceSpotBuy(ctx context.Context, req PlaceOrderReq) (OrderRef, error)
    CancelByCloid(ctx context.Context, asset Asset, cloid Cloid) error
    OpenOrders(ctx context.Context, asset Asset) ([]OrderRef, error)
}
```

Domain types defined in `internal/biz/exchange.go`. All errors mapped to kratos typed errors (`errors.BadRequest`, `errors.NotFound`).

---

## DCA Ladder Math

Each tick:

1. Fetch `midPx` from `/info allMids`.
2. Compute `N` rung prices from `price_offsets_bps` config (negative = below mid):
   ```
   rungPx[i] = midPx * (1 + offsets_bps[i] / 10000)
   ```
3. Compute rung size: `rungSize = quoteAmount / N / rungPx[i]` (rounded to asset step).
4. Derive deterministic cloid per rung:
   ```
   cloid[i] = keccak256(strategyID + ":" + runID + ":" + i)[:16]
   ```
5. Call `Exchange.PlaceSpotBuy` for each rung (post-only `Alo` TIF).
6. Persist `DCARun{id, ts, rungs, cloids, status}` to Redis.

Example config producing a 4-rung ladder 0.5–2% below mid at weekly cadence:

```yaml
strategy:
  asset: BTC
  quote_amount: "100"      # USDC per tick
  interval: 168h           # weekly
  ladder_size: 4
  price_offsets_bps: ["-50", "-100", "-150", "-200"]
  max_slippage_bps: "300"
```

---

## Nonce Management

Single `atomic.Int64` per signer. Each call to minter:

```go
func (s *Signer) NextNonce() int64 {
    now := time.Now().UnixMilli()
    for {
        prev := s.nonce.Load()
        next := max(prev+1, now)
        if s.nonce.CompareAndSwap(prev, next) {
            return next
        }
    }
}
```

Guarantees strictly increasing nonces even at >1 order/ms.

---

## Cloid Idempotency

Same `(strategyID, runID, rungIndex)` → same cloid across restarts. On startup, `RecursiveBuyUsecase` loads the last `DCARun` from Redis and reconciles open orders via `Exchange.OpenOrders`. Rungs with matching cloids already open are skipped, not re-placed.

---

## Auth & Secrets

| Key | Where | Notes |
|-----|-------|-------|
| `HL_AGENT_KEY` | env var | Agent private key (hex, no `0x`). Never commit. |
| `master_address` | `configs/config.yaml` | Master wallet address. Only used for `/info` reads. |

Agent wallet can trade, **cannot withdraw** — leaked key has bounded blast radius.

Config block in `internal/conf/conf.proto`:

```proto
message Exchange {
  message Hyperliquid {
    string api_url        = 1;
    string master_address = 2;
    string agent_key_env  = 3;  // env var name, default "HL_AGENT_KEY"
    bool   testnet        = 4;
  }
  Hyperliquid hyperliquid = 1;
}
```

---

## State (Redis)

Key schema:

```
strategy:{strategyID}:run:{runID}  →  JSON(DCARun), TTL 30d
strategy:{strategyID}:latest       →  runID string
```

`DCARun` structure:

```go
type DCARun struct {
    ID       string
    StrategyID string
    Ts       time.Time
    Rungs    []Rung    // {Cloid, Asset, Px, Sz, Status, OrderRef}
    Status   RunStatus // pending | placed | partial | filled | cancelled
}
```

---

## Scheduler Server

Implements `kratos transport.Server`:

```go
func (s *Scheduler) Start(ctx context.Context) error {
    for _, strategy := range s.strategies {
        go s.run(ctx, strategy)
    }
    return nil
}

func (s *Scheduler) run(ctx context.Context, st DCAStrategy) {
    ticker := time.NewTicker(st.Interval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            s.usecase.Tick(ctx, st)
        case <-ctx.Done():
            return
        }
    }
}
```

Registered in `newApp` alongside HTTP and gRPC servers.

---

## API: StrategyService

```proto
service StrategyService {
  rpc Status(StatusRequest) returns (StatusReply) {}
}

message StatusRequest { string strategy_id = 1; }
message StatusReply {
  string strategy_id = 1;
  repeated RunSummary runs = 2;  // last 10
}
```

gRPC only for MVP. No Start/Stop — strategy config is static from `config.yaml`.

---

## New Dependencies

| Module | Purpose |
|--------|---------|
| `github.com/sonirico/go-hyperliquid` | REST + EIP-712 signing |
| `github.com/redis/go-redis/v9` | Strategy state |
| `github.com/shopspring/decimal` | Price/size math (no float) |

(`time.Ticker` for scheduler — no cron library needed for MVP.)

---

## Build Sequence

1. Add deps (`go get`).
2. Extend `internal/conf/conf.proto` → `make config`.
3. Define `biz.Exchange` interface + domain types.
4. Implement `data/hyperliquid/{client,exchange}.go`; signing-parity test green.
5. Implement `data/strategy_state.go` (Redis).
6. Implement `biz.RecursiveBuyUsecase` + unit tests.
7. Implement `internal/server/scheduler.go`; wire.
8. Add `api/trading/v1/strategy.proto` + service → `make api`.
9. `wire` regen → `make all` → testnet smoke test.

---

## Verification

### Unit tests

- `biz/recursive_buy_test.go` — ladder rung computation, deterministic cloid.
- `data/hyperliquid/nonce_test.go` — monotonic nonce under concurrent goroutines.

### Signing parity (block mainnet until green)

Build a known order action in Go, compare `action_hash` hex to the python SDK output for the same input. Fixture lives in `data/hyperliquid/sign_test.go`.

References: [`hyperliquid-python-sdk/hyperliquid/utils/signing.py`](https://github.com/hyperliquid-dex/hyperliquid-python-sdk/blob/main/hyperliquid/utils/signing.py)

### Testnet integration

```bash
export HL_AGENT_KEY=<testnet-agent-hex>
# configs/config.yaml: testnet: true, tiny quote_amount
./bin/tradingbot -conf ./configs
```

Verify:
- Orders appear in testnet UI (`app.hyperliquid-testnet.xyz`).
- Redis keys `strategy:*:run:*` populated.
- Restart → same cloids → no duplicate orders.
- `grpcurl -plaintext :9000 trading.v1.StrategyService/Status` returns runs.

---

## Related docs

- [`docs/hyperliquid.md`](hyperliquid.md) — exchange API, signing, rate limits, SDK details.
- [`docs/dca-strategy.md`](dca-strategy.md) — backtest evidence, sentiment/on-chain multipliers (Phase 2).
