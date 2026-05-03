import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CartData } from "@/lib/types";

vi.mock("@/lib/api", () => ({
  getCart: vi.fn(),
}));

vi.mock("../order", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../order")>();
  return {
    ...actual,
    createOrder: vi.fn(),
  };
});

import { getCart } from "@/lib/api";
import {
  completeCheckout,
  createCheckout,
  getCheckoutSession,
  updateCheckout,
} from "../checkout";
import { createOrder } from "../order";

const mockedGetCart = vi.mocked(getCart);
const mockedCreateOrder = vi.mocked(createOrder);

function makeCart(overrides: Partial<CartData> = {}): CartData {
  return {
    id: "cart-1",
    sessionId: "sess-1",
    items: [
      {
        id: "item-1",
        productId: "prod-1",
        name: "Widget",
        priceCents: 1000,
        currency: "USD",
        imageUrl: "https://example.com/w.png",
        quantity: 2,
        subtotalCents: 2000,
      },
    ],
    totalCents: 2000,
    ...overrides,
  };
}

describe("createCheckout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns 400 empty_cart when cart is null", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(null);

    // Act
    const result = await createCheckout(
      { cart_session_id: "missing", currency: "USD" },
      {},
    );

    // Assert
    expect(result).toMatchObject({
      error: true,
      status: 400,
      code: "empty_cart",
    });
  });

  it("returns 400 empty_cart when cart has no items", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart({ items: [], totalCents: 0 }));

    // Act
    const result = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      {},
    );

    // Assert
    expect(result).toMatchObject({
      error: true,
      status: 400,
      code: "empty_cart",
    });
  });

  it("maps cart correctly into a checkout session", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());

    // Act
    const result = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      { userId: "user-1" },
    );

    // Assert
    if ("error" in result) throw new Error("expected session");
    expect(result.session.cart_session_id).toBe("sess-1");
    expect(result.session.totals.subtotal_cents).toBe(2000);
    expect(result.session.status).toBe("incomplete");
    expect(result.session.user_id).toBe("user-1");
    expect(
      result.session.messages.some((m) => m.code === "missing_email"),
    ).toBe(true);
  });

  it("uses 'guest' as user_id when not provided", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());

    // Act
    const result = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      {},
    );

    // Assert
    if ("error" in result) throw new Error("expected session");
    expect(result.session.user_id).toBe("guest");
  });

  it("returns 400 currency_mismatch when cart items currency differs", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(
      makeCart({
        items: [
          {
            id: "item-1",
            productId: "prod-1",
            name: "Widget",
            priceCents: 1000,
            currency: "EUR",
            imageUrl: "",
            quantity: 1,
            subtotalCents: 1000,
          },
        ],
      }),
    );

    // Act
    const result = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      {},
    );

    // Assert
    expect(result).toMatchObject({
      error: true,
      status: 400,
      code: "currency_mismatch",
    });
  });
});

describe("updateCheckout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("stays incomplete when only buyer.email is provided (missing shipping+payment)", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());
    const created = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      {},
    );
    if ("error" in created) throw new Error("expected session");

    // Act
    const updated = await updateCheckout(created.session.id, {
      buyer: { email: "buyer@example.com" },
    });

    // Assert
    expect(updated).not.toBeNull();
    expect(updated?.status).toBe("incomplete");
    expect(updated?.buyer?.email).toBe("buyer@example.com");
    const codes = updated?.messages.map((m) => m.code) ?? [];
    expect(codes).toContain("missing_shipping_address");
    expect(codes).toContain("missing_payment");
  });

  it("update from incomplete → ready_for_complete when all fields supplied", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());
    const created = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      {},
    );
    if ("error" in created) throw new Error("expected session");
    expect(created.session.status).toBe("incomplete");

    // Act
    const updated = await updateCheckout(created.session.id, {
      buyer: { email: "buyer@example.com", name: "Alice" },
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
    });

    // Assert
    expect(updated).not.toBeNull();
    expect(updated?.status).toBe("ready_for_complete");
    expect(updated?.messages).toHaveLength(0);
    expect(updated?.payment).toEqual({ brand: "visa", last4: "4242" });
    expect(updated?.shipping_address?.city).toBe("Springfield");
  });

  it("raw card number is NOT present anywhere in JSON.stringify(session) after update", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());
    const created = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      {},
    );
    if ("error" in created) throw new Error("expected session");

    // Act
    const updated = await updateCheckout(created.session.id, {
      buyer: { email: "buyer@example.com" },
      payment: {
        card_number: "4242 4242 4242 4242",
        exp: "12/26",
        cvc: "123",
      },
    });

    // Assert — no 16-digit string in the serialized session
    const json = JSON.stringify(updated);
    expect(json).not.toMatch(/\b\d{16}\b/);
    expect(json).not.toMatch(/4242424242424242/);
  });
});

describe("createCheckout with full buyer+shipping+payment", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("create with full body → status ready_for_complete, payment redacted to {brand, last4}", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());

    // Act
    const result = await createCheckout(
      {
        cart_session_id: "sess-1",
        currency: "USD",
        buyer: { email: "buyer@example.com", name: "Alice" },
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
      },
      { userId: "user-1" },
    );

    // Assert
    if ("error" in result) throw new Error("expected session");
    expect(result.session.status).toBe("ready_for_complete");
    expect(result.session.messages).toHaveLength(0);
    expect(result.session.payment).toEqual({ brand: "visa", last4: "4242" });
    expect(result.session.buyer?.email).toBe("buyer@example.com");
  });

  it("create with email only → status incomplete, messages contain missing_shipping_address and missing_payment", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());

    // Act
    const result = await createCheckout(
      {
        cart_session_id: "sess-1",
        currency: "USD",
        buyer: { email: "buyer@example.com" },
      },
      {},
    );

    // Assert
    if ("error" in result) throw new Error("expected session");
    expect(result.session.status).toBe("incomplete");
    const codes = result.session.messages.map((m) => m.code);
    expect(codes).toContain("missing_shipping_address");
    expect(codes).toContain("missing_payment");
    expect(codes).not.toContain("missing_email");
  });

  it("raw card number is NOT present in JSON.stringify(session) after create", async () => {
    // Arrange
    mockedGetCart.mockResolvedValue(makeCart());

    // Act
    const result = await createCheckout(
      {
        cart_session_id: "sess-1",
        currency: "USD",
        buyer: { email: "buyer@example.com" },
        payment: {
          card_number: "4111 1111 1111 1111",
          exp: "01/27",
          cvc: "456",
        },
      },
      {},
    );

    // Assert
    if ("error" in result) throw new Error("expected session");
    const json = JSON.stringify(result.session);
    expect(json).not.toMatch(/\b\d{16}\b/);
    expect(json).not.toMatch(/4111111111111111/);
  });
});

describe("completeCheckout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  async function setupReadySession() {
    mockedGetCart.mockResolvedValue(makeCart());
    const created = await createCheckout(
      { cart_session_id: "sess-1", currency: "USD" },
      {},
    );
    if ("error" in created) throw new Error("expected session");
    const updated = await updateCheckout(created.session.id, {
      buyer: { email: "buyer@example.com" },
      shipping_address: {
        line1: "1 Test St",
        city: "Testville",
        state: "TX",
        postal_code: "12345",
        country: "US",
      },
      payment: {
        card_number: "4242 4242 4242 4242",
        exp: "12/26",
        cvc: "123",
      },
    });
    if (!updated) throw new Error("expected updated session");
    return updated;
  }

  it("transitions to completed and calls createOrder with correct args", async () => {
    // Arrange
    const session = await setupReadySession();
    mockedCreateOrder.mockResolvedValue({ id: "order-123" });

    // Act
    const result = await completeCheckout(session.id, {});

    // Assert
    if ("error" in result) throw new Error("expected session");
    expect(result.session.status).toBe("completed");
    expect(mockedCreateOrder).toHaveBeenCalledTimes(1);
    const [passedSession, passedItems] = mockedCreateOrder.mock.calls[0];
    expect(passedSession.id).toBe(session.id);
    expect(passedSession.cart_session_id).toBe("sess-1");
    expect(passedItems).toHaveLength(1);
    expect(passedItems[0].productId).toBe("prod-1");
  });

  it("returns 502 and reverts status when order creation fails", async () => {
    // Arrange
    const session = await setupReadySession();
    mockedCreateOrder.mockRejectedValue(new Error("boom"));

    // Act
    const result = await completeCheckout(session.id, {});

    // Assert
    expect(result).toMatchObject({
      error: true,
      status: 502,
      code: "order_creation_failed",
    });
    const persisted = await getCheckoutSession(session.id);
    expect(persisted?.status).toBe("ready_for_complete");
  });

  it("returns 409 on concurrent double-complete", async () => {
    // Arrange
    const session = await setupReadySession();
    let resolveOrder: (v: { id: string }) => void = () => {};
    mockedCreateOrder.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveOrder = resolve;
        }),
    );

    // Act
    const p1 = completeCheckout(session.id, {});
    const p2 = completeCheckout(session.id, {});
    resolveOrder({ id: "order-123" });
    const [r1, r2] = await Promise.all([p1, p2]);

    // Assert
    const statuses = [
      "error" in r1 ? r1.status : 200,
      "error" in r2 ? r2.status : 200,
    ];
    expect(statuses).toContain(409);
  });

  it("returns cached result for same Idempotency-Key", async () => {
    // Arrange
    const session = await setupReadySession();
    mockedCreateOrder.mockResolvedValue({ id: "order-123" });
    const key = "idem-key-1";

    // Act
    const r1 = await completeCheckout(session.id, { idempotencyKey: key });
    const r2 = await completeCheckout(session.id, { idempotencyKey: key });

    // Assert
    if ("error" in r1) throw new Error("expected session");
    if ("error" in r2) throw new Error("expected session");
    expect(r1.session.id).toBe(r2.session.id);
    expect(mockedCreateOrder).toHaveBeenCalledTimes(1);
  });

  it("idempotency replay returns frozen snapshot, immune to later mutations", async () => {
    // Arrange — complete once with a key, capture the snapshot values.
    const session = await setupReadySession();
    mockedCreateOrder.mockResolvedValue({ id: "order-123" });
    const key = "idem-key-frozen";
    const r1 = await completeCheckout(session.id, { idempotencyKey: key });
    if ("error" in r1) throw new Error("expected session");
    const cachedAt = r1.session.updated_at;
    const cachedEmail = r1.session.buyer?.email;
    const cachedStatus = r1.session.status;

    // Mutate the live session directly — bypassing the public API so the
    // change is observable regardless of timer/clock resolution.
    const live = await getCheckoutSession(session.id);
    if (!live) throw new Error("expected live session");
    live.updated_at = "MUTATED-AFTER-CACHE";
    live.buyer = { email: "tampered@evil.example" };
    live.status = "incomplete";

    // Act — replay the original idempotency key.
    const r2 = await completeCheckout(session.id, { idempotencyKey: key });

    // Assert — replayed snapshot is the original, NOT the mutated live state.
    if ("error" in r2) throw new Error("expected session");
    expect(r2.session.updated_at).toBe(cachedAt);
    expect(r2.session.buyer?.email).toBe(cachedEmail);
    expect(r2.session.status).toBe(cachedStatus);
    // r1 and r2 must be independent objects so callers can't poison cache.
    expect(r1.session).not.toBe(r2.session);
  });

  it("logs the underlying error when createOrder fails", async () => {
    // Arrange
    const session = await setupReadySession();
    const upstreamErr = new Error("upstream 500: kratos error_reason=BOOM");
    mockedCreateOrder.mockRejectedValue(upstreamErr);
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    // Act
    await completeCheckout(session.id, {});

    // Assert
    expect(errSpy).toHaveBeenCalledWith(
      "[ucp] createOrder failed:",
      upstreamErr,
    );
    errSpy.mockRestore();
  });
});
