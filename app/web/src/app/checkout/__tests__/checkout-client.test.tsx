import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useCartStore } from "@/store/cart";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush, replace: vi.fn(), back: vi.fn() }),
  usePathname: () => "/checkout",
  useSearchParams: () => new URLSearchParams(),
}));

const { CheckoutClient } = await import("../checkout-client");

const CART_ITEMS = [
  {
    id: "prod-1",
    name: "Test Widget",
    priceCents: 2999,
    imageUrl: "",
    quantity: 2,
  },
];

function seedCart() {
  useCartStore.setState({
    items: CART_ITEMS,
    isLoading: false,
    loadCart: async () => {},
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockPush.mockReset();
  vi.stubGlobal("fetch", vi.fn());
});

describe("CheckoutClient", () => {
  it("renders contact, shipping, and payment fields", async () => {
    seedCart();
    render(<CheckoutClient defaultEmail="user@example.com" />);

    expect(screen.getByLabelText(/email/i)).toBeDefined();
    expect(screen.getByLabelText(/^name/i)).toBeDefined();
    expect(screen.getByLabelText(/address line 1/i)).toBeDefined();
    expect(screen.getByLabelText(/city/i)).toBeDefined();
    expect(screen.getByLabelText(/postal code/i)).toBeDefined();
    expect(screen.getByLabelText(/country/i)).toBeDefined();
    expect(screen.getByLabelText(/card number/i)).toBeDefined();
    expect(screen.getByLabelText(/expiry/i)).toBeDefined();
    expect(screen.getByLabelText(/cvc/i)).toBeDefined();
  });

  it("pre-fills email from defaultEmail prop", async () => {
    seedCart();
    render(<CheckoutClient defaultEmail="prefilled@example.com" />);
    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    expect(emailInput.value).toBe("prefilled@example.com");
  });

  it("shows validation errors on empty submit", async () => {
    seedCart();
    render(<CheckoutClient defaultEmail="" />);

    fireEvent.click(screen.getByRole("button", { name: /place order/i }));

    await waitFor(() => {
      const errors = document.querySelectorAll("[data-slot='form-message']");
      expect(errors.length).toBeGreaterThan(0);
    });
  });

  it("redirects to /cart when cart is empty after loading", async () => {
    useCartStore.setState({
      items: [],
      isLoading: false,
      loadCart: async () => {},
    });
    render(<CheckoutClient defaultEmail="" />);

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/cart");
    });
  });

  it("does not redirect while cart is still loading", () => {
    useCartStore.setState({
      items: [],
      isLoading: true,
      loadCart: async () => {},
    });
    render(<CheckoutClient defaultEmail="" />);
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("happy path: submits form and redirects to success page", async () => {
    seedCart();

    const mockFetch = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({
          id: "sess-abc123",
          session_id: "sess-abc123",
          status: "ready_for_complete",
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ id: "sess-abc123", status: "completed" }),
      });
    vi.stubGlobal("fetch", mockFetch);

    render(<CheckoutClient defaultEmail="user@test.com" />);

    const user = userEvent.setup();

    await user.clear(screen.getByLabelText(/email/i));
    await user.type(screen.getByLabelText(/email/i), "user@test.com");
    await user.type(screen.getByLabelText(/^name/i), "Jane Doe");
    await user.type(screen.getByLabelText(/address line 1/i), "123 Main St");
    await user.type(screen.getByLabelText(/city/i), "Springfield");
    await user.type(screen.getByLabelText(/state/i), "IL");
    await user.type(screen.getByLabelText(/postal code/i), "62701");
    await user.type(screen.getByLabelText(/country/i), "US");
    await user.type(
      screen.getByLabelText(/card number/i),
      "4242 4242 4242 4242",
    );
    await user.type(screen.getByLabelText(/expiry/i), "12/26");
    await user.type(screen.getByLabelText(/cvc/i), "123");

    await user.click(screen.getByRole("button", { name: /place order/i }));

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith(
        "/checkout/success?id=sess-abc123",
      );
    });
    expect(mockFetch).toHaveBeenCalledTimes(2);
  });

  it("attaches Idempotency-Key header on POST", async () => {
    seedCart();

    const mockFetch = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({
          id: "sess-xyz",
          session_id: "sess-xyz",
          status: "ready_for_complete",
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ id: "sess-xyz", status: "completed" }),
      });
    vi.stubGlobal("fetch", mockFetch);

    render(<CheckoutClient defaultEmail="user@test.com" />);
    const user = userEvent.setup();

    await user.clear(screen.getByLabelText(/email/i));
    await user.type(screen.getByLabelText(/email/i), "user@test.com");
    await user.type(screen.getByLabelText(/^name/i), "Jane Doe");
    await user.type(screen.getByLabelText(/address line 1/i), "123 Main St");
    await user.type(screen.getByLabelText(/city/i), "Springfield");
    await user.type(screen.getByLabelText(/state/i), "IL");
    await user.type(screen.getByLabelText(/postal code/i), "62701");
    await user.type(screen.getByLabelText(/country/i), "US");
    await user.type(
      screen.getByLabelText(/card number/i),
      "4242 4242 4242 4242",
    );
    await user.type(screen.getByLabelText(/expiry/i), "12/26");
    await user.type(screen.getByLabelText(/cvc/i), "123");
    await user.click(screen.getByRole("button", { name: /place order/i }));

    await waitFor(() => expect(mockFetch).toHaveBeenCalled());
    const [_url, init] = mockFetch.mock.calls[0] as [string, RequestInit];
    const headers = init.headers as Record<string, string>;
    expect(headers["Idempotency-Key"]).toBeDefined();
    expect(headers["Idempotency-Key"].length).toBeGreaterThan(0);
  });

  it("shows error banner and re-enables button on 502", async () => {
    seedCart();

    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValueOnce({
        ok: false,
        status: 502,
        json: async () => ({
          error: { code: "upstream_error", content: "Gateway timeout" },
        }),
      }),
    );

    render(<CheckoutClient defaultEmail="user@test.com" />);
    const user = userEvent.setup();

    await user.clear(screen.getByLabelText(/email/i));
    await user.type(screen.getByLabelText(/email/i), "user@test.com");
    await user.type(screen.getByLabelText(/^name/i), "Jane Doe");
    await user.type(screen.getByLabelText(/address line 1/i), "123 Main St");
    await user.type(screen.getByLabelText(/city/i), "Springfield");
    await user.type(screen.getByLabelText(/state/i), "IL");
    await user.type(screen.getByLabelText(/postal code/i), "62701");
    await user.type(screen.getByLabelText(/country/i), "US");
    await user.type(
      screen.getByLabelText(/card number/i),
      "4242 4242 4242 4242",
    );
    await user.type(screen.getByLabelText(/expiry/i), "12/26");
    await user.type(screen.getByLabelText(/cvc/i), "123");
    await user.click(screen.getByRole("button", { name: /place order/i }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeDefined();
    });
    const btn = screen.getByRole("button", {
      name: /place order/i,
    }) as HTMLButtonElement;
    expect(btn.disabled).toBe(false);
  });
});
