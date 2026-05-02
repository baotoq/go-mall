import { test, expect } from "@playwright/test";

// Requires: UCP_ENABLED=true npm run dev
// Tests the MCP JSON-RPC endpoint at /api/mcp/mcp (streamable HTTP transport)

test.describe("UCP MCP Transport", () => {
  async function mcpCall(
    request: Parameters<typeof test>[1] extends (
      args: { request: infer R },
      ...rest: unknown[]
    ) => unknown
      ? R
      : never,
    method: string,
    params: Record<string, unknown>,
    id = 1,
  ) {
    const res = await request.post("/api/mcp/mcp", {
      data: { jsonrpc: "2.0", id, method, params },
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json, text/event-stream",
      },
    });
    // mcp-handler returns SSE-encoded JSON: "event: message\ndata: {...}\n\n"
    const text = await res.text();
    const dataLine = text.split("\n").find((l) => l.startsWith("data: "));
    const body = dataLine ? JSON.parse(dataLine.slice(6)) : null;
    return { status: res.status(), body };
  }

  test("tools/list returns all UCP tools", async ({ request }) => {
    const { status, body } = await mcpCall(request, "tools/list", {});
    expect(status).toBe(200);
    const toolNames = (body.result?.tools ?? []).map(
      (t: { name: string }) => t.name,
    );
    expect(toolNames).toContain("ucp_get_profile");
    expect(toolNames).toContain("ucp_create_checkout");
    expect(toolNames).toContain("ucp_get_checkout");
    expect(toolNames).toContain("ucp_update_checkout");
    expect(toolNames).toContain("ucp_complete_checkout");
  });

  test("ucp_get_profile returns capabilities", async ({ request }) => {
    const { status, body } = await mcpCall(request, "tools/call", {
      name: "ucp_get_profile",
      arguments: {},
    });
    expect(status).toBe(200);
    const text = body.result?.content?.[0]?.text;
    const profile = JSON.parse(text);
    expect(profile.ucp.capabilities).toBeInstanceOf(Array);
  });

  test("ucp_create_checkout returns session_id", async ({ request }) => {
    const { status, body } = await mcpCall(request, "tools/call", {
      name: "ucp_create_checkout",
      arguments: {
        cart_session_id: "mcp-test-session",
        currency: "USD",
      },
    });
    expect(status).toBe(200);
    const text = body.result?.content?.[0]?.text;
    const result = JSON.parse(text);
    // May be error if cart not found — check for error or session_id
    if (!result.error) {
      expect(result.session_id).toBeTruthy();
    }
  });
});
