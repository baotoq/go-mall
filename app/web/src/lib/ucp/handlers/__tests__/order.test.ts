import { afterEach, describe, expect, it, vi } from "vitest";
import type { CartItemData } from "@/lib/types";
import type { CheckoutSession } from "@/lib/ucp/types/checkout";
import { createOrder } from "../order";

afterEach(() => {
  vi.restoreAllMocks();
});

function makeSession(): CheckoutSession {
  return {
    id: "sess-1",
    status: "ready_for_complete",
    currency: "USD",
    cart_session_id: "cart-1",
    user_id: "guest",
    totals: { subtotal_cents: 1000, currency: "USD" },
    messages: [],
    expires_at: new Date(Date.now() + 60_000).toISOString(),
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
}

function makeItems(): CartItemData[] {
  return [
    {
      id: "item-1",
      productId: "prod-1",
      name: "Widget",
      priceCents: 1000,
      currency: "USD",
      imageUrl: "https://example.com/w.png",
      quantity: 1,
      subtotalCents: 1000,
    },
  ];
}

describe("createOrder", () => {
  it("includes upstream response body in thrown Error on non-2xx", async () => {
    const fetchSpy = vi.fn().mockResolvedValue({
      ok: false,
      status: 422,
      text: async () => '{"reason":"INVALID_USER_ID"}',
      json: async () => ({}),
    });
    vi.stubGlobal("fetch", fetchSpy);

    await expect(createOrder(makeSession(), makeItems())).rejects.toThrow(
      /422.*INVALID_USER_ID/,
    );
  });

  it("passes AbortSignal for timeout to fetch", async () => {
    const fetchSpy = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "order-1" }),
    });
    vi.stubGlobal("fetch", fetchSpy);

    await createOrder(makeSession(), makeItems());

    const init = fetchSpy.mock.calls[0][1] as RequestInit;
    expect(init.signal).toBeInstanceOf(AbortSignal);
  });

  it("returns parsed JSON body on success", async () => {
    const fetchSpy = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "order-42" }),
    });
    vi.stubGlobal("fetch", fetchSpy);

    const result = await createOrder(makeSession(), makeItems());
    expect(result.id).toBe("order-42");
  });
});
