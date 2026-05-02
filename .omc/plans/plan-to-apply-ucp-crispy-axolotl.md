# Plan: Apply `/ucp` skill to `app/web/`

## Context

`app/web/` is the Next.js 16.2.4 (App Router, React 19, React Compiler) storefront for the go-mall monorepo. It already renders products and a session-keyed cart against the Go `catalog` (`:8001`) and `cart` (`:8002`) services. There is no checkout flow today — `order` and `payment` exist as Go services but the web frontend never calls them.

The goal is to layer the **Universal Commerce Protocol** on top of `app/web/` so AI agents (MCP-capable clients) and external platforms can drive a standards-compliant checkout against this storefront. The web app plays the **business / merchant of record** role.

v1 scope: REST + MCP transports, no AP2, no identity linking, no PSP. v1 checkout creates a Go `order` in `PENDING` state then marks the UCP session `completed` (pending spec confirmation — see Phase 0 below). Go services stay untouched.

## Stack snapshot (verified)

| Area | Current state | File |
|---|---|---|
| Runtime | Next.js 16.2.4, App Router, `reactCompiler: true`, no edge | `app/web/next.config.ts:1` |
| Existing API client | REST calls to `:8001` / `:8002`, snake_case ↔ camelCase mappers | `app/web/src/lib/api.ts:1` |
| Domain types | `Product`, `Category`, `CartData`, `CartItemData` | `app/web/src/lib/types.ts:1` |
| Auth | NextAuth v5 beta; middleware matcher is `/cart, /cart/:path*` only | `app/web/src/proxy.ts:9` |
| Tests | Vitest unit (`*.test.tsx` next to source), Playwright e2e in `e2e/` | `app/web/vitest.config.mts`, `app/web/playwright.config.ts` |
| Lint/format | Biome 2.2 (`npm run format` mandated after every code pass) | `app/web/biome.json`, `app/web/AGENTS.md` |
| Already installed | `zod@^4.4.1`, `@tanstack/react-query@^5.100.6` (line 18), `next-auth@^5.0.0-beta.31`, `zustand@^5.0.12` | `app/web/package.json:17-18,25,31` |
| New deps needed | `mcp-handler@^1.1.0`, `@modelcontextprotocol/sdk@^1.26.0` (required peer of mcp-handler) | install in Phase 0 |
| `mcp-handler` compat | Declares `peerDeps: { next: '>=13.0.0', '@modelcontextprotocol/sdk': '1.26.0' }` — both satisfied | confirmed |

**Service ports** (source of truth: `Tiltfile:18-21`):

| Service | HTTP | gRPC | Delve |
|---|---|---|---|
| catalog | 8001 | 9001 | 7001 |
| cart | 8002 | 9002 | 7002 |
| payment | 8003 | 9003 | 7003 |
| order | 8004 | 9004 | 7004 |

**NextAuth middleware**: `/api/ucp/**` and `/api/mcp/**` are **not matched** by `proxy.ts:9` (`matcher: ["/cart", "/cart/:path*"]`). No change to `proxy.ts` is needed; assert this with a unit test.

**Constraint from `app/web/AGENTS.md`:** "This is NOT the Next.js you know" — Next 16 has breaking changes; read `node_modules/next/dist/docs/` before writing route handlers. TDD mandatory (test-first) for all code.

## Session ID strategy for cross-origin agents

UCP/MCP clients are cross-origin by design. Cookies (`SameSite=Lax`) block cross-site `POST`, so **header-based session** is used instead:

- `POST /api/ucp/checkout` returns `{ ..., "session_id": "<uuid>" }` in the JSON body.
- Subsequent requests pass `X-UCP-Session: <uuid>` header.
- If a Keycloak-authenticated NextAuth session is present (`session.user.sub`), use that as the `session_id` and skip UUID generation.
- Session TTL: 30 min (matches checkout expiry). Sessions that expire return `410 Gone`.
- MCP callers: `ucp_create_checkout` returns `session_id` in the tool response; callers pass it to subsequent tools.

## Session store strategy

In-memory `Map<string, CheckoutSession>` for dev only, guarded:

```ts
if (process.env.NODE_ENV === 'production') {
  throw new Error('UCP in-memory store cannot be used in production; set REDIS_URL');
}
```

Redis is provisioned in Tilt. v2 promotion path exists. `UCP_ENABLED=false` (default) is the kill switch and rollback path — all UCP route handlers return `503 { error: { code: "ucp_disabled" } }` when unset.

## Idempotency design

Both `POST /api/ucp/checkout` (create) and `POST /api/ucp/checkout/[id]` (complete) support `Idempotency-Key`:

- **Cache**: same in-memory map, keyed `idempotency:<key>`. Periodic eviction sweep every 60s prevents dev server memory leak.
- **Key format**: 1–64 chars, alphanumeric + `-`. Missing or empty → idempotency skipped (not an error). Malformed → `400`.
- **Body canonicalization**: Zod-parse the request, then `JSON.stringify(sortKeysDeep(parsed))` + SHA-256 hex. Ensures field-order-independent comparison (`{"a":1,"b":2}` ≡ `{"b":2,"a":1}`).
- **Match policy**: same key + same canonical hash → `200` with cached response (original status code preserved). Same key + different hash → `409 Conflict`.
- **TTL**: 24 hours. Expired keys treated as absent (fresh request).
- **HMR caveat**: keys lost on dev-server restart → duplicate Go orders possible in dev only. Documented v1 limitation; Redis in v2 resolves.
- **Concurrency**: for `complete`, status is flipped to `completing` synchronously (before any `await`) to prevent two concurrent calls both passing the status gate. Concurrent call sees `completing` and returns `409`. On Go failure, status is reverted to `ready_for_complete`.

## v1 payment model

`payment_handlers: []`. The Go `payment` service is a ledger only. In v1:

- UCP checkout reaches terminal state after creating a Go order at `:8004`.
- The UCP session status **must be verified against `.ucp-spec/`** (see Phase 0.4) — use whichever terminal state the spec permits when `payment_handlers` is empty. Expected to be `completed` but confirm before writing tests.
- Go order status: `PENDING`.
- Negative test: `POST .../action=complete → HTTP 200, ucp_session.status=<spec-confirmed>, go_order.status="PENDING"`.

## Security baseline

- **CORS**: `UCP_ALLOWED_ORIGINS` env var (comma-separated, no wildcard in v1). Credentialed requests (`Access-Control-Allow-Credentials: true`) not required (header-based session, not cookies). CSRF protection is intrinsic to header-based session — there is no cookie for an attacker to forge.
- **Request validation**: All UCP request bodies validated with Zod schemas before handler logic. Error envelope shape confirmed from spec in Phase 0.4; applied uniformly to all 400/409/502/503 responses.
- **Discovery cache**: `GET /.well-known/ucp` returns `Cache-Control: public, max-age=3600, Vary: Accept`.
- **Error on missing cart**: `getCart(cartSessionId)` returns `null` (`src/lib/api.ts:106-115`). If cart is null or empty, `createCheckout` returns `400` with spec-confirmed error envelope — not a silent $0 order.

## Session ID and cart bridge

The UCP checkout operates over two distinct session concepts:

- **Cart session ID** (`cart_session_id`): the key that identifies a cart in the Go cart service (`/v1/carts/:sessionId`). Provided by the caller in `POST /api/ucp/checkout` as `{ "cart_session_id": "<value>" }`. Used to fetch cart items via `getCart(cartSessionId)` at `src/lib/api.ts:106`.
- **UCP checkout session ID** (`ucp_id`): a new UUID generated per checkout. Returned in the response as `{ "session_id": "<uuid>" }` and passed back by callers as `X-UCP-Session`.

Anonymous checkout is allowed — matching the Go cart service's guest-session model. If a NextAuth session is present, `session.user.sub` is used as `user_id` in the Go order; otherwise `user_id` is empty string. **Verify in Phase 0**: confirm `app/order/internal/biz/order.go` accepts empty `user_id` for guest checkout (no validation that rejects it). If a sentinel value is required, substitute a constant like `"guest"`.

Session UUID uses `crypto.randomUUID()` (available in Node.js 15+ and Web Crypto — compatible with `runtime = 'nodejs'`).

Currency consistency: if `CartItemData.currency` disagrees with the UCP checkout `currency` field, `createCheckout` returns `400 { code: "currency_mismatch" }` rather than silently computing wrong totals.

## Order creation mapping

`createCheckout(complete)` posts to `ORDER_API_URL/v1/orders` (server-only env var, default `http://localhost:8004` — **no `NEXT_PUBLIC_` prefix**). Field mapping against `api/order/v1/order.proto:68-81`:

```
CreateOrderRequest {
  user_id    ← session.user.sub if authenticated, else ""
  session_id ← cart_session_id (links order back to cart)
  currency   ← ucp checkout currency
  items[]    ← for each CartItemData from getCart():
    product_id  ← CartItemData.productId         (REQUIRED — sourced from cart)
    name        ← CartItemData.name
    price_cents ← CartItemData.priceCents         (already minor units)
    image_url   ← CartItemData.imageUrl
    quantity    ← CartItemData.quantity
}
```

**Note:** `buyer.email` has no field in `CreateOrderRequest`. v1 limitation: email is dropped after checkout. Go order totals are computed server-side from items — no `total_amount` field sent. No `subtotal_cents` on `CreateOrderItem` (only on the response `OrderItem`).

Go order failure (5xx / timeout): flip UCP session status back to `ready_for_complete` (must be set to `completing` synchronously **before** the first `await` to prevent concurrent complete calls from both passing the status gate). Return HTTP `502` with a `recoverable` UCP message.

## TDD scope clarification

`app/web/AGENTS.md` mandates TDD for all code. Config files (`ucp.config.json`) are generated by the `/ucp` skill, not hand-coded — they fall outside the "write code" scope. However, a contract test that asserts the expected config shape is written **before Phase 1** and passes **after Phase 1**:

```ts
// src/lib/ucp/__tests__/config.contract.test.ts (written before /ucp init)
it('ucp.config.json has required shape', async () => {
  const config = await import('../../ucp.config.json');
  expect(config.roles).toContain('business');
  expect(config.transports).toContain('rest');
  expect(config.transports).toContain('mcp');
  expect(config.payment_handlers).toHaveLength(0);
  expect(config.features.ap2_mandates).toBe(false);
});
```

This test is red until Phase 1 completes.

## Phase 0 — Preconditions

0.1. **Install new deps**: `cd app/web && npm i mcp-handler@^1.1.0 @modelcontextprotocol/sdk@^1.26.0`

0.2. **Write config contract test** (`src/lib/ucp/__tests__/config.contract.test.ts`) — confirm it fails (no `ucp.config.json` yet).

0.3. **Run `/ucp init`** — clones `.ucp-spec/` into `app/web/.ucp-spec/`, writes `ucp.config.json`. (Phase 0.4 greps require this to complete first.) Answer init questions:
   - Role: `business`
   - Runtime: `nodejs`
   - Domain: `process.env.UCP_DOMAIN ?? 'localhost:3000'`
   - Transports: `rest` + `mcp`

0.4. **Read spec terminal state AND error envelope**:
   ```bash
   grep -A10 "payment_handlers" .ucp-spec/docs/specification/checkout.md
   grep -A10 "error" .ucp-spec/docs/specification/checkout-rest.md | head -40
   ```
   Confirm: (a) which terminal status the spec permits with `payment_handlers: []`; (b) the canonical error envelope shape (`{ error: { code, message, ... } }` or spec-defined variant). All 400/409/502/503 responses must use the spec-confirmed envelope. Record both values before writing any tests.

0.5. **Verify `.ucp-spec/` added to `.gitignore`**.

## Phase 1 — Consult and lock config (`/ucp consult`)

Walk the 12 questions. Hand-edit `ucp.config.json` to match:

```json
{
  "ucp_version": "2026-01-11",
  "roles": ["business"],
  "runtime": "nodejs",
  "capabilities": {
    "core": ["dev.ucp.shopping.checkout"],
    "extensions": [
      "dev.ucp.shopping.fulfillment",
      "dev.ucp.shopping.discount",
      "dev.ucp.shopping.order"
    ]
  },
  "transports": ["rest", "mcp"],
  "transport_priority": ["rest", "mcp"],
  "payment_handlers": [],
  "features": {
    "ap2_mandates": false,
    "identity_linking": false,
    "multi_destination_fulfillment": false
  },
  "existing_apis": {
    "cart_get":     "GET http://localhost:8002/v1/carts/:sessionId",
    "order_create": "POST http://localhost:8004/v1/orders"
  },
  "policy_urls": { "privacy": "TODO", "terms": "TODO", "refunds": "TODO", "shipping": "TODO" }
}
```

Confirm config contract test goes green after this step.

## Phase 2 — Gap analysis (`/ucp gaps`)

Pre-stated gaps (save time re-deriving):

- **GAP-001** — no `/.well-known/ucp` route.
- **GAP-002** — `src/lib/api.ts` uses error-swallowing `catch { return null }`. Build parallel UCP layer; do not retrofit.
- **GAP-003** — no `UCP-Agent` header parsing or capability negotiation.
- **GAP-004** — no checkout session state machine.
- **GAP-005** — no `ORDER_API_URL`, `UCP_DOMAIN`, `UCP_ENABLED`, `UCP_ALLOWED_ORIGINS` env vars.

## Phase 3 — TDD implementation loop

Follow red→green for each module. After each sub-phase: `npm run format && npm run test`.

### 3a — Core types, schemas, store

| Test file | Tests |
|---|---|
| `src/lib/ucp/schemas/__tests__/checkout.test.ts` | valid `CreateCheckout` (with `cart_session_id`) passes; missing `cart_session_id` rejects; malformed `Idempotency-Key` (>64 chars) rejects; missing `Idempotency-Key` is valid (idempotency optional) |
| `src/lib/ucp/__tests__/store.test.ts` | store throws in prod env; get/set works in test env; expired sessions return null |

Implement: `src/lib/ucp/types/checkout.ts`, `src/lib/ucp/schemas/checkout.ts`, `src/lib/ucp/store.ts`

### 3b — Capability negotiation + response helpers

| Test file | Tests |
|---|---|
| `src/lib/ucp/__tests__/negotiation.test.ts` | `parseUCPAgent` parses profile header; returns null on missing; `negotiateCapabilities` computes intersection; fallback on fetch error |
| `src/lib/ucp/__tests__/response.test.ts` | `wrapResponse` includes `ucp.capabilities`; `errorResponse` has correct shape |

Implement: `src/lib/ucp/negotiation.ts`, `src/lib/ucp/response.ts`

### 3c — Discovery profile

| Test file | Tests |
|---|---|
| `src/lib/ucp/__tests__/profile.test.ts` | `generateProfile()` returns all configured capabilities; service endpoints use `UCP_DOMAIN` |
| `src/app/.well-known/ucp/__tests__/route.test.ts` | GET returns 200 JSON; `Cache-Control: public, max-age=3600` header present |

Implement: `src/lib/ucp/profile.ts`, `src/app/.well-known/ucp/route.ts` (`export const runtime = 'nodejs'`)

### 3d — Checkout handler

| Test file | Tests |
|---|---|
| `src/lib/ucp/handlers/__tests__/checkout.test.ts` | `createCheckout` with null cart returns 400; maps `CartItemData` fields correctly to `LineItem`; initial status is `incomplete`; PATCH with `buyer.email` → `ready_for_complete`; complete creates Go order, UCP session reaches spec-confirmed terminal state, Go order is `PENDING`; idempotent complete with same key returns cached result; concurrent complete with different key returns 409; Go order failure → 502, session stays `ready_for_complete` |

Implement: `src/lib/ucp/handlers/checkout.ts` (delegates to `getCart` at `src/lib/api.ts:106`), stub handlers `src/lib/ucp/handlers/fulfillment.ts` (flat-rate: $5.99 standard / $12.99 express), `src/lib/ucp/handlers/discount.ts` (`valid: false, reason: "no_codes_configured"`), `src/lib/ucp/handlers/order.ts` (POST to `ORDER_API_URL`)

### 3e — Route handlers

| Test file | Tests |
|---|---|
| `src/app/api/ucp/checkout/__tests__/route.test.ts` | POST creates session, returns `session_id` in body; `Idempotency-Key` dedup works |
| `src/app/api/ucp/checkout/[id]/__tests__/route.test.ts` | GET returns session; PATCH updates; POST complete transitions state; 404 on unknown id |
| `src/app/__tests__/proxy-exclusion.test.ts` | `/api/ucp/test` and `/api/mcp/test` do NOT match `proxy.ts` matcher |

Implement: `src/app/api/ucp/checkout/route.ts`, `src/app/api/ucp/checkout/[id]/route.ts` (both `runtime = 'nodejs'`)

### 3f — Stub handlers

Implement (covered by checkout tests from 3d):
- `src/lib/ucp/handlers/fulfillment.ts` — flat-rate: $5.99 standard / $12.99 express
- `src/lib/ucp/handlers/discount.ts` — `{ valid: false, reason: "no_codes_configured" }`
- `src/lib/ucp/handlers/order.ts` — POST to `process.env.ORDER_API_URL ?? 'http://localhost:8004'`

### 3g — MCP transport

| Test file | Tests |
|---|---|
| `src/app/api/mcp/__tests__/tool-registration.test.ts` | `tools/list` response contains `ucp_get_profile`, `ucp_create_checkout`, `ucp_get_checkout`, `ucp_update_checkout`, `ucp_complete_checkout` |
| `src/app/api/mcp/__tests__/tool-invocation.test.ts` | `ucp_get_profile` returns capabilities array; `ucp_create_checkout` with valid `cart_session_id` returns `session_id`; `ucp_get_checkout` returns session; `ucp_complete_checkout` transitions state; unknown tool name returns JSON-RPC error |
| `src/app/api/mcp/__tests__/session-bridging.test.ts` | `session_id` from `ucp_create_checkout` accepted by `ucp_get_checkout`; missing session returns 404-equivalent; expired session returns 410-equivalent |

Implement: `src/app/api/mcp/[transport]/route.ts` using `mcp-handler@^1.1.0` (`runtime = 'nodejs'`), `src/lib/ucp/transports/mcp-tools.ts`

## Phase 4 — Project integration

1. **Env** — add to `app/web/.env.example`:
   ```
   UCP_ENABLED=false
   UCP_DOMAIN=localhost:3000
   ORDER_API_URL=http://localhost:8004
   UCP_ALLOWED_ORIGINS=http://localhost:3000
   ```
2. **Config helper** — `src/lib/ucp/config.ts` (typed re-export of `ucp.config.json` via `resolveJsonModule`).
3. **`proxy.ts`** — **no change needed**; confirmed matcher at `:9` does not cover `/api/ucp/*` or `/api/mcp/*`. Proxy exclusion test in 3e asserts this.
4. **Docs** — append UCP section to `app/web/CLAUDE.md` (not `AGENTS.md`). Update root `CLAUDE.md` Services table and Architecture section to note web→order cross-service call.

## Phase 5 — E2E tests (Playwright)

After all unit tests pass, with `make dev` + `UCP_ENABLED=true npm run dev`:

- `e2e/ucp-discovery.spec.ts`:
  ```ts
  const res = await request.get('/.well-known/ucp');
  expect(res.status()).toBe(200);
  const body = await res.json();
  expect(body.ucp.capabilities.map(c => c.name)).toContain('dev.ucp.shopping.checkout');
  ```
- `e2e/ucp-checkout-rest.spec.ts` — create (get `session_id`) → PATCH buyer.email → POST complete → assert `status=<terminal>` and `:8004/v1/orders` has new order.
- `e2e/ucp-mcp.spec.ts` — JSON-RPC `tools/call ucp_get_profile`, `ucp_create_checkout`, `ucp_get_checkout`.

Run: `npm run test:e2e`.

## Phase 6 — Validate

1. `/ucp validate` against `.ucp-spec/schemas/shopping/checkout.json` and `spec/discovery/profile_schema.json`.
2. `/ucp profile` — eyeball capabilities array.

## Files to create / modify

### Create (16 files)

| File | Module | Test file |
|---|---|---|
| `src/lib/ucp/config.ts` | Config re-export | — |
| `src/lib/ucp/types/checkout.ts` | Types | — |
| `src/lib/ucp/schemas/checkout.ts` | Zod schemas | `schemas/__tests__/checkout.test.ts` |
| `src/lib/ucp/store.ts` | In-memory store | `__tests__/store.test.ts` |
| `src/lib/ucp/profile.ts` | `generateProfile()` | `__tests__/profile.test.ts` |
| `src/lib/ucp/negotiation.ts` | UCP-Agent header + capability intersection | `__tests__/negotiation.test.ts` |
| `src/lib/ucp/response.ts` | Response helpers | `__tests__/response.test.ts` |
| `src/lib/ucp/handlers/checkout.ts` | State machine + cart bridge | `handlers/__tests__/checkout.test.ts` |
| `src/lib/ucp/handlers/fulfillment.ts` | Flat-rate stub | (covered by checkout tests) |
| `src/lib/ucp/handlers/discount.ts` | No-op stub | (covered by checkout tests) |
| `src/lib/ucp/handlers/order.ts` | POST to `:8004` | (covered by checkout tests) |
| `src/lib/ucp/transports/mcp-tools.ts` | MCP tool schemas | — |
| `src/app/.well-known/ucp/route.ts` | Discovery endpoint | `.well-known/ucp/__tests__/route.test.ts` |
| `src/app/api/ucp/checkout/route.ts` | POST create | `checkout/__tests__/route.test.ts` |
| `src/app/api/ucp/checkout/[id]/route.ts` | GET / PATCH / POST complete | `checkout/[id]/__tests__/route.test.ts` |
| `src/app/api/mcp/[transport]/route.ts` | MCP JSON-RPC | `mcp/__tests__/route.test.ts` |

Plus 8 test files (collocated per project convention).

### Modify (4 files)

| File | Change |
|---|---|
| `app/web/.env.example` | Add `UCP_ENABLED`, `UCP_DOMAIN`, `ORDER_API_URL`, `UCP_ALLOWED_ORIGINS` |
| `app/web/.gitignore` | Add `.ucp-spec/` (verify skill does this) |
| `app/web/CLAUDE.md` | Append UCP section |
| root `CLAUDE.md` | Note web→order cross-service call |

**Do not touch**: `src/lib/api.ts`, `src/lib/types.ts`, `src/proxy.ts`, existing pages.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Next 16 Route Handler API differs from skill templates | Read `node_modules/next/dist/docs/` per AGENTS.md before writing; adjust imports |
| v1 `completed` state disallowed by spec when `payment_handlers: []` | Phase 0.4 reads spec section before writing any tests; use spec-confirmed state |
| In-memory store lost on HMR / dev restart | Documented v1 limitation; prod guard throws; `UCP_ENABLED=false` default; Redis path in v2 |
| Duplicate Go orders on HMR (idempotency cache lost) | Dev-only limitation; documented; Redis in v2 resolves |
| Web→order convention break | Root `CLAUDE.md` updated in same PR; web treated as BFF orchestrator |
| Cart `null` silently producing $0 order | `createCheckout` returns `400 empty_cart` if `getCart` returns null or empty |
| Go order `:8004` unreachable | `order.ts` returns `502`; session stays `ready_for_complete`; test covers this |
| Cross-origin agent session loss (cookie blocked) | Header-based session (`X-UCP-Session`) avoids cookie; `session_id` in JSON response |
| `proxy.ts` accidentally modified to guard `/api/ucp` | Proxy exclusion test (`src/app/__tests__/proxy-exclusion.test.ts`) fails if matcher changes |

## Verification

With `make dev` and `UCP_ENABLED=true npm run dev`:

```bash
# 1. Discovery
curl -s http://localhost:3000/.well-known/ucp | jq '.ucp.capabilities[].name'
# → "dev.ucp.shopping.checkout", "dev.ucp.shopping.fulfillment", ...

# 2. Create checkout (requires a cart with items at session "test-session")
curl -s -X POST http://localhost:3000/api/ucp/checkout \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test-session"}' | jq '.status, .session_id'
# → "incomplete", "<uuid>"

# 3. Add buyer (required to reach ready_for_complete)
curl -s -X PATCH http://localhost:3000/api/ucp/checkout/<id> \
  -H "Content-Type: application/json" \
  -H "X-UCP-Session: <uuid>" \
  -d '{"buyer":{"email":"test@example.com"}}' | jq '.status'
# → "ready_for_complete"

# 4. Complete
curl -s -X POST "http://localhost:3000/api/ucp/checkout/<id>?action=complete" \
  -H "Content-Type: application/json" \
  -H "X-UCP-Session: <uuid>" \
  -d '{}' | jq '.status'
# → "<spec-confirmed terminal state>"

# 5. Verify Go order in PENDING
curl -s http://localhost:8004/v1/orders/<order-id> | jq '.status'
# → "PENDING"
```

Additional:
- `npm run format && npm run lint && npm run test` — all green.
- `npm run test:e2e` — all 3 Playwright specs green.
- `/ucp validate` reports PASS.

## Out of scope (explicitly deferred)

- A2A and embedded transports.
- AP2 mandate signing, OAuth identity linking.
- Multi-destination fulfillment.
- Real PSP integration and Go order status advancement to PAID.
- Persistent checkout session store (Redis/Postgres) — v2.
- Rate limiting / abuse prevention.
- Observability / distributed tracing.
- Modifying `src/lib/api.ts` to a typed SDK.
- Vercel deployment artefacts.
- Structured logging for the UCP layer.
