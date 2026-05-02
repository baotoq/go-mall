import { describe, it, expect, vi, afterEach } from "vitest";
import { parseUCPAgent, negotiateCapabilities } from "../negotiation";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("parseUCPAgent", () => {
  it("parses profile header", () => {
    const result = parseUCPAgent(
      'profile="https://example.com/.well-known/ucp"',
    );
    expect(result).toEqual({ profile: "https://example.com/.well-known/ucp" });
  });

  it("returns null on null header", () => {
    expect(parseUCPAgent(null)).toBeNull();
  });

  it("returns null on empty header", () => {
    expect(parseUCPAgent("")).toBeNull();
  });
});

describe("negotiateCapabilities", () => {
  it("returns business caps when no platform URL", async () => {
    const result = await negotiateCapabilities();
    expect(result.capabilities).toContain("dev.ucp.shopping.checkout");
  });

  it("computes intersection with platform profile", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          ucp: {
            capabilities: [{ name: "dev.ucp.shopping.checkout" }],
          },
        }),
      }),
    );

    const result = await negotiateCapabilities(
      "https://platform.example.com/.well-known/ucp",
    );
    expect(result.capabilities).toEqual(["dev.ucp.shopping.checkout"]);
    expect(result.capabilities).not.toContain("dev.ucp.shopping.fulfillment");
  });

  it("fallback on fetch error returns full business capabilities", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("network error")),
    );

    const result = await negotiateCapabilities(
      "https://platform.example.com/.well-known/ucp",
    );
    expect(result.capabilities).toContain("dev.ucp.shopping.checkout");
    expect(result.capabilities).toContain("dev.ucp.shopping.fulfillment");
  });
});
