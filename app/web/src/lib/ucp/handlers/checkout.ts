import { createHash } from "node:crypto";
import { getCart } from "@/lib/api";
import {
  type CreateCheckoutInput,
  type UpdateCheckoutInput,
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

export type CreateCheckoutResult =
  | { session: CheckoutSession }
  | { error: true; status: number; code: string; content: string };

export type CompleteCheckoutResult =
  | { session: CheckoutSession }
  | { error: true; status: number; code: string; content: string };

function hashSession(id: string): string {
  return createHash("sha256").update(id).digest("hex");
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
  const messages: CheckoutMessage[] = [];
  if (!input.buyer?.email) {
    messages.push(MISSING_EMAIL_MESSAGE);
  }

  const session: CheckoutSession = {
    id,
    status: "incomplete",
    currency: input.currency,
    cart_session_id: input.cart_session_id,
    user_id: opts.userId || "guest",
    buyer: input.buyer,
    totals: {
      subtotal_cents: cart.totalCents,
      currency: input.currency,
    },
    messages,
    expires_at: new Date(now.getTime() + SESSION_TTL_MS).toISOString(),
    created_at: now.toISOString(),
    updated_at: now.toISOString(),
  };

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

  if (input.buyer?.email) {
    session.buyer = { ...session.buyer, email: input.buyer.email };
    session.messages = session.messages.filter(
      (m) => m.code !== "missing_email",
    );
    session.status = "ready_for_complete";
  }

  session.updated_at = new Date().toISOString();
  setSession(id, session);
  return session;
}

export async function completeCheckout(
  id: string,
  opts: { idempotencyKey?: string | null; userId?: string },
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
      return { session: cached.response as CheckoutSession };
    }
  }

  // SYNCHRONOUS race-condition gate — must occur before any await
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
  } catch {
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
    setIdempotency(idempotencyKey, {
      response: session,
      hash: hashSession(id),
      expires_at: Date.now() + IDEMPOTENCY_TTL_MS,
    });
  }

  return { session };
}
