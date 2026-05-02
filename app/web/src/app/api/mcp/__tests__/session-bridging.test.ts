import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/ucp/handlers/checkout");

import { getCheckoutSession } from "@/lib/ucp/handlers/checkout";

describe("MCP session bridging", () => {
  it("session_id from create can be used with get", async () => {
    // Arrange
    vi.mocked(getCheckoutSession).mockResolvedValue({
      id: "sess-123",
      status: "incomplete",
      currency: "USD",
      cart_session_id: "cart-xyz",
      user_id: "guest",
      totals: { subtotal_cents: 2000, currency: "USD" },
      messages: [],
      expires_at: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });

    // Act
    const session = await getCheckoutSession("sess-123");

    // Assert
    expect(session).not.toBeNull();
    expect(session?.id).toBe("sess-123");
  });

  it("missing session returns null", async () => {
    // Arrange
    vi.mocked(getCheckoutSession).mockResolvedValue(null);

    // Act
    const result = await getCheckoutSession("non-existent");

    // Assert
    expect(result).toBeNull();
  });
});
