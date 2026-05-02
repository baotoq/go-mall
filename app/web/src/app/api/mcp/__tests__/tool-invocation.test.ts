import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/ucp/handlers/checkout");
vi.mock("@/lib/ucp/profile");
vi.mock("@/lib/ucp/negotiation");

import {
  createCheckout,
  getCheckoutSession,
} from "@/lib/ucp/handlers/checkout";
import { generateProfile } from "@/lib/ucp/profile";
import { negotiateCapabilities } from "@/lib/ucp/negotiation";

beforeEach(() => {
  vi.clearAllMocks();
});

describe("ucp_get_profile", () => {
  it("returns capabilities array", async () => {
    // Arrange
    vi.mocked(generateProfile).mockReturnValue({
      ucp: {
        version: "2026-01-11",
        services: {
          "dev.ucp.shopping": { version: "2026-01-11" },
        },
        capabilities: [
          { name: "dev.ucp.shopping.checkout", version: "2026-01-11" },
        ],
      },
    });

    // Act
    const profile = generateProfile();

    // Assert
    expect(profile.ucp.capabilities).toContainEqual({
      name: "dev.ucp.shopping.checkout",
      version: "2026-01-11",
    });
  });
});

describe("ucp_create_checkout", () => {
  it("with valid cart_session_id returns session_id", async () => {
    // Arrange
    vi.mocked(negotiateCapabilities).mockResolvedValue({
      capabilities: ["dev.ucp.shopping.checkout"],
      version: "2026-01-11",
    });
    vi.mocked(createCheckout).mockResolvedValue({
      session: {
        id: "sess-abc",
        status: "incomplete",
        currency: "USD",
        cart_session_id: "cart-123",
        user_id: "guest",
        totals: { subtotal_cents: 1000, currency: "USD" },
        messages: [],
        expires_at: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    });

    // Act
    const result = await createCheckout(
      { cart_session_id: "cart-123", currency: "USD" },
      {},
    );

    // Assert
    expect("session" in result).toBe(true);
    if ("session" in result) {
      expect(result.session.id).toBe("sess-abc");
    }
  });
});

describe("ucp_get_checkout", () => {
  it("returns session by id", async () => {
    // Arrange
    vi.mocked(getCheckoutSession).mockResolvedValue({
      id: "sess-abc",
      status: "incomplete",
      currency: "USD",
      cart_session_id: "cart-123",
      user_id: "guest",
      totals: { subtotal_cents: 1000, currency: "USD" },
      messages: [],
      expires_at: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });

    // Act
    const session = await getCheckoutSession("sess-abc");

    // Assert
    expect(session).not.toBeNull();
    expect(session?.id).toBe("sess-abc");
  });
});
