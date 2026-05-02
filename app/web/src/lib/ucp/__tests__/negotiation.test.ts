import { describe, it, expect, vi, afterEach } from "vitest";
import {
  isSafeProfileUrl,
  parseUCPAgent,
  negotiateCapabilities,
} from "../negotiation";

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

describe("isSafeProfileUrl (SSRF guard)", () => {
  it.each([
    "https://platform.example.com/.well-known/ucp",
    "https://api.merchant.io/path",
    "https://example.com",
  ])("accepts public https URL: %s", (url) => {
    expect(isSafeProfileUrl(url)).toBe(true);
  });

  it.each([
    "http://platform.example.com/", // non-https
    "ftp://platform.example.com/",
    "file:///etc/passwd",
    "javascript:alert(1)",
  ])("rejects non-https scheme: %s", (url) => {
    expect(isSafeProfileUrl(url)).toBe(false);
  });

  it.each([
    "https://localhost/",
    "https://localhost:8004/.well-known/ucp",
    "https://127.0.0.1/",
    "https://127.10.20.30/",
    "https://0.0.0.0/",
    "https://10.0.0.1/",
    "https://172.16.0.1/",
    "https://172.31.255.255/",
    "https://192.168.1.1/",
    "https://169.254.169.254/", // AWS metadata
  ])("rejects private/loopback IPv4 host: %s", (url) => {
    expect(isSafeProfileUrl(url)).toBe(false);
  });

  it.each([
    "https://[::1]/",
    "https://[fe80::1]/",
    "https://[fc00::1]/",
    "https://[fd12:3456:789a::1]/",
  ])("rejects private/loopback IPv6 host: %s", (url) => {
    expect(isSafeProfileUrl(url)).toBe(false);
  });

  it("rejects malformed URL", () => {
    expect(isSafeProfileUrl("not a url")).toBe(false);
    expect(isSafeProfileUrl("")).toBe(false);
  });
});

describe("negotiateCapabilities SSRF behavior", () => {
  it("does not fetch when URL is unsafe (private host)", async () => {
    const fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);
    const result = await negotiateCapabilities(
      "http://169.254.169.254/latest/meta-data/",
    );
    expect(fetchSpy).not.toHaveBeenCalled();
    expect(result.capabilities).toContain("dev.ucp.shopping.checkout");
  });

  it("does not fetch when URL scheme is non-https", async () => {
    const fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);
    const result = await negotiateCapabilities(
      "http://platform.example.com/.well-known/ucp",
    );
    expect(fetchSpy).not.toHaveBeenCalled();
    expect(result.capabilities).toContain("dev.ucp.shopping.checkout");
  });

  it("passes AbortSignal to fetch for timeout", async () => {
    const fetchSpy = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ ucp: { capabilities: [] } }),
    });
    vi.stubGlobal("fetch", fetchSpy);
    await negotiateCapabilities("https://platform.example.com/.well-known/ucp");
    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const init = fetchSpy.mock.calls[0][1] as RequestInit;
    expect(init.signal).toBeInstanceOf(AbortSignal);
  });
});
