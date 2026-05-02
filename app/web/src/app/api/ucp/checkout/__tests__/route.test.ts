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

// Real-ish in-memory store so idempotency dedup is actually exercised.
const idemStore = new Map<
  string,
  { response: unknown; hash: string; expires_at: number }
>();
vi.mock("@/lib/ucp/store", () => ({
  getIdempotency: vi.fn((key: string) => idemStore.get(key) ?? null),
  setIdempotency: vi.fn(
    (
      key: string,
      entry: { response: unknown; hash: string; expires_at: number },
    ) => {
      idemStore.set(key, entry);
    },
  ),
  getSession: vi.fn(),
  setSession: vi.fn(),
  initStore: vi.fn(),
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
    mockCreateCheckout.mockReset();
    vi.clearAllMocks();
    idemStore.clear();
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
    // First call creates session with id "idem-id"
    mockCreateCheckout.mockResolvedValueOnce({
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
    // Second call would create a DIFFERENT session if dedup were absent.
    mockCreateCheckout.mockResolvedValueOnce({
      session: {
        id: "different-id",
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

    // Second request with same key must return the cached session, not "different-id".
    const req2 = makeRequest(
      { cart_session_id: "cart-456", currency: "USD" },
      { "Idempotency-Key": "key-abc-123" },
    );
    const res2 = await POST(req2);
    expect(res2.status).toBe(201);
    const body2 = await res2.json();
    expect(body2.session_id).toBe("idem-id");
    // createCheckout must only have been called once — the second was served from cache.
    expect(mockCreateCheckout).toHaveBeenCalledTimes(1);
  });

  it("Idempotency-Key scoping: same key with different cart_session_id is NOT deduped", async () => {
    // Two distinct carts with the same key must each create their own session.
    mockCreateCheckout
      .mockResolvedValueOnce({
        session: {
          id: "session-cart-a",
          status: "incomplete",
          currency: "USD",
          cart_session_id: "cart-a",
          user_id: "guest",
          totals: { subtotal_cents: 100, currency: "USD" },
          messages: [],
          expires_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      })
      .mockResolvedValueOnce({
        session: {
          id: "session-cart-b",
          status: "incomplete",
          currency: "USD",
          cart_session_id: "cart-b",
          user_id: "guest",
          totals: { subtotal_cents: 200, currency: "USD" },
          messages: [],
          expires_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      });

    const { POST } = await import("../route");

    const res1 = await POST(
      makeRequest(
        { cart_session_id: "cart-a", currency: "USD" },
        { "Idempotency-Key": "shared-key" },
      ),
    );
    const res2 = await POST(
      makeRequest(
        { cart_session_id: "cart-b", currency: "USD" },
        { "Idempotency-Key": "shared-key" },
      ),
    );

    expect((await res1.json()).session_id).toBe("session-cart-a");
    expect((await res2.json()).session_id).toBe("session-cart-b");
    // Both carts must have triggered a real createCheckout call.
    expect(mockCreateCheckout).toHaveBeenCalledTimes(2);
  });
});
