import { createHash } from "node:crypto";
import { getCart } from "@/lib/api";
import {
  type CreateCheckoutInput,
  type UpdateCheckoutInput,
  summarizePayment,
  validateIdempotencyKey,
} from "@/lib/ucp/schemas/checkout";
import {
  getIdempotency,
  getSession,
  setIdempotency,
  setSession,
} from "@/lib/ucp/store";
import type {
  CheckoutMessage,
  CheckoutSession,
} from "@/lib/ucp/types/checkout";
import { createOrder, getSessionItems, storeSessionItems } from "./order";

const SESSION_TTL_MS = 30 * 60 * 1000;
const IDEMPOTENCY_TTL_MS = 24 * 60 * 60 * 1000;

const MISSING_EMAIL_MESSAGE: CheckoutMessage = {
  type: "error",
  code: "missing_email",
  path: "$.buyer.email",
  content: "Buyer email is required",
  severity: "recoverable",
};

const MISSING_SHIPPING_MESSAGE: CheckoutMessage = {
  type: "error",
  code: "missing_shipping_address",
  path: "$.shipping_address",
  content: "Shipping address is required",
  severity: "recoverable",
};

const MISSING_PAYMENT_MESSAGE: CheckoutMessage = {
  type: "error",
  code: "missing_payment",
  path: "$.payment",
  content: "Payment information is required",
  severity: "recoverable",
};

export type CreateCheckoutResult =
  | { session: CheckoutSession }
  | { error: true; status: number; code: string; content: string };

export type CompleteCheckoutResult =
  | { session: CheckoutSession }
  | { error: true; status: number; code: string; content: string };

function hashSession(id: string): string {
  return createHash("sha256").update(id).digest("hex");
}

function recomputeStatus(session: CheckoutSession): void {
  const messages: CheckoutMessage[] = [];
  if (!session.buyer?.email) {
    messages.push(MISSING_EMAIL_MESSAGE);
  }
  if (!session.shipping_address) {
    messages.push(MISSING_SHIPPING_MESSAGE);
  }
  if (!session.payment) {
    messages.push(MISSING_PAYMENT_MESSAGE);
  }
  session.messages = messages;
  session.status = messages.length === 0 ? "ready_for_complete" : "incomplete";
}

export async function createCheckout(
  input: CreateCheckoutInput,
  opts: { userId?: string },
): Promise<CreateCheckoutResult> {
  const cart = await getCart(input.cart_session_id);
  if (!cart || cart.items.length === 0) {
    return {
      error: true,
      status: 400,
      code: "empty_cart",
      content: "Cart is empty or not found",
    };
  }

  const mismatched = cart.items.find(
    (item) => item.currency !== input.currency,
  );
  if (mismatched) {
    return {
      error: true,
      status: 400,
      code: "currency_mismatch",
      content: `Cart item currency (${mismatched.currency}) does not match checkout currency (${input.currency})`,
    };
  }

  const now = new Date();
  const id = crypto.randomUUID();

  const session: CheckoutSession = {
    id,
    status: "incomplete",
    currency: input.currency,
    cart_session_id: input.cart_session_id,
    user_id: opts.userId || "guest",
    buyer: input.buyer,
    shipping_address: input.shipping_address,
    payment: input.payment ? summarizePayment(input.payment) : undefined,
    totals: {
      subtotal_cents: cart.totalCents,
      currency: input.currency,
    },
    messages: [],
    expires_at: new Date(now.getTime() + SESSION_TTL_MS).toISOString(),
    created_at: now.toISOString(),
    updated_at: now.toISOString(),
  };

  recomputeStatus(session);

  storeSessionItems(id, cart.items);
  setSession(id, session);
  return { session };
}

export async function getCheckoutSession(
  id: string,
): Promise<CheckoutSession | null> {
  return getSession(id);
}

export async function updateCheckout(
  id: string,
  input: UpdateCheckoutInput,
): Promise<CheckoutSession | null> {
  const session = getSession(id);
  if (!session) return null;

  if (input.buyer !== undefined) {
    session.buyer = input.buyer;
  }
  if (input.shipping_address !== undefined) {
    session.shipping_address = input.shipping_address;
  }
  if (input.payment !== undefined) {
    session.payment = summarizePayment(input.payment);
  }

  recomputeStatus(session);

  session.updated_at = new Date().toISOString();
  setSession(id, session);
  return session;
}

export async function completeCheckout(
  id: string,
  opts: { idempotencyKey?: string | null },
): Promise<CompleteCheckoutResult> {
  const session = getSession(id);
  if (!session) {
    return {
      error: true,
      status: 404,
      code: "not_found",
      content: "Checkout session not found",
    };
  }

  const idempotencyKey = opts.idempotencyKey ?? null;
  if (idempotencyKey) {
    const validation = validateIdempotencyKey(idempotencyKey);
    if (!validation.valid) {
      return {
        error: true,
        status: 400,
        code: "invalid_idempotency_key",
        content: validation.error ?? "Invalid Idempotency-Key",
      };
    }
    const cached = getIdempotency(idempotencyKey);
    if (cached) {
      // Clone on read so callers can't mutate the cached snapshot through the
      // returned reference (defense in depth — we also clone on write below).
      return {
        session: structuredClone(cached.response) as CheckoutSession,
      };
    }
  }

  // SYNCHRONOUS race-condition gate — must occur before any await.
  // Any future refactor that introduces an `await` between `getSession`
  // above and the `setSession` write below silently reopens the
  // double-complete race. Keep this block synchronous.
  if (session.status !== "ready_for_complete") {
    return {
      error: true,
      status: 409,
      code: "invalid_status",
      content: `Cannot complete checkout in status '${session.status}'`,
    };
  }
  session.status = "complete_in_progress";
  session.updated_at = new Date().toISOString();
  setSession(id, session);

  const items = getSessionItems(id);

  try {
    await createOrder(session, items);
  } catch (err) {
    // Surface the upstream failure to logs — silent catches here previously
    // hid the real status/body from the order service and made
    // "order_creation_failed" un-debuggable.
    console.error("[ucp] createOrder failed:", err);
    session.status = "ready_for_complete";
    session.updated_at = new Date().toISOString();
    setSession(id, session);
    return {
      error: true,
      status: 502,
      code: "order_creation_failed",
      content: "Recoverable: order service unavailable",
    };
  }

  session.status = "completed";
  session.updated_at = new Date().toISOString();
  setSession(id, session);

  if (idempotencyKey) {
    // Clone the session snapshot at completion time so later updateCheckout
    // calls (which mutate the live ref) cannot retroactively rewrite the
    // cached idempotency response.
    setIdempotency(idempotencyKey, {
      response: structuredClone(session) as CheckoutSession,
      hash: hashSession(id),
      expires_at: Date.now() + IDEMPOTENCY_TTL_MS,
    });
  }

  return { session };
}
