# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

@AGENTS.md

> **AGENTS.md note:** Next.js 16 in this repo has breaking changes from older versions. Read `node_modules/next/dist/docs/` (or use `context7`) before relying on training-data knowledge.

## Commands

```bash
npm run dev         # next dev (port 3000)
npm run build       # next build
npm run lint        # biome check
npm run format      # biome format --write — run after editing
npm test            # vitest run (jsdom; co-located `__tests__/` + `*.test.ts(x)`)
npm run test:watch  # vitest watch
npm run test:e2e    # playwright (auto-starts dev server with MOCK_ORDER_SERVICE=true)
```

Run a single vitest file or test name:
```bash
npx vitest run src/lib/api.test.ts
npx vitest run -t "addCartItem"
```

Run a single Playwright spec:
```bash
npx playwright test e2e/checkout-flow.spec.ts
npx playwright test --grep "completes checkout"
```

## Architecture

**App Router + React 19 + React Compiler.** `next.config.ts` enables `reactCompiler: true`, so avoid manual `useMemo`/`useCallback` for typical cases — let the compiler handle memoization. Path alias: `@/*` → `src/*`.

### Auth flow (NextAuth v5 beta + Keycloak)

- `src/auth.config.ts` — edge-safe config (no providers, just `authorized` callback). Used by `src/proxy.ts` (Next.js middleware) which currently gates only `/cart` routes.
- `src/auth.ts` — full config with the `Credentials` provider that exchanges email/password for a Keycloak token via `src/lib/keycloak.ts`. JWT callback handles refresh on expiry; `error: "RefreshTokenError"` propagates to the session when refresh fails.
- The middleware file is **`src/proxy.ts`**, not `middleware.ts` — this is a Next.js 16 convention shift.
- Server components use `await auth()` (see `app/layout.tsx`); routes/components access `session.access_token` (typed via module augmentation in `auth.ts`).

### Data layer (`src/lib/`)

- `api.ts` — thin `fetch` wrappers over Go services. Reads `NEXT_PUBLIC_API_URL` (catalog, default `:8001`) and `NEXT_PUBLIC_CART_API_URL` (cart, default `:8002`). All functions **swallow errors and return `null`/`[]`** rather than throwing — callers must handle the empty case (no try/catch needed at call sites).
- `keycloak.ts` — Keycloak token endpoint client (used only by `auth.ts`).
- `types.ts` — shared `Product`, `Category`, `CartData`, `CartItemData` shapes (camelCase TS, snake_case wire).
- Field mapping happens in `mapProduct`/`mapCart`/`mapCartItem`. When the Go API adds fields, update the mapper — raw `unknown` payloads come in, normalized objects go out.

### Routing layout

- `src/app/(auth)/signin`, `(auth)/signup` — auth pages (route group, no URL prefix).
- `src/app/products`, `cart`, `checkout`, `checkout/success` — user pages.
- `src/app/api/auth/[...nextauth]` — NextAuth handlers.
- `src/app/api/ucp/...` and `api/mcp/[transport]` — UCP REST + MCP transports.
- `src/app/checkout/checkout-client.tsx` is the client form; `page.tsx` is the server entry that hydrates it. `sections/` holds form steps.

### State

Zustand store in `src/store/cart.ts` (with sibling `cart.test.ts`) — guest-cart session id and item state. Server cart truth lives in the Go cart service; the store mirrors locally for UX.

## UCP (Universal Commerce Protocol)

`app/web/` is a UCP v2026-01-11 **business node**. UCP routes are always enabled.

### Env vars

| Var | Default | Purpose |
|-----|---------|---------|
| `UCP_DOMAIN` | `localhost:3000` | Domain in `/.well-known/ucp` |
| `ORDER_API_URL` | `http://localhost:8004` | Go order service (server-only — **never** prefix with `NEXT_PUBLIC_`) |
| `UCP_ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowlist |
| `MOCK_ORDER_SERVICE` | `false` | When `true`, `createOrder` returns `{ id: "mock_<first-8>" }` instead of calling `ORDER_API_URL` |

For local checkout without the Go order service: set `MOCK_ORDER_SERVICE=true` in `.env.local`. Playwright already sets this in `playwright.config.ts`. **Never set `MOCK_ORDER_SERVICE` in production.**

### Routes

| Route | Description |
|-------|-------------|
| `GET /.well-known/ucp` | Discovery profile (built from `ucp.config.json`) |
| `POST /api/ucp/checkout` | Create session → returns `session_id` in JSON body |
| `GET /api/ucp/checkout/[id]` | Fetch session |
| `PATCH /api/ucp/checkout/[id]` | Update buyer info |
| `POST /api/ucp/checkout/[id]?action=complete` | Complete → calls Go order service |
| `GET/POST /api/mcp/[transport]` | MCP JSON-RPC |

### Session model

- Sessions identified by UUID returned in JSON body as `session_id` (cross-origin safe — no cookies).
- Callers must pass `X-UCP-Session: <uuid>` on subsequent requests.
- Anonymous: `user_id = "guest"` (Go order service rejects empty `user_id`).
- 30-min TTL, **in-memory store only** — not production-safe.

### Source layout

```
src/lib/ucp/
  config.ts              typed ucp.config.json re-export
  types/checkout.ts      CheckoutSession, CheckoutStatus
  schemas/checkout.ts    Zod schemas for handler input validation
  store.ts               in-memory session + idempotency store
  profile.ts             /.well-known/ucp generator
  negotiation.ts         parseUCPAgent, negotiateCapabilities
  response.ts            wrapResponse, errorResponse
  cors.ts                CORS helpers honoring UCP_ALLOWED_ORIGINS
  handlers/{checkout,order,fulfillment,discount}.ts
  transports/mcp-tools.ts
```

## Testing & Verification

For UI/feature work, validate at **two layers** before marking done:

1. **Playwright** (`npm run test:e2e`) — regression coverage; specs live in `e2e/`. Required.
2. **agent-browser skill** — exploratory walk-through. Use the `agent-browser` Skill tool to navigate, fill forms, click, screenshot. Save screenshots to `.qa/screenshots/<feature>`.

The two are complementary: Playwright catches regressions; agent-browser surfaces visual glitches, color contrast, real network behavior, and console warnings that specs miss. Don't claim a UI feature is shipped on Playwright alone.

**TDD discipline:** Write tests first, confirm they fail for the right reason, then implement the minimal fix and re-run. Test behavior, not implementation — no exhaustive mocks, no re-asserting framework behavior.

## Workflow

- Run `npm run format` after editing — Biome enforces 2-space indent, organizes imports, applies Next.js + React lint domains.
- TS strict mode is on with `noEmit: true`; type-check with `npx tsc --noEmit`.
- Treat `api.ts`'s null-returns as the contract: callers must handle missing data; do not "fix" by adding throws.
