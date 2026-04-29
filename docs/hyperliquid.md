# Hyperliquid Integration

Hyperliquid is a perp DEX on its own L1. REST API has two endpoints: `/info` (unsigned reads) and `/exchange` (signed writes). WebSocket for real-time fills/orders.

## Endpoints

| Env | REST | WebSocket |
|-----|------|-----------|
| Mainnet | `https://api.hyperliquid.xyz` | `wss://api.hyperliquid.xyz/ws` |
| Testnet | `https://api.hyperliquid-testnet.xyz` | `wss://api.hyperliquid-testnet.xyz/ws` |

## Auth — EIP-712 Signing

All trading actions are signed with an Ethereum keypair.

### Agent (API) wallet — recommended

Create a separate keypair and authorize it once via `approveAgent` from the main wallet. Agent can trade, **cannot withdraw**. Leaked bot key has no withdrawal blast radius.

Account queries (`clearinghouseState`, `openOrders`, `userFills`) must use the **master address** — the agent address returns empty.

### Signing L1 actions (orders, cancels)

EIP-712 domain: `{name:"Exchange", chainId:1337, version:"1", verifyingContract:0x0}`

Signed struct is `Agent{source, connectionId}` where:
- `source = "a"` mainnet, `"b"` testnet
- `connectionId = keccak256(msgpack(action) || nonce_be8 || vaultByte)`

**Gotcha:** msgpack key ordering and str/bin mode must be identical to the python SDK or the signature will be wrong. Run parity check (see Verification section).

### Signing user actions (approveAgent, transfers)

Domain: `HyperliquidSignTransaction`, `chainId: 0x66eee`. Action body includes `hyperliquidChain: "Mainnet"|"Testnet"`.

## Order wire shape

`POST /exchange` with action type `order`:

```json
{
  "a": 0,              // asset id int from /info {type:"meta"}; perp and spot ids use different ranges
  "b": true,           // isBuy
  "p": "95000.0",      // limit price string
  "s": "0.01",         // size string
  "r": false,          // reduceOnly
  "t": { "limit": { "tif": "Gtc" } },  // Gtc | Ioc | Alo
  "c": "0xabcd..."     // optional cloid, 16-byte hex
}
```

- Market order = `Ioc` aggressive limit (price far from mid)
- Post-only = `Alo`
- Cancel: `cancelByCloid` or `cancel` by `{a, o}` (asset + oid)

## Key `/info` queries

| Query | body `type` | Use |
|-------|------------|-----|
| Asset ids | `meta` | resolve coin name → asset id |
| Mid prices | `allMids` | sizing, entry price estimate |
| Order book | `l2Book` + `coin` | precise bid/ask depth |
| Balance / positions | `clearinghouseState` + `user` (master addr) | margin, open positions |
| Open orders | `openOrders` + `user` (master addr) | reconcile on startup |
| Recent fills | `userFills` + `user` (master addr) | fill reconciliation |

## WebSocket subscriptions

No auth required for user feeds (use master address).

- `{"type":"orderUpdates","user":"0x..."}` — real-time order state
- `{"type":"userFills","user":"0x..."}` — fills as they happen
- `{"type":"l2Book","coin":"BTC"}` — book updates (optional)

WS = fast path. REST = authoritative. Always reconcile from REST on reconnect.

## Rate limits

| Scope | Limit |
|-------|-------|
| IP | 1200 weight/min |
| Info weights | 2 (l2Book/allMids/clearinghouseState), 20 standard |
| Exchange action | 1 per call; batch `1 + floor(n/40)` |
| Address | 1 req per 1 USDC traded + 10k buffer |
| Open orders | 1000–5000 depending on tier |

Address-based limit is usually the binding constraint for an active buy bot.

## Nonce management

- Nonce = millisecond timestamp, must be strictly increasing per signer.
- HL keeps the top 100 nonces per address; accepted within `(T-2d, T+1d)`.
- `time.Now().UnixMilli()` alone breaks at >1 order/ms.
- Use a single-goroutine nonce minter **or** `atomic.Int64` with monotonic bump: `max(prev+1, now)`.

## Go SDK

Use **`github.com/sonirico/go-hyperliquid`** — actively maintained (v0.36.0, Apr 2026), covers REST + WS + EIP-712 signing via `go-ethereum/crypto`.

Wrap it behind the `biz.Exchange` interface so the signing path can be replaced without touching business logic.

Required go deps:
```
github.com/sonirico/go-hyperliquid
github.com/ethereum/go-ethereum  // pin directly even though it's transitive
```

## Proposed code structure

```
internal/
  biz/
    exchange.go          # Exchange interface, domain types (Order, Fill, Position, Balance)
    recursive_buy.go     # RecursiveBuyUsecase — strategy state machine
  data/
    hyperliquid/
      client.go          # REST client, implements biz.Exchange
      signer.go          # EIP-712 + msgpack signing (isolated for replacement)
      ws.go              # WS subscriber → channel
      wire.go            # ProviderSet
api/
  trading/v1/
    recursive_buy.proto  # Start/Stop/Status RPCs
internal/
  service/
    recursive_buy.go     # delegates to RecursiveBuyUsecase
  conf/
    conf.proto           # extend with Exchange.Hyperliquid config block
```

Config block to add to `internal/conf/conf.proto`:
```proto
message Exchange {
  message Hyperliquid {
    string api_url = 1;
    string ws_url = 2;
    bool testnet = 3;
    string account_address = 4;   // master wallet address (for /info queries)
    string agent_private_key = 5; // load from env, never commit
  }
  Hyperliquid hyperliquid = 1;
}
```

## Recursive buy design requirements

1. **Nonce** — single-goroutine minter or atomic monotonic bump (see above).
2. **Idempotency** — deterministic cloid per ladder rung: `keccak("strategy:"+strategyID+":"+stepIdx)[:16]`. Safe to retry on network errors.
3. **Partial fill tracking** — accumulate `userFills.sz` per cloid from WS; cross-check `clearinghouseState` at startup and reconnect.
4. **Reconnect** — re-fetch open orders + recent fills from REST before resuming WS.
5. **Agent rotation** — never reuse a deregistered agent address; old nonces become replayable.
6. **Strategy state** — recommend Redis (already in conf) keyed by `strategyId`.

## Testnet setup

1. Create a new Ethereum keypair for the agent wallet.
2. Fund the main wallet: `https://app.hyperliquid-testnet.xyz/drip` (1000 mock USDC / 4h).
   - **Gotcha:** faucet requires a prior mainnet deposit to unlock the address.
3. Call `approveAgent` once from the main wallet, signing the agent address.
4. Bot uses agent private key; all `/info` queries use master address.

## Verification checklist

- [ ] **Signing parity** — sign a known `order` action in both python SDK and Go, compare `action_hash` bytes hex. Must be identical.
- [ ] `GetMidPrice("BTC")` returns sane value from `/info allMids`.
- [ ] `PlaceOrder` (Alo, far from mid, tiny size) returns `oid`; visible in WS `orderUpdates`.
- [ ] `CancelOrder` succeeds; WS confirms cancellation.
- [ ] 5-rung recursive buy ladder: kill mid-flight, restart, cloid reconciliation prevents duplicates.
- [ ] Error mapping: bad nonce, duplicate cloid, insufficient margin → typed `biz` errors.
- [ ] `go test ./...` green.

## Open questions before coding

| Question | Recommendation |
|----------|---------------|
| Agent vs main wallet? | Agent (no withdrawal risk) |
| Strategy state store? | Redis (already in conf), key = strategyId |
| Secret loading? | `HL_AGENT_KEY` env var in dev; K8s Secret in deploy |

## Sources

- [API overview](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api)
- [Info endpoint](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint)
- [Exchange endpoint](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/exchange-endpoint)
- [WebSocket](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/websocket)
- [Nonces & API wallets](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/nonces-and-api-wallets)
- [Rate limits](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/rate-limits-and-user-limits)
- [Python SDK signing reference](https://github.com/hyperliquid-dex/hyperliquid-python-sdk/blob/main/hyperliquid/utils/signing.py)
- [Go SDK: sonirico/go-hyperliquid](https://github.com/sonirico/go-hyperliquid)
