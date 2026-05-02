import { describe, expect, it } from "vitest";

describe("ucp.config.json contract", () => {
  it("has required shape", async () => {
    const config = await import("../../../../ucp.config.json");
    expect(config.roles).toContain("business");
    expect(config.transports).toContain("rest");
    expect(config.transports).toContain("mcp");
    expect(config.payment_handlers).toHaveLength(0);
    expect(config.features.ap2_mandates).toBe(false);
  });
});
