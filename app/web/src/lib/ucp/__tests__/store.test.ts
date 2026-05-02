import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("ucp store", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it("store throws in prod env", async () => {
    // Arrange
    vi.stubEnv("NODE_ENV", "production");

    // Act & Assert
    await expect(import("../store")).rejects.toThrow(/production/);
  });

  it("get/set works in test env", async () => {
    // Arrange
    vi.stubEnv("NODE_ENV", "test");
    const { setSession, getSession } = await import("../store");
    const session = {
      id: "sess-1",
      status: "incomplete" as const,
      currency: "USD",
      cart_session_id: "cart-1",
      user_id: "user-1",
      totals: { subtotal_cents: 1000, currency: "USD" },
      messages: [],
      expires_at: new Date(Date.now() + 60_000).toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    // Act
    setSession("sess-1", session);
    const result = getSession("sess-1");

    // Assert
    expect(result).toEqual(session);
  });

  it("expired sessions return null", async () => {
    // Arrange
    vi.stubEnv("NODE_ENV", "test");
    const { setSession, getSession } = await import("../store");
    const expired = {
      id: "sess-expired",
      status: "incomplete" as const,
      currency: "USD",
      cart_session_id: "cart-1",
      user_id: "user-1",
      totals: { subtotal_cents: 0, currency: "USD" },
      messages: [],
      expires_at: new Date(Date.now() - 1000).toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    // Act
    setSession("sess-expired", expired);
    const result = getSession("sess-expired");

    // Assert
    expect(result).toBeNull();
  });

  it("setIdempotency/getIdempotency stores and retrieves entry", async () => {
    // Arrange
    vi.stubEnv("NODE_ENV", "test");
    const { setIdempotency, getIdempotency } = await import("../store");
    const entry = {
      response: { ok: true },
      hash: "abc123",
      expires_at: Date.now() + 60_000,
    };

    // Act
    setIdempotency("key-1", entry);
    const result = getIdempotency("key-1");

    // Assert
    expect(result).toEqual(entry);
  });

  it("expired idempotency entry returns null", async () => {
    // Arrange
    vi.stubEnv("NODE_ENV", "test");
    const { setIdempotency, getIdempotency } = await import("../store");
    const expired = {
      response: { ok: true },
      hash: "abc123",
      expires_at: Date.now() - 1000,
    };

    // Act
    setIdempotency("key-expired", expired);
    const result = getIdempotency("key-expired");

    // Assert
    expect(result).toBeNull();
  });
});
