import { describe, it, expect } from "vitest";

describe("checkout schemas", () => {
  describe("CreateCheckoutInputSchema", () => {
    it("valid CreateCheckoutInput passes", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = { cart_session_id: "sess123", currency: "USD" };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(true);
    });

    it("missing cart_session_id rejects", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = { currency: "USD" };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(false);
      if (!result.success) {
        const messages = result.error.issues
          .map((i) => i.path.join(".") + " " + i.message)
          .join(", ");
        expect(messages).toMatch(/cart_session_id/);
      }
    });

    it("currency must be 3 chars", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = { cart_session_id: "x", currency: "US" };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(false);
    });
  });

  describe("validateIdempotencyKey", () => {
    it("Idempotency-Key >64 chars rejects", async () => {
      // Arrange
      const { validateIdempotencyKey } = await import("../checkout");
      const longKey = "a".repeat(65);

      // Act
      const result = validateIdempotencyKey(longKey);

      // Assert
      expect(result.valid).toBe(false);
      expect(result.error).toBeDefined();
    });

    it("missing Idempotency-Key is valid", async () => {
      // Arrange
      const { validateIdempotencyKey } = await import("../checkout");

      // Act & Assert
      expect(validateIdempotencyKey(null).valid).toBe(true);
      expect(validateIdempotencyKey(undefined).valid).toBe(true);
    });

    it("malformed Idempotency-Key (special chars) rejects", async () => {
      // Arrange
      const { validateIdempotencyKey } = await import("../checkout");

      // Act
      const result = validateIdempotencyKey("key!@#");

      // Assert
      expect(result.valid).toBe(false);
      expect(result.error).toBeDefined();
    });
  });
});
