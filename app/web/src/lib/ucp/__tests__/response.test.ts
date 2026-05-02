import { describe, it, expect } from "vitest";
import { wrapResponse, errorResponse } from "../response";
import type { NegotiationResult } from "../negotiation";

describe("wrapResponse", () => {
  it("includes ucp.capabilities in response body", async () => {
    const negotiation: NegotiationResult = {
      capabilities: ["dev.ucp.shopping.checkout"],
      version: "2026-01-11",
    };
    const data = { orderId: "123", total: 99.99 };
    const response = wrapResponse(data, negotiation);
    const body = await response.json();
    expect(body["ucp"]).toBeDefined();
    expect(body["ucp"].capabilities).toEqual(["dev.ucp.shopping.checkout"]);
  });

  it("uses provided status code", async () => {
    const negotiation: NegotiationResult = {
      capabilities: ["dev.ucp.shopping.checkout"],
      version: "2026-01-11",
    };
    const response = wrapResponse({ orderId: "456" }, negotiation, 201);
    expect(response.status).toBe(201);
  });
});

describe("errorResponse", () => {
  it("has spec-confirmed shape { code, content }", async () => {
    const response = errorResponse(400, "invalid_request", "Bad input");
    const body = await response.json();
    expect(body).toEqual({ code: "invalid_request", content: "Bad input" });
  });

  it("status code matches", () => {
    const response = errorResponse(400, "invalid_request", "Bad input");
    expect(response.status).toBe(400);
  });
});
