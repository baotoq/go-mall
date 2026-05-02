import type { CartItemData } from "@/lib/types";
import type { CheckoutSession } from "@/lib/ucp/types/checkout";

const ORDER_API_URL = process.env.ORDER_API_URL ?? "http://localhost:8004";

const sessionItemsStore = new Map<string, CartItemData[]>();

export function storeSessionItems(id: string, items: CartItemData[]): void {
  sessionItemsStore.set(id, items);
}

export function getSessionItems(id: string): CartItemData[] {
  return sessionItemsStore.get(id) ?? [];
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
  });
  if (!res.ok) {
    throw new Error(`Order creation failed: ${res.status}`);
  }
  return res.json();
}
