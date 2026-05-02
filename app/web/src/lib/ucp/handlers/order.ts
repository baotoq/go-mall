import type { CartItemData } from "@/lib/types";
import type { CheckoutSession } from "@/lib/ucp/types/checkout";

const ORDER_API_URL = process.env.ORDER_API_URL ?? "http://localhost:8004";
const ORDER_TIMEOUT_MS = 5_000;
const ITEMS_TTL_MS = 30 * 60 * 1000;

type ItemsEntry = { items: CartItemData[]; expires_at: number };

// Survive HMR: see lib/ucp/store.ts for rationale.
declare global {
  // eslint-disable-next-line no-var
  var __UCP_SESSION_ITEMS__: Map<string, ItemsEntry> | undefined;
}

const sessionItemsStore: Map<string, ItemsEntry> =
  globalThis.__UCP_SESSION_ITEMS__ ??
  (globalThis.__UCP_SESSION_ITEMS__ = new Map());

export function storeSessionItems(id: string, items: CartItemData[]): void {
  sessionItemsStore.set(id, {
    items,
    expires_at: Date.now() + ITEMS_TTL_MS,
  });
}

export function getSessionItems(id: string): CartItemData[] {
  const entry = sessionItemsStore.get(id);
  if (!entry) return [];
  if (entry.expires_at < Date.now()) {
    sessionItemsStore.delete(id);
    return [];
  }
  return entry.items;
}

export async function createOrder(
  session: CheckoutSession,
  items: CartItemData[],
): Promise<{ id: string }> {
  const res = await fetch(`${ORDER_API_URL}/v1/orders`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      user_id: session.user_id,
      session_id: session.cart_session_id,
      currency: session.currency,
      items: items.map((item) => ({
        product_id: item.productId,
        name: item.name,
        price_cents: item.priceCents,
        image_url: item.imageUrl,
        quantity: item.quantity,
      })),
    }),
    signal: AbortSignal.timeout(ORDER_TIMEOUT_MS),
  });
  if (!res.ok) {
    // Surface the upstream error_reason so callers see why kratos rejected.
    const detail = await res.text().catch(() => "");
    const suffix = detail ? ` ${detail}` : "";
    throw new Error(`Order creation failed: ${res.status}${suffix}`);
  }
  return res.json();
}
