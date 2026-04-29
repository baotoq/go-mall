const CART_URL =
  process.env.CART_SERVICE_URL ||
  process.env.NEXT_PUBLIC_CART_URL ||
  "http://localhost:9002";

export interface CartItem {
  id: string;
  productId: string;
  quantity: number;
  createdAt: number;
  updatedAt: number;
}

export interface CartListResponse {
  items: CartItem[];
  total: number;
}

const getHeaders = (sessionId: string) => ({
  "Content-Type": "application/json",
  "X-Session-Id": sessionId,
});

export async function getCart(sessionId: string): Promise<CartListResponse> {
  const res = await fetch(`${CART_URL}/api/v1/cart/items`, {
    headers: getHeaders(sessionId),
    cache: "no-store",
  });
  if (!res.ok) throw new Error("Failed to fetch cart");
  return res.json();
}

export async function addToCart(
  sessionId: string,
  productId: string,
  quantity: number,
): Promise<CartItem> {
  const res = await fetch(`${CART_URL}/api/v1/cart/items`, {
    method: "POST",
    headers: getHeaders(sessionId),
    body: JSON.stringify({ productId, quantity }),
  });
  if (!res.ok) throw new Error("Failed to add to cart");
  return res.json();
}

export async function updateCartItem(
  sessionId: string,
  productId: string,
  quantity: number,
): Promise<CartItem> {
  const res = await fetch(`${CART_URL}/api/v1/cart/items`, {
    method: "PATCH",
    headers: getHeaders(sessionId),
    body: JSON.stringify({ productId, quantity }),
  });
  if (!res.ok) throw new Error("Failed to update cart item");
  return res.json();
}

export async function removeCartItem(
  sessionId: string,
  productId: string,
): Promise<void> {
  const res = await fetch(`${CART_URL}/api/v1/cart/items/${productId}`, {
    method: "DELETE",
    headers: getHeaders(sessionId),
  });
  if (!res.ok) throw new Error("Failed to remove cart item");
}
