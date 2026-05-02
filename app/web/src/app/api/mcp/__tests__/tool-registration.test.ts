import { describe, it, expect } from "vitest";
import { MCP_TOOL_NAMES } from "@/lib/ucp/transports/mcp-tools";

describe("MCP tool registration", () => {
  it("includes all required UCP tools", () => {
    expect(MCP_TOOL_NAMES).toContain("ucp_get_profile");
    expect(MCP_TOOL_NAMES).toContain("ucp_create_checkout");
    expect(MCP_TOOL_NAMES).toContain("ucp_get_checkout");
    expect(MCP_TOOL_NAMES).toContain("ucp_update_checkout");
    expect(MCP_TOOL_NAMES).toContain("ucp_complete_checkout");
  });

  it("has exactly 5 tools", () => {
    expect(MCP_TOOL_NAMES).toHaveLength(5);
  });
});
