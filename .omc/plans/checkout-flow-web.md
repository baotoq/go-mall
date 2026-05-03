# Plan: Checkout Flow for `app/web/`

**Date:** 2026-05-03
**Scope:** Add a single-page checkout UI at `/checkout` that extends the existing UCP checkout backend. Order creation is mocked (no live Go order service required).
**Mode:** Interview-driven plan (4 design decisions captured below).

---

## Requirements Summary

A buyer who has items in their cart and is signed in can:

1. Click "Checkout" on `/cart` and land on `/checkout`.
2. See a sticky right-side order summary (items, subtotal, shipping, total).
3. Fill three sections on the left: **Contact** (email, name, phone) → **Shipping address** (line1/2, city, state, postal code, country) → **Payment (mock)** (card number, exp, cvc).
4. Submit "Place order", see a brief processing state, and land on `/checkout/success?id=<session_id>` showing order id, item recap, shipping address, and total.
5. After success, the cart is cleared.
6. Unauthenticated visitors are redirected to `/signin?callbackUrl=/checkout`.
7. Empty-cart visitors are redirected to `/cart`.

The checkout uses the existing UCP state machine end-to-end (`incomplete → ready_for_complete → complete_in_progress → completed`); the only change to UCP is **(a)** an extended buyer/shipping/payment schema and **(b)** an env-gated short-circuit of `createOrder` so the Go order service is not required in dev.

---

## Design Decisions (captured from interview)

| # | Decision | Choice | Source |
|---|----------|--------|--------|
| 1 | Flow shape | **Single-page checkout** (one `/checkout` + sticky summary) | user pick |
| 2 | Backend wiring | **Extend existing UCP route** (`/api/ucp/checkout` + `/[id]`); stub `createOrder` via env | user pick |
| 3 | Field scope | **Contact + shipping + mock payment** (card stored as last4 + brand only) | user pick |
| 4 | Auth gate | **Match `/cart`** — server-redirect to `/signin?callbackUrl=/checkout` | user pick |

---

## Acceptance Criteria (testable)

### Routing & guards

- [ ] `GET /checkout` without a NextAuth session returns a redirect to `/signin?callbackUrl=%2Fcheckout` (assert in `checkout/__tests__/page.test.tsx`).
- [ ] `GET /checkout` while signed in but with a server-side cart that has zero items redirects to `/cart` (server fetch via `getCart(sessionId)` is not feasible because session id lives in localStorage; instead the **client** runs an effect: if `useCartStore().items.length === 0 && !isLoading`, push `/cart`).
- [ ] Cart page button "Checkout" navigates to `/checkout` — replace the current `<Button>` at `src/app/cart/cart-client.tsx:163-165` with `<Link href="/checkout" className={cn(buttonVariants({ size: 'lg' }), 'w-full')}>`.
- [ ] `GET /checkout/success?id=<missing>` renders an empty-state with a "Continue shopping" link.

### Form behavior

- [ ] Each required field shows a red Zod error message under it on submit when empty (use `react-hook-form` + `@hookform/resolvers/zod` + shadcn `form` primitive).
- [ ] Email field pre-fills from `session.user?.email` on first render (server passes this prop).
- [ ] Phone is optional; all other fields are required.
- [ ] Card number accepts only digits and spaces, max 19 chars; exp accepts `MM/YY`; cvc 3-4 digits. **No Luhn check** — keep validation light per `AGENTS.md`.
- [ ] On submit while pending, the "Place order" button shows a spinner and is disabled.
- [ ] Submit attaches `Idempotency-Key: <crypto.randomUUID()>` so the second click of a double-clicked button is a no-op (idempotency cache hit).

### UCP wiring

- [ ] `POST /api/ucp/checkout` with the new body shape (containing `buyer`, `shipping_address`, `payment`) returns 201 with a `session_id`. Existing email-only callers (the MCP flow) keep working — all new fields are **optional at the Zod schema level**.
- [ ] `PATCH /api/ucp/checkout/[id]` accepts updates to `buyer`, `shipping_address`, and `payment`. After patch, `status` is `ready_for_complete` iff `buyer.email` AND `shipping_address` AND `payment` are all present; otherwise `incomplete` with `messages[]` describing what is missing.
- [ ] `POST /api/ucp/checkout/[id]?action=complete` returns 200 with the `session.status === 'completed'` when `MOCK_ORDER_SERVICE=true` (no network call to `:8004`). Without the env, behavior is unchanged.
- [ ] Payment fields are **never echoed back in the session response**. The server stores only `{ brand, last4 }` (e.g. `{ brand: 'visa', last4: '4242' }`); raw `card_number`/`cvc`/full `exp` never persist beyond request scope.

### Success page

- [ ] `/checkout/success?id=<session_id>` server-fetches the session via direct call to `getCheckoutSession(id)` (same Node process; no HTTP loopback needed), renders order id (rendered as `ord_<first 8 of UUID>`), item lines, shipping address recap, total.
- [ ] On mount, the success client clears the local cart (`useCartStore.clearCart()`).

### Coverage

- [ ] Vitest unit + component tests for: `checkout-client` (renders fields, validation errors, submit happy path mocked), `payment` schema (valid/invalid card formats), `checkout` handler (status transitions with shipping+payment optional/present).
- [ ] Playwright e2e at `e2e/checkout.spec.ts`: cart → checkout → fill all fields → success page shows order id. Includes empty-cart redirect and unauthenticated redirect specs.
- [ ] `npm run lint && npm run format` clean. `npm run test` green. `npm run test:e2e -- checkout` green.

---

## Implementation Steps (TDD: write the test first)

### Step 0 — Add missing shadcn primitives

Currently only `button`, `input`, `label` are in `src/components/ui/`. We need:

```bash
cd app/web
npx shadcn@latest add card form separator
```

This drops `card.tsx`, `form.tsx`, `separator.tsx` into `src/components/ui/`. `form.tsx` exposes `<Form>`, `<FormField>`, `<FormItem>`, `<FormLabel>`, `<FormControl>`, `<FormMessage>` — the standard shadcn react-hook-form bindings.

> **Heads-up:** `app/web/AGENTS.md:1-5` says "This is NOT the Next.js you know" — confirm shadcn output compiles against Next 16 / React 19 before relying on the generated files. If the registry produces a deprecated `forwardRef` pattern, prefer the React 19 ref-as-prop form.

### Step 1 — Extend UCP types & schemas

`src/lib/ucp/types/checkout.ts`

```ts
export interface BuyerInfo {
  email?: string;
  name?: string;
  phone?: string;
}

export interface ShippingAddress {
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postal_code: string;
  country: string; // ISO 3166-1 alpha-2
}

export interface PaymentSummary {
  brand: string; // 'visa' | 'mastercard' | 'amex' | 'unknown'
  last4: string; // 4 digits
}

export interface CheckoutSession {
  // ...existing fields...
  buyer?: BuyerInfo;
  shipping_address?: ShippingAddress;
  payment?: PaymentSummary;
}
```

`src/lib/ucp/schemas/checkout.ts` — extend `CreateCheckoutInputSchema` and `UpdateCheckoutInputSchema`. Keep all new fields **optional** at the schema level so the existing MCP machine-flow (email-only) keeps passing validation. Enforce presence only at the state-transition layer (Step 2).

```ts
const ShippingAddressSchema = z.object({
  line1: z.string().min(1).max(200),
  line2: z.string().max(200).optional(),
  city: z.string().min(1).max(100),
  state: z.string().min(1).max(100),
  postal_code: z.string().min(1).max(20),
  country: z.string().length(2),
});

const PaymentInputSchema = z.object({
  card_number: z.string().regex(/^[\d\s]{12,19}$/),
  exp: z.string().regex(/^(0[1-9]|1[0-2])\/\d{2}$/),
  cvc: z.string().regex(/^\d{3,4}$/),
});

export const CreateCheckoutInputSchema = z.object({
  cart_session_id: z.string().min(1),
  currency: z.string().length(3),
  buyer: z.object({
    email: z.email().optional(),
    name: z.string().min(1).max(120).optional(),
    phone: z.string().min(3).max(30).optional(),
  }).optional(),
  shipping_address: ShippingAddressSchema.optional(),
  payment: PaymentInputSchema.optional(),
});

export const UpdateCheckoutInputSchema = z.object({
  buyer: /* same as above */,
  shipping_address: ShippingAddressSchema.optional(),
  payment: PaymentInputSchema.optional(),
});

export function summarizePayment(p: { card_number: string }): PaymentSummary {
  const digits = p.card_number.replace(/\s/g, '');
  const last4 = digits.slice(-4);
  const brand =
    digits.startsWith('4') ? 'visa' :
    /^(5[1-5]|2[2-7])/.test(digits) ? 'mastercard' :
    /^3[47]/.test(digits) ? 'amex' : 'unknown';
  return { brand, last4 };
}
```

**Tests** (`src/lib/ucp/schemas/__tests__/checkout.test.ts` — new):
- valid create input with full buyer + shipping + payment passes
- create input with only `cart_session_id + currency` still passes (back-compat)
- invalid card number / exp / cvc fail
- `summarizePayment('4242 4242 4242 4242')` → `{ brand: 'visa', last4: '4242' }`

### Step 2 — Extend the checkout handler

`src/lib/ucp/handlers/checkout.ts`:

- In `createCheckout`, after merging `input.buyer/shipping_address/payment` into the session, call a new helper `recomputeStatus(session)` to set `status` and `messages[]`. Required fields for `ready_for_complete`: `buyer.email`, `shipping_address` (all required sub-fields), `payment`. If any missing, status stays `incomplete` and messages list each gap with a stable `code` (`missing_email`, `missing_shipping_address`, `missing_payment`).
- `updateCheckout` accepts the new fields, replaces them on the session (full-replace per field, not deep merge — simpler and matches PATCH-as-replace semantics for objects), then `recomputeStatus`.
- **Important: redact payment** — store `session.payment = summarizePayment(input.payment)` (only `brand` + `last4`), never the raw card body.

**Tests** (extend `src/lib/ucp/handlers/__tests__/checkout.test.ts` if it exists, otherwise create):
- create with full body → `status === 'ready_for_complete'`, `messages.length === 0`, `session.payment.last4 === '4242'`
- create with email only → `status === 'incomplete'`, messages contain `missing_shipping_address` and `missing_payment`
- update from incomplete → ready_for_complete when all fields supplied
- raw card number is NOT present anywhere on the returned session (snapshot the JSON, assert no 16-digit string match)

### Step 3 — Mock the order service

`src/lib/ucp/handlers/order.ts`:

```ts
const MOCK = process.env.MOCK_ORDER_SERVICE === 'true';

export async function createOrder(
  session: CheckoutSession,
  items: CartItemData[],
): Promise<{ id: string }> {
  if (MOCK) {
    // Deterministic id derived from session id so success page lookup is stable.
    return { id: `mock_${session.id.slice(0, 8)}` };
  }
  // ...existing fetch...
}
```

**Test** (extend `__tests__` for order.ts — create one if missing): with `MOCK_ORDER_SERVICE=true`, `createOrder` resolves to `{ id: 'mock_...' }` without making a fetch (use `vi.spyOn(globalThis, 'fetch')` and assert `.not.toHaveBeenCalled()`).

### Step 4 — Update `.env.example`

Append:

```
# Enable UCP routes (required for /checkout to function)
UCP_ENABLED=true
# Skip the Go order service in dev; createOrder returns a mock id.
MOCK_ORDER_SERVICE=true
```

Document this in `app/web/CLAUDE.md` under a new "Checkout (mock mode)" subsection (3-4 lines).

### Step 5 — Build the checkout page

**`src/app/checkout/page.tsx`** (server component):

```tsx
import { redirect } from 'next/navigation';
import { auth } from '@/auth';
import { CheckoutClient } from './checkout-client';

export default async function CheckoutPage() {
  const session = await auth();
  if (!session) redirect('/signin?callbackUrl=/checkout');
  return <CheckoutClient defaultEmail={session.user?.email ?? ''} />;
}
```

**`src/app/checkout/checkout-client.tsx`** (client):

- Reads `useCartStore()` for `items, totalItems, totalPrice, clearCart, loadCart`.
- `useEffect(() => { loadCart() }, [loadCart])` (same pattern as `cart-client.tsx`).
- `useEffect`: if `!isLoading && items.length === 0` → `router.push('/cart')`.
- `useForm` with the `CheckoutFormSchema` (zod, separate from UCP schema — adds non-server fields like card_number); on submit:
  1. `POST /api/ucp/checkout` with `{ cart_session_id, currency: 'USD', buyer, shipping_address, payment }` and `Idempotency-Key: crypto.randomUUID()`. Get back `session_id`.
  2. `POST /api/ucp/checkout/[session_id]?action=complete`.
  3. On success: `router.push(\`/checkout/success?id=\${session_id}\`)`.
  4. On 4xx/5xx: surface the UCP `error.code` + `content` in a banner above the form.
- Layout: 2-col CSS grid (`md:grid-cols-[1fr_360px]`); right col is `position: sticky; top: 1rem`. Mobile: stacked.

**`src/app/checkout/order-summary.tsx`** — pure presentational; takes `items, subtotalCents, shippingCents, totalCents`. Subtotal = sum of `priceCents * quantity`. Shipping = `599` (standard from `getFulfillmentOptions`).

**`src/app/checkout/sections/contact-section.tsx`**, `shipping-section.tsx`, `payment-section.tsx` — each receives the form `control` and renders shadcn `<FormField>`s. Splitting now keeps each file <80 lines and easier to read.

**Tests** (`src/app/checkout/__tests__/checkout-client.test.tsx`):
- renders all required fields
- empty submit shows three section-level errors (one for each missing required block)
- happy path: mock `fetch` → POST + complete → assert `router.push` called with `/checkout/success?id=...`
- 502 from complete → error banner is shown, button re-enabled
- double-click submit only fires the first POST (assert `fetch` call count when in-flight)

### Step 6 — Build the success page

**`src/app/checkout/success/page.tsx`** (server):

```tsx
import { notFound } from 'next/navigation';
import { getCheckoutSession } from '@/lib/ucp/handlers/checkout';
import { SuccessClient } from './success-client';

export default async function Page({
  searchParams,
}: { searchParams: Promise<{ id?: string }> }) {
  const { id } = await searchParams;
  if (!id) return <SuccessClient session={null} />;
  const session = await getCheckoutSession(id);
  return <SuccessClient session={session} />;
}
```

> Next 16 makes `searchParams` a promise — `app/web/AGENTS.md:1-5` is the canonical reminder. Confirm via `node_modules/next/dist/docs/` if behavior surprises.

**`src/app/checkout/success/success-client.tsx`** (client) — calls `useCartStore.clearCart()` once on mount; renders order summary, address recap, "Continue shopping" CTA.

**Tests** (`src/app/checkout/success/__tests__/page.test.tsx`):
- with valid `id` and a stored session, renders order id and item lines
- without `id`, renders "no order" empty state
- on mount, `clearCart` is called once (mock the store)

### Step 7 — Wire the cart "Checkout" button

`src/app/cart/cart-client.tsx:163-165` — replace:

```tsx
<Button size="lg" className="w-full">Checkout</Button>
```

with:

```tsx
<Link href="/checkout" className={cn(buttonVariants({ size: 'lg' }), 'w-full')}>
  Checkout
</Link>
```

Update `e2e/cart.spec.ts` if it asserts the button text — keep "Checkout" verbatim.

### Step 8 — End-to-end test

`e2e/checkout.spec.ts` (new):

1. Seed product (re-use `SeedButton` flow) → add to cart → sign in (re-use existing fixture) → navigate to `/checkout`.
2. Fill every form field via labels.
3. Click "Place order" → expect URL `/checkout/success?id=*` and an "Order #ord_…" heading.
4. Empty-cart spec: sign in with empty cart → visit `/checkout` → expect redirect to `/cart`.
5. Unauth spec: sign out → visit `/checkout` → expect redirect to `/signin?callbackUrl=%2Fcheckout`.

Set `UCP_ENABLED=true MOCK_ORDER_SERVICE=true` in `playwright.config.ts` `webServer.env` so e2e doesn't depend on Go services.

### Step 9 — Polish & ship

- `npm run format` (per `app/web/AGENTS.md:7`).
- `npm run lint && npm run test && npm run test:e2e`.
- Update `app/web/CLAUDE.md` UCP section with the new field shape and the `MOCK_ORDER_SERVICE` flag.

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Schema changes ripple to `mcp-tools.ts` and break MCP integration tests | Medium | Medium | Keep all new fields **optional** at the Zod level. The MCP flow already passes only `cart_session_id + currency + buyer.email`. Run `app/web/src/app/api/mcp/__tests__/*.test.ts` after Step 1. |
| `summarizePayment` leaks raw card number through structured-clone of session | Low | High (PII) | Test asserts no 16-digit string appears anywhere in `JSON.stringify(session)`. Compute summary **before** `setSession` and never store the raw `payment` input on the session object. |
| shadcn `form` primitive emits Next-15-incompatible `forwardRef` | Low | Low | `app/web/AGENTS.md:1-5` flags this. If the generated file fails to typecheck, hand-port to React 19 ref-as-prop. |
| `getCheckoutSession` is in-memory & per-process; serverless dev (Turbopack HMR) can drop sessions between request and success-page fetch | Medium | Low | The store already uses `globalThis.__UCP_SESSIONS__` (see `lib/ucp/store.ts`) to survive HMR. Same pattern is used in `order.ts:13-19`. No change needed. |
| Idempotency-Key collision if user resubmits with edited form | Low | Medium | UCP already scopes the key by SHA-256 of (key + body) — different bodies get different cache slots. `route.ts:64-69` is correct. |
| Cart auto-syncs from backend on mount and may briefly show 0 items, triggering the empty-cart redirect on `/checkout` | Medium | Medium | Gate the redirect on `!isLoading && items.length === 0` (the cart store already exposes `isLoading` — see `store/cart.ts:32`). |
| Double-click "Place order" creates two orders | Medium | High | Disable button while the request is in flight AND attach `Idempotency-Key`. Both belt and braces; UCP handles the second case server-side. |
| New `/checkout/success` page leaks order data via shareable URL | Low | Low | Sessions expire in 30 min (`SESSION_TTL_MS`). Acceptable for mock-mode demo; documented. |
| `MOCK_ORDER_SERVICE=true` accidentally ships to prod | Medium | High | Document in `app/web/CLAUDE.md`. Default in `.env.example` stays as a dev-only setting (do NOT add to prod env templates). |

---

## Verification Steps (run in order)

1. `cd app/web && npm install` (if shadcn add pulled new deps).
2. `npm run test` — all unit/component tests green, including new schema/handler/checkout-client tests.
3. `npm run test:e2e -- checkout` — all 5 specs pass.
4. `npm run lint` — biome clean.
5. Manual smoke (per `CLAUDE.md` "test the golden path" rule):
   - `npm run dev`
   - Sign in, add a product to cart, click Checkout → `/checkout` renders with prefilled email.
   - Submit empty form → see 3 section error states.
   - Fill correctly → success page shows order id and clears the cart badge.
   - Browser back from success → does not re-submit (cart is empty so should redirect to `/cart`).
6. Confirm payment fields are absent from network responses by inspecting the response body of `POST /api/ucp/checkout` and `POST /api/ucp/checkout/[id]?action=complete` in DevTools — the only payment field present is `{ brand, last4 }`.

---

## File Manifest

**Create (10 files):**
- `src/app/checkout/page.tsx`
- `src/app/checkout/checkout-client.tsx`
- `src/app/checkout/order-summary.tsx`
- `src/app/checkout/sections/contact-section.tsx`
- `src/app/checkout/sections/shipping-section.tsx`
- `src/app/checkout/sections/payment-section.tsx`
- `src/app/checkout/__tests__/checkout-client.test.tsx`
- `src/app/checkout/success/page.tsx`
- `src/app/checkout/success/success-client.tsx`
- `src/app/checkout/success/__tests__/page.test.tsx`
- `src/lib/ucp/schemas/__tests__/checkout.test.ts` (if not present)
- `e2e/checkout.spec.ts`
- `src/components/ui/{card,form,separator}.tsx` (via shadcn add)

**Modify (5 files):**
- `src/lib/ucp/types/checkout.ts`
- `src/lib/ucp/schemas/checkout.ts`
- `src/lib/ucp/handlers/checkout.ts`
- `src/lib/ucp/handlers/order.ts`
- `src/app/cart/cart-client.tsx` (one-line button → Link swap)
- `app/web/.env.example`
- `app/web/CLAUDE.md` (docs)
- `app/web/playwright.config.ts` (env for e2e)

**Lines of code (rough estimate):** ~700 LOC of source + ~400 LOC of tests.

---

## Out of Scope (explicit non-goals)

- Real payment processing or any third-party gateway SDK.
- Tax computation.
- Discount/coupon entry.
- Multi-currency. Hard-coded `USD`.
- Multiple shipping options selectable in UI (we use the standard `$5.99` from `getFulfillmentOptions` without a selector).
- Address autocomplete / Google Places.
- Saving addresses or cards to the user's profile.
- Auth changes for `/cart` (guest checkout was explicitly declined in interview).
- MCP-side tool description updates (the new fields are optional, MCP behavior unchanged).
- Webhook delivery, order email confirmation.

---

## Build Sequence (suggested)

| Order | Step | Why first |
|-------|------|-----------|
| 1 | Step 0 (shadcn add) | Build will fail without `form.tsx`; do it first to discover any Next-16 compat issues early. |
| 2 | Step 1 + tests | Schema is the contract; lock it before handler logic. |
| 3 | Step 2 + tests | Handler depends on schema. |
| 4 | Step 3 + tests | Tiny env-gate change — slot in before UI to keep API responses honest. |
| 5 | Step 4 (env docs) | So Step 5 dev-mode actually works. |
| 6 | Step 5 + tests | The bulk of the UI work. |
| 7 | Step 6 + tests | Depends on Step 5 having a navigable end. |
| 8 | Step 7 (cart link) | Trivial, do once everything else navigates correctly. |
| 9 | Step 8 (e2e) | Validates the seam between all of the above. |
| 10 | Step 9 (polish) | Last. |
