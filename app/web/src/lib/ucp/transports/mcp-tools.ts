export const MCP_TOOL_NAMES = [
  "ucp_get_profile",
  "ucp_create_checkout",
  "ucp_get_checkout",
  "ucp_update_checkout",
  "ucp_complete_checkout",
] as const;

export type McpToolName = (typeof MCP_TOOL_NAMES)[number];
