import { describe, expect, it } from "vitest";

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
          .map((i) => `${i.path.join(".")} ${i.message}`)
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

    it("valid create input with full buyer + shipping + payment passes", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = {
        cart_session_id: "sess123",
        currency: "USD",
        buyer: { email: "buyer@example.com", name: "Alice", phone: "+1555" },
        shipping_address: {
          line1: "123 Main St",
          city: "Springfield",
          state: "IL",
          postal_code: "62701",
          country: "US",
        },
        payment: {
          card_number: "4242 4242 4242 4242",
          exp: "12/26",
          cvc: "123",
        },
      };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(true);
    });

    it("create input with only cart_session_id + currency still passes (back-compat)", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = { cart_session_id: "sess123", currency: "USD" };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(true);
    });

    it("invalid card_number fails", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = {
        cart_session_id: "sess123",
        currency: "USD",
        payment: {
          card_number: "abc",
          exp: "12/26",
          cvc: "123",
        },
      };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(false);
    });

    it("invalid exp format fails", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = {
        cart_session_id: "sess123",
        currency: "USD",
        payment: {
          card_number: "4242 4242 4242 4242",
          exp: "13/26",
          cvc: "123",
        },
      };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(false);
    });

    it("invalid cvc fails", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = {
        cart_session_id: "sess123",
        currency: "USD",
        payment: {
          card_number: "4242 4242 4242 4242",
          exp: "12/26",
          cvc: "12",
        },
      };

      // Act
      const result = CreateCheckoutInputSchema.safeParse(input);

      // Assert
      expect(result.success).toBe(false);
    });

    it("shipping_address country must be 2 chars", async () => {
      // Arrange
      const { CreateCheckoutInputSchema } = await import("../checkout");
      const input = {
        cart_session_id: "sess123",
        currency: "USD",
        shipping_address: {
          line1: "123 Main St",
          city: "Springfield",
          state: "IL",
          postal_code: "62701",
          country: "USA",
        },
      };

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

  describe("summarizePayment", () => {
    it("visa card returns correct brand and last4", async () => {
      // Arrange
      const { summarizePayment } = await import("../checkout");

      // Act
      const result = summarizePayment({ card_number: "4242 4242 4242 4242" });

      // Assert
      expect(result).toEqual({ brand: "visa", last4: "4242" });
    });

    it("mastercard returns correct brand", async () => {
      // Arrange
      const { summarizePayment } = await import("../checkout");

      // Act
      const result = summarizePayment({ card_number: "5555 5555 5555 4444" });

      // Assert
      expect(result.brand).toBe("mastercard");
      expect(result.last4).toBe("4444");
    });

    it("amex returns correct brand", async () => {
      // Arrange
      const { summarizePayment } = await import("../checkout");

      // Act
      const result = summarizePayment({ card_number: "3714 496353 98431" });

      // Assert
      expect(result.brand).toBe("amex");
      expect(result.last4).toBe("8431");
    });

    it("unknown card returns 'unknown' brand", async () => {
      // Arrange
      const { summarizePayment } = await import("../checkout");

      // Act
      const result = summarizePayment({ card_number: "6011 1111 1111 1117" });

      // Assert
      expect(result.brand).toBe("unknown");
      expect(result.last4).toBe("1117");
    });
  });
});
