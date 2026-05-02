@AGENTS.md

## UCP (Universal Commerce Protocol) Integration

`app/web/` implements UCP v2026-01-11 as a **business/merchant** node. It exposes a standards-compliant checkout flow over REST and MCP transports.

### Config
`ucp.config.json` at project root. Kill switch: `UCP_ENABLED=false` (default) — all UCP handlers return 503 when unset.

### Key env vars
| Var | Default | Purpose |
|-----|---------|---------|
| `UCP_ENABLED` | `false` | Feature kill switch |
| `UCP_DOMAIN` | `localhost:3000` | Domain for `/.well-known/ucp` |
| `ORDER_API_URL` | `http://localhost:8004` | Go order service (server-only, no `NEXT_PUBLIC_`) |
| `UCP_ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins |

### Routes
| Route | Description |
|-------|-------------|
| `GET /.well-known/ucp` | UCP discovery profile |
| `POST /api/ucp/checkout` | Create checkout session; returns `session_id` in body |
| `GET /api/ucp/checkout/[id]` | Get session |
| `PATCH /api/ucp/checkout/[id]` | Update buyer info |
| `POST /api/ucp/checkout/[id]?action=complete` | Complete checkout → creates Go order |
| `GET/POST /api/mcp/[transport]` | MCP JSON-RPC endpoint |

### Session design
- Sessions identified by UUID returned as `session_id` in JSON body (cross-origin safe, no cookies).
- Callers pass `X-UCP-Session: <uuid>` on subsequent requests.
- Anonymous checkout: `user_id = "guest"` (Go order service requires non-empty user_id).
- TTL: 30 min. In-memory store only — **not production-safe**. Set `UCP_ENABLED=false` in prod.

### Source layout
```
src/lib/ucp/
  config.ts              ← typed ucp.config.json re-export
  types/checkout.ts      ← CheckoutSession, CheckoutStatus, etc.
  schemas/checkout.ts    ← Zod validation schemas
  store.ts               ← in-memory session + idempotency store
  profile.ts             ← generateProfile() for /.well-known/ucp
  negotiation.ts         ← parseUCPAgent, negotiateCapabilities
  response.ts            ← wrapResponse, errorResponse helpers
  handlers/
    checkout.ts          ← state machine (create, update, complete)
    order.ts             ← POST to ORDER_API_URL/v1/orders
    fulfillment.ts       ← flat-rate stub ($5.99/$12.99)
    discount.ts          ← no-op stub
  transports/
    mcp-tools.ts         ← MCP tool definitions
src/app/
  .well-known/ucp/route.ts
  api/ucp/checkout/route.ts
  api/ucp/checkout/[id]/route.ts
  api/mcp/[transport]/route.ts
```
