import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { corsHeaders, withCors } from "../cors";

describe("corsHeaders", () => {
  beforeEach(() => {
    process.env.UCP_ALLOWED_ORIGINS =
      "http://allowed.example,https://other.allowed";
  });

  afterEach(() => {
    delete process.env.UCP_ALLOWED_ORIGINS;
  });

  it("reflects Origin when allow-listed", () => {
    const req = new Request("http://localhost/x", {
      headers: { Origin: "http://allowed.example" },
    });
    const h = corsHeaders(req);
    expect(h["Access-Control-Allow-Origin"]).toBe("http://allowed.example");
  });

  it("omits Allow-Origin when Origin is not allow-listed", () => {
    const req = new Request("http://localhost/x", {
      headers: { Origin: "http://evil.example" },
    });
    const h = corsHeaders(req);
    expect(h["Access-Control-Allow-Origin"]).toBeUndefined();
  });

  it("omits Allow-Origin when no Origin header is present", () => {
    const req = new Request("http://localhost/x");
    const h = corsHeaders(req);
    expect(h["Access-Control-Allow-Origin"]).toBeUndefined();
  });

  it("always includes Vary: Origin and Max-Age", () => {
    const req = new Request("http://localhost/x", {
      headers: { Origin: "http://allowed.example" },
    });
    const h = corsHeaders(req);
    expect(h.Vary).toBe("Origin");
    expect(h["Access-Control-Max-Age"]).toBeTruthy();
  });
});

describe("withCors", () => {
  beforeEach(() => {
    process.env.UCP_ALLOWED_ORIGINS = "http://allowed.example";
  });

  afterEach(() => {
    delete process.env.UCP_ALLOWED_ORIGINS;
  });

  it("merges CORS headers into existing response", () => {
    const req = new Request("http://localhost/x", {
      headers: { Origin: "http://allowed.example" },
    });
    const res = new Response("hi", {
      status: 200,
      headers: { "X-Custom": "1" },
    });
    const out = withCors(res, req);
    expect(out.headers.get("Access-Control-Allow-Origin")).toBe(
      "http://allowed.example",
    );
    expect(out.headers.get("X-Custom")).toBe("1");
  });
});
