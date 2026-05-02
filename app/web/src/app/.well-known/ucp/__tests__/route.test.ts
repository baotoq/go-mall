import { describe, expect, it } from "vitest";
import { GET } from "../route";

describe("GET /.well-known/ucp", () => {
  it("GET returns 200 JSON", async () => {
    // Arrange & Act
    const response = await GET();

    // Assert
    expect(response.status).toBe(200);
  });

  it("Cache-Control header is public, max-age=3600", async () => {
    // Arrange & Act
    const response = await GET();

    // Assert
    expect(response.headers.get("Cache-Control")).toBe("public, max-age=3600");
  });

  it("Vary: Accept header is set", async () => {
    // Arrange & Act
    const response = await GET();

    // Assert
    expect(response.headers.get("Vary")).toBe("Accept");
  });

  it("response body has ucp.capabilities array", async () => {
    // Arrange & Act
    const response = await GET();
    const body = await response.json();

    // Assert
    expect(Array.isArray(body.ucp.capabilities)).toBe(true);
  });
});
