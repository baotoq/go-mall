import { describe, it, expect, vi } from "vitest";

vi.mock("next-auth", () => ({ default: vi.fn(() => ({ auth: vi.fn() })) }));
vi.mock("@/auth.config", () => ({ authConfig: {} }));

const { config } = await import("../../proxy");

describe("proxy.ts matcher excludes UCP and MCP routes", () => {
  const { matcher } = config;

  function matches(path: string): boolean {
    return matcher.some((pattern) => {
      const re = new RegExp(
        `^${pattern.replace(/:[^/]+\*/g, ".*").replace(/:[^/]+/g, "[^/]+")}$`,
      );
      return re.test(path);
    });
  }

  it("matches /cart", () => {
    expect(matches("/cart")).toBe(true);
  });

  it("matches /cart/anything", () => {
    expect(matches("/cart/items")).toBe(true);
  });

  it("does NOT match /api/ucp/test", () => {
    expect(matches("/api/ucp/test")).toBe(false);
  });

  it("does NOT match /api/mcp/test", () => {
    expect(matches("/api/mcp/test")).toBe(false);
  });

  it("does NOT match /api/ucp/checkout", () => {
    expect(matches("/api/ucp/checkout")).toBe(false);
  });
});
