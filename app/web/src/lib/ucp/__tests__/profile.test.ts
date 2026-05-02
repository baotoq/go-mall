import { describe, it, expect, afterEach, vi } from "vitest";

describe("generateProfile", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("returns all configured capabilities", async () => {
    // Arrange
    const { generateProfile } = await import("../profile");

    // Act
    const result = generateProfile();

    // Assert
    const capNames = result.ucp.capabilities.map((c) => c.name);
    expect(capNames).toContain("dev.ucp.shopping.checkout");
    expect(capNames).toContain("dev.ucp.shopping.fulfillment");
    expect(capNames).toContain("dev.ucp.shopping.discount");
    expect(capNames).toContain("dev.ucp.shopping.order");
  });

  it("service endpoints use UCP_DOMAIN env var", async () => {
    // Arrange
    vi.stubEnv("UCP_DOMAIN", "shop.example.com");
    const { generateProfile } = await import("../profile");

    // Act
    const result = generateProfile();

    // Assert
    const service = result.ucp.services["dev.ucp.shopping"];
    expect(service.rest?.endpoint).toContain("shop.example.com");
  });

  it("service endpoints default to localhost:3000", async () => {
    // Arrange
    vi.stubEnv("UCP_DOMAIN", "");
    const { generateProfile } = await import("../profile");

    // Act
    const result = generateProfile();

    // Assert
    const service = result.ucp.services["dev.ucp.shopping"];
    expect(service.rest?.endpoint).toContain("localhost:3000");
  });

  it("profile has correct ucp_version", async () => {
    // Arrange
    const { generateProfile } = await import("../profile");

    // Act
    const result = generateProfile();

    // Assert
    expect(result.ucp.version).toBe("2026-01-11");
  });
});
