import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useCartStore } from "@/store/cart";
import type { CheckoutSession } from "@/lib/ucp/types/checkout";

const { SuccessClient } = await import("../success-client");

const MOCK_SESSION: CheckoutSession = {
  id: "12345678-1234-4abc-8def-abcdef012345",
  status: "completed",
  currency: "USD",
  cart_session_id: "cart-abc",
  user_id: "user-1",
  buyer: { email: "jane@example.com", name: "Jane Doe" },
  shipping_address: {
    line1: "123 Main St",
    city: "Springfield",
    state: "IL",
    postal_code: "62701",
    country: "US",
  },
  payment: { brand: "visa", last4: "4242" },
  totals: { subtotal_cents: 5998, currency: "USD" },
  messages: [],
  expires_at: new Date(Date.now() + 3600000).toISOString(),
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

beforeEach(() => {
  useCartStore.setState({
    items: [
      {
        id: "prod-1",
        name: "Test Widget",
        priceCents: 2999,
        imageUrl: "",
        quantity: 2,
      },
    ],
    isLoading: false,
    clearCart: vi.fn(),
  });
});

describe("SuccessClient", () => {
  it("renders order id in ord_XXXXXXXX format", () => {
    render(<SuccessClient session={MOCK_SESSION} />);
    expect(screen.getByText(/ord_12345678/i)).toBeDefined();
  });

  it("renders shipping address recap", () => {
    render(<SuccessClient session={MOCK_SESSION} />);
    expect(screen.getByText(/123 Main St/i)).toBeDefined();
    expect(screen.getByText(/Springfield/i)).toBeDefined();
  });

  it("renders total from session totals", () => {
    render(<SuccessClient session={MOCK_SESSION} />);
    // subtotal 59.98 + shipping 5.99 = 65.97
    expect(screen.getByText(/\$65\.97/)).toBeDefined();
  });

  it("calls clearCart on mount", () => {
    const clearCart = vi.fn();
    useCartStore.setState({ clearCart });
    render(<SuccessClient session={MOCK_SESSION} />);
    expect(clearCart).toHaveBeenCalledTimes(1);
  });

  it("renders empty state with Continue shopping link when session is null", () => {
    render(<SuccessClient session={null} />);
    expect(screen.getByText(/continue shopping/i)).toBeDefined();
  });

  it("does not render order id when session is null", () => {
    render(<SuccessClient session={null} />);
    expect(screen.queryByText(/ord_/)).toBeNull();
  });
});
