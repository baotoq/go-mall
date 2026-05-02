import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useCartStore } from "@/store/cart";

beforeEach(() => {
  useCartStore.setState({ items: [], isLoading: true });
  localStorage.setItem("cart_session_id", "test-session-id");
});

afterEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});

describe("loadCart", () => {
  it("starts with isLoading true and sets it false when done", async () => {
    // Arrange
    vi.spyOn(global, "fetch").mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: "c", sessionId: "s", items: [], totalCents: 0 }),
    } as Response);
    expect(useCartStore.getState().isLoading).toBe(true);

    // Act
    await useCartStore.getState().loadCart();

    // Assert
    expect(useCartStore.getState().isLoading).toBe(false);
  });

  it("sets isLoading false even when fetch fails", async () => {
    // Arrange
    vi.spyOn(global, "fetch").mockResolvedValueOnce({
      ok: false,
    } as Response);

    // Act
    await useCartStore.getState().loadCart();

    // Assert
    expect(useCartStore.getState().isLoading).toBe(false);
  });

  it("maps camelCase API response fields into store items", async () => {
    // Arrange
    vi.spyOn(global, "fetch").mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        id: "cart-uuid",
        sessionId: "test-session-id",
        items: [
          {
            id: "item-uuid",
            productId: "product-uuid",
            name: "Wireless Headphones",
            priceCents: 9999,
            currency: "USD",
            imageUrl: "http://example.com/img.jpg",
            quantity: 2,
            subtotalCents: 19998,
          },
        ],
        totalCents: 19998,
      }),
    } as Response);

    // Act
    await useCartStore.getState().loadCart();

    // Assert
    const items = useCartStore.getState().items;
    expect(items).toHaveLength(1);
    expect(items[0].id).toBe("product-uuid");
    expect(items[0].priceCents).toBe(9999);
    expect(items[0].imageUrl).toBe("http://example.com/img.jpg");
    expect(items[0].quantity).toBe(2);
  });
});
