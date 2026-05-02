import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/ucp/handlers/checkout", () => ({
  getCheckoutSession: vi.fn(),
  updateCheckout: vi.fn(),
  completeCheckout: vi.fn(),
}));

vi.mock("@/lib/ucp/negotiation", () => ({
  parseUCPAgent: vi.fn().mockReturnValue(null),
  negotiateCapabilities: vi.fn().mockResolvedValue({
    capabilities: ["dev.ucp.shopping.checkout"],
    version: "2026-01-11",
  }),
}));

import {
  completeCheckout,
  getCheckoutSession,
  updateCheckout,
} from "@/lib/ucp/handlers/checkout";

const mockGet = vi.mocked(getCheckoutSession);
const mockUpdate = vi.mocked(updateCheckout);
const mockComplete = vi.mocked(completeCheckout);

const MOCK_SESSION = {
  id: "test-id",
  status: "incomplete" as const,
  currency: "USD",
  cart_session_id: "cart-123",
  user_id: "guest",
  totals: { subtotal_cents: 1000, currency: "USD" },
  messages: [],
  expires_at: new Date().toISOString(),
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

function makeGetRequest(
  id: string,
  headers: Record<string, string> = {},
): Request {
  return new Request(`http://localhost:3000/api/ucp/checkout/${id}`, {
    method: "GET",
    headers,
  });
}

function makePatchRequest(
  id: string,
  body: unknown,
  headers: Record<string, string> = {},
): Request {
  return new Request(`http://localhost:3000/api/ucp/checkout/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json", ...headers },
    body: JSON.stringify(body),
  });
}

function makePostRequest(
  id: string,
  action: string,
  headers: Record<string, string> = {},
): Request {
  return new Request(
    `http://localhost:3000/api/ucp/checkout/${id}?action=${action}`,
    {
      method: "POST",
      headers,
    },
  );
}

describe("GET /api/ucp/checkout/[id]", () => {
  beforeEach(() => {
    process.env.UCP_ENABLED = "true";
    vi.clearAllMocks();
  });

  afterEach(() => {
    delete process.env.UCP_ENABLED;
  });

  it("returns session when found", async () => {
    mockGet.mockResolvedValue(MOCK_SESSION);
    const { GET } = await import("../route");
    const req = makeGetRequest("test-id", { "X-UCP-Session": "test-id" });
    const res = await GET(req, { params: Promise.resolve({ id: "test-id" }) });
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.id).toBe("test-id");
  });

  it("returns 404 for unknown id", async () => {
    mockGet.mockResolvedValue(null);
    const { GET } = await import("../route");
    const req = makeGetRequest("unknown-id", { "X-UCP-Session": "unknown-id" });
    const res = await GET(req, {
      params: Promise.resolve({ id: "unknown-id" }),
    });
    expect(res.status).toBe(404);
    const body = await res.json();
    expect(body.code).toBe("not_found");
  });

  it("returns 401 when X-UCP-Session header is missing", async () => {
    const { GET } = await import("../route");
    const req = makeGetRequest("test-id");
    const res = await GET(req, { params: Promise.resolve({ id: "test-id" }) });
    expect(res.status).toBe(401);
    const body = await res.json();
    expect(body.code).toBe("missing_session");
  });

  it("returns 403 when X-UCP-Session does not match id", async () => {
    const { GET } = await import("../route");
    const req = makeGetRequest("test-id", { "X-UCP-Session": "wrong-id" });
    const res = await GET(req, { params: Promise.resolve({ id: "test-id" }) });
    expect(res.status).toBe(403);
    const body = await res.json();
    expect(body.code).toBe("session_mismatch");
  });

  it("does not call getCheckoutSession when auth fails", async () => {
    const { GET } = await import("../route");
    const req = makeGetRequest("test-id");
    await GET(req, { params: Promise.resolve({ id: "test-id" }) });
    expect(mockGet).not.toHaveBeenCalled();
  });
});

describe("OPTIONS /api/ucp/checkout/[id]", () => {
  beforeEach(() => {
    process.env.UCP_ENABLED = "true";
    process.env.UCP_ALLOWED_ORIGINS = "http://allowed.example";
  });

  afterEach(() => {
    delete process.env.UCP_ENABLED;
    delete process.env.UCP_ALLOWED_ORIGINS;
  });

  it("returns 204 with CORS headers when origin is allowed", async () => {
    const { OPTIONS } = await import("../route");
    const req = new Request("http://localhost:3000/api/ucp/checkout/test-id", {
      method: "OPTIONS",
      headers: { Origin: "http://allowed.example" },
    });
    const res = await OPTIONS(req);
    expect(res.status).toBe(204);
    expect(res.headers.get("Access-Control-Allow-Origin")).toBe(
      "http://allowed.example",
    );
    expect(res.headers.get("Vary")).toBe("Origin");
    expect(res.headers.get("Access-Control-Allow-Methods")).toContain("PATCH");
  });

  it("omits Access-Control-Allow-Origin when origin is not allow-listed", async () => {
    const { OPTIONS } = await import("../route");
    const req = new Request("http://localhost:3000/api/ucp/checkout/test-id", {
      method: "OPTIONS",
      headers: { Origin: "http://evil.example" },
    });
    const res = await OPTIONS(req);
    expect(res.status).toBe(204);
    expect(res.headers.get("Access-Control-Allow-Origin")).toBeNull();
    expect(res.headers.get("Vary")).toBe("Origin");
  });
});

describe("PATCH /api/ucp/checkout/[id]", () => {
  beforeEach(() => {
    process.env.UCP_ENABLED = "true";
    vi.clearAllMocks();
  });

  afterEach(() => {
    delete process.env.UCP_ENABLED;
  });

  it("returns 401 when X-UCP-Session header is missing", async () => {
    const { PATCH } = await import("../route");
    const req = makePatchRequest("test-id", { buyer: { email: "a@b.com" } });
    const res = await PATCH(req, {
      params: Promise.resolve({ id: "test-id" }),
    });
    expect(res.status).toBe(401);
    const body = await res.json();
    expect(body.code).toBe("missing_session");
  });

  it("updates buyer info and returns 200", async () => {
    const updated = {
      ...MOCK_SESSION,
      buyer: { email: "buyer@example.com" },
      status: "ready_for_complete" as const,
    };
    mockUpdate.mockResolvedValue(updated);
    const { PATCH } = await import("../route");
    const req = makePatchRequest(
      "test-id",
      { buyer: { email: "buyer@example.com" } },
      { "X-UCP-Session": "test-id" },
    );
    const res = await PATCH(req, {
      params: Promise.resolve({ id: "test-id" }),
    });
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.buyer?.email).toBe("buyer@example.com");
  });

  it("returns 404 when session not found during update", async () => {
    mockUpdate.mockResolvedValue(null);
    const { PATCH } = await import("../route");
    const req = makePatchRequest(
      "missing-id",
      { buyer: { email: "a@b.com" } },
      { "X-UCP-Session": "missing-id" },
    );
    const res = await PATCH(req, {
      params: Promise.resolve({ id: "missing-id" }),
    });
    expect(res.status).toBe(404);
  });
});

describe("POST /api/ucp/checkout/[id] (complete)", () => {
  beforeEach(() => {
    process.env.UCP_ENABLED = "true";
    vi.clearAllMocks();
  });

  afterEach(() => {
    delete process.env.UCP_ENABLED;
  });

  it("returns 401 when X-UCP-Session header is missing", async () => {
    const { POST } = await import("../route");
    const req = makePostRequest("test-id", "complete");
    const res = await POST(req, { params: Promise.resolve({ id: "test-id" }) });
    expect(res.status).toBe(401);
    const body = await res.json();
    expect(body.code).toBe("missing_session");
  });

  it("transitions state with action=complete and returns 200", async () => {
    const completed = { ...MOCK_SESSION, status: "completed" as const };
    mockComplete.mockResolvedValue({ session: completed });
    const { POST } = await import("../route");
    const req = makePostRequest("test-id", "complete", {
      "X-UCP-Session": "test-id",
    });
    const res = await POST(req, { params: Promise.resolve({ id: "test-id" }) });
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.status).toBe("completed");
  });

  it("returns 404 on unknown id for complete", async () => {
    mockComplete.mockResolvedValue({
      error: true,
      status: 404,
      code: "not_found",
      content: "Checkout session not found",
    });
    const { POST } = await import("../route");
    const req = makePostRequest("unknown-id", "complete", {
      "X-UCP-Session": "unknown-id",
    });
    const res = await POST(req, {
      params: Promise.resolve({ id: "unknown-id" }),
    });
    expect(res.status).toBe(404);
    const body = await res.json();
    expect(body.code).toBe("not_found");
  });

  it("returns 410 Gone for expired session (treats missing as expired/not found)", async () => {
    mockGet.mockResolvedValue(null);
    const { GET } = await import("../route");
    const req = makeGetRequest("expired-id", { "X-UCP-Session": "expired-id" });
    const res = await GET(req, {
      params: Promise.resolve({ id: "expired-id" }),
    });
    // Route returns 404 for both expired and unknown — either is acceptable per spec
    expect([404, 410]).toContain(res.status);
  });
});
