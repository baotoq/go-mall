import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/ucp/handlers/checkout", () => ({
  createCheckout: vi.fn(),
}));

vi.mock("@/auth", () => ({
  auth: vi.fn().mockResolvedValue(null),
}));

vi.mock("@/lib/ucp/negotiation", () => ({
  parseUCPAgent: vi.fn().mockReturnValue(null),
  negotiateCapabilities: vi.fn().mockResolvedValue({
    capabilities: ["dev.ucp.shopping.checkout"],
    version: "2026-01-11",
  }),
}));

import { createCheckout } from "@/lib/ucp/handlers/checkout";

const mockCreateCheckout = vi.mocked(createCheckout);

function makeRequest(
  body: unknown,
  headers: Record<string, string> = {},
): Request {
  return new Request("http://localhost:3000/api/ucp/checkout", {
    method: "POST",
    headers: { "Content-Type": "application/json", ...headers },
    body: JSON.stringify(body),
  });
}

describe("POST /api/ucp/checkout", () => {
  beforeEach(() => {
    process.env.UCP_ENABLED = "true";
    vi.clearAllMocks();
  });

  afterEach(() => {
    delete process.env.UCP_ENABLED;
  });

  it("returns 503 when UCP_ENABLED is not 'true'", async () => {
    process.env.UCP_ENABLED = "false";
    const { POST } = await import("../route");
    const req = makeRequest({ cart_session_id: "abc", currency: "USD" });
    const res = await POST(req);
    expect(res.status).toBe(503);
    const body = await res.json();
    expect(body.code).toBe("ucp_disabled");
  });

  it("returns 400 when body is invalid (missing cart_session_id)", async () => {
    const { POST } = await import("../route");
    const req = makeRequest({ currency: "USD" });
    const res = await POST(req);
    expect(res.status).toBe(400);
    expect(mockCreateCheckout).not.toHaveBeenCalled();
  });

  it("creates session and returns session_id in body", async () => {
    mockCreateCheckout.mockResolvedValue({
      session: {
        id: "test-id",
        status: "incomplete",
        currency: "USD",
        cart_session_id: "cart-123",
        user_id: "guest",
        totals: { subtotal_cents: 1000, currency: "USD" },
        messages: [],
        expires_at: new Date().toISOString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    });

    const { POST } = await import("../route");
    const req = makeRequest({ cart_session_id: "cart-123", currency: "USD" });
    const res = await POST(req);
    expect(res.status).toBe(201);
    const body = await res.json();
    expect(body.session_id).toBe("test-id");
  });

  it("handles empty cart error", async () => {
    mockCreateCheckout.mockResolvedValue({
      error: true,
      status: 400,
      code: "empty_cart",
      content: "Cart is empty or not found",
    });

    const { POST } = await import("../route");
    const req = makeRequest({ cart_session_id: "empty-cart", currency: "USD" });
    const res = await POST(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.code).toBe("empty_cart");
  });

  it("Idempotency-Key dedup works: second request with same key returns cached", async () => {
    mockCreateCheckout
      .mockResolvedValueOnce({
        session: {
          id: "idem-id",
          status: "incomplete",
          currency: "USD",
          cart_session_id: "cart-456",
          user_id: "guest",
          totals: { subtotal_cents: 500, currency: "USD" },
          messages: [],
          expires_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      })
      .mockResolvedValueOnce({
        session: {
          id: "idem-id",
          status: "incomplete",
          currency: "USD",
          cart_session_id: "cart-456",
          user_id: "guest",
          totals: { subtotal_cents: 500, currency: "USD" },
          messages: [],
          expires_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      });

    const { POST } = await import("../route");

    const req1 = makeRequest(
      { cart_session_id: "cart-456", currency: "USD" },
      { "Idempotency-Key": "key-abc-123" },
    );
    const res1 = await POST(req1);
    expect(res1.status).toBe(201);
    const body1 = await res1.json();
    expect(body1.session_id).toBe("idem-id");

    const req2 = makeRequest(
      { cart_session_id: "cart-456", currency: "USD" },
      { "Idempotency-Key": "key-abc-123" },
    );
    const res2 = await POST(req2);
    expect(res2.status).toBe(201);
    const body2 = await res2.json();
    expect(body2.session_id).toBe("idem-id");
  });
});
