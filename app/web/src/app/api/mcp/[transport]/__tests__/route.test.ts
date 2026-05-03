import { beforeEach, describe, expect, it, vi } from "vitest";

// mcp-handler returns a function-shaped handler; we mock it so the test stays
// hermetic and does not require a live MCP transport.
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

describe("MCP transport", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
  });

  it("delegates to MCP handler on POST", async () => {
    const { POST } = await import("../route");
    const res = await POST(makeRequest("POST"));
    expect(res.status).toBe(200);
  });

  it("delegates to MCP handler on GET", async () => {
    const { GET } = await import("../route");
    const res = await GET(makeRequest("GET"));
    expect(res.status).toBe(200);
  });
});
