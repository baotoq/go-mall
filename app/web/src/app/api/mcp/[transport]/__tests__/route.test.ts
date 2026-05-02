import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// mcp-handler returns a function-shaped handler; we mock it so the kill-switch
// test is hermetic and does not require a live MCP transport.
vi.mock("mcp-handler", () => ({
  createMcpHandler: vi.fn(
    () =>
      async function fakeHandler() {
        return new Response(
          JSON.stringify({ ok: true, jsonrpc: "2.0", id: 1, result: {} }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      },
  ),
}));

function makeRequest(method: "GET" | "POST"): Request {
  return new Request("http://localhost:3000/api/mcp/http", {
    method,
    headers: { "Content-Type": "application/json" },
    body: method === "POST" ? JSON.stringify({}) : undefined,
  });
}

describe("MCP transport kill switch", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
  });

  afterEach(() => {
    delete process.env.UCP_ENABLED;
  });

  it("returns 503 ucp_disabled on GET when UCP_ENABLED is unset", async () => {
    delete process.env.UCP_ENABLED;
    const { GET } = await import("../route");
    const res = await GET(makeRequest("GET"));
    expect(res.status).toBe(503);
    const body = await res.json();
    expect(body.code).toBe("ucp_disabled");
  });

  it("returns 503 ucp_disabled on POST when UCP_ENABLED=false", async () => {
    process.env.UCP_ENABLED = "false";
    const { POST } = await import("../route");
    const res = await POST(makeRequest("POST"));
    expect(res.status).toBe(503);
    const body = await res.json();
    expect(body.code).toBe("ucp_disabled");
  });

  it("delegates to MCP handler when UCP_ENABLED=true", async () => {
    process.env.UCP_ENABLED = "true";
    const { POST } = await import("../route");
    const res = await POST(makeRequest("POST"));
    expect(res.status).toBe(200);
  });
});
