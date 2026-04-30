import { render, screen, fireEvent, waitFor } from "@testing-library/react"
import { describe, it, expect, beforeEach, vi } from "vitest"

// Provide localStorage stub for jsdom environment (no URL set)
const localStorageStore: Record<string, string> = {}
vi.stubGlobal("localStorage", {
  getItem: (key: string) => localStorageStore[key] ?? null,
  setItem: (key: string, value: string) => { localStorageStore[key] = value },
  removeItem: (key: string) => { delete localStorageStore[key] },
  clear: () => { Object.keys(localStorageStore).forEach((k) => delete localStorageStore[k]) },
  get length() { return Object.keys(localStorageStore).length },
  key: (i: number) => Object.keys(localStorageStore)[i] ?? null,
})
import { CheckoutClient } from "./checkout-client"
import { useCheckoutStore } from "@/store/checkout"
import { useCartStore } from "@/store/cart"

// Mock cart API to prevent localStorage errors from cart store's loadCart
vi.mock("@/lib/api", () => ({
  getCart: vi.fn().mockResolvedValue(null),
  addCartItem: vi.fn().mockResolvedValue(null),
  updateCartItem: vi.fn().mockResolvedValue(null),
  removeCartItem: vi.fn().mockResolvedValue(null),
  clearCart: vi.fn().mockResolvedValue(undefined),
}))

// Mock orders API
vi.mock("@/lib/orders-api", () => ({
  createOrder: vi.fn().mockResolvedValue({
    id: "order-test-123",
    sessionId: "sess-1",
    items: [],
    address: {},
    paymentMethod: { type: "card", cardLastFour: "4242" },
    status: "pending",
    totalCents: 4500,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }),
  processPayment: vi.fn().mockResolvedValue({
    id: "order-test-123",
    sessionId: "sess-1",
    items: [],
    address: {},
    paymentMethod: { type: "card", cardLastFour: "4242" },
    status: "paid",
    totalCents: 4500,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }),
}))

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn() }),
  usePathname: () => "/checkout",
  useSearchParams: () => new URLSearchParams(),
}))

const mockAddress = {
  fullName: "Jane Doe",
  line1: "123 Main St",
  line2: "",
  city: "New York",
  state: "NY",
  postalCode: "10001",
  country: "US",
}

beforeEach(() => {
  useCheckoutStore.setState({
    step: "address",
    address: null,
    paymentMethod: null,
    order: null,
    errorMessage: null,
  })
  useCartStore.setState({
    items: [{ id: "p1", name: "Widget", priceCents: 1000, imageUrl: "", quantity: 2 }],
  })
})

describe("CheckoutClient", () => {
  it("shows address form on initial render", () => {
    render(<CheckoutClient />)
    expect(screen.getByTestId("address-form")).toBeDefined()
  })

  it("shows empty cart message when cart is empty", () => {
    useCartStore.setState({ items: [] })
    render(<CheckoutClient />)
    expect(screen.getByText(/your cart is empty/i)).toBeDefined()
  })

  it("advances to payment step after address submit", async () => {
    render(<CheckoutClient />)

    fireEvent.change(screen.getByLabelText(/full name/i), { target: { value: mockAddress.fullName } })
    fireEvent.change(screen.getByLabelText(/address line 1/i), { target: { value: mockAddress.line1 } })
    fireEvent.change(screen.getByLabelText(/address line 2/i), { target: { value: mockAddress.line2 } })
    fireEvent.change(screen.getByLabelText(/city/i), { target: { value: mockAddress.city } })
    fireEvent.change(screen.getByLabelText(/state/i), { target: { value: mockAddress.state } })
    fireEvent.change(screen.getByLabelText(/postal code/i), { target: { value: mockAddress.postalCode } })
    fireEvent.change(screen.getByLabelText(/country/i), { target: { value: mockAddress.country } })

    fireEvent.click(screen.getByText(/continue to payment/i))

    await waitFor(() => {
      expect(screen.getByTestId("payment-form")).toBeDefined()
    })
  })

  it("shows review step when address and payment are set in store", () => {
    useCheckoutStore.setState({
      step: "review",
      address: mockAddress,
      paymentMethod: { type: "card", cardLastFour: "4242" },
      order: null,
      errorMessage: null,
    })
    render(<CheckoutClient />)
    expect(screen.getByTestId("review-step")).toBeDefined()
    expect(screen.getByText(/jane doe/i)).toBeDefined()
  })

  it("shows processing state when step is processing", () => {
    useCheckoutStore.setState({
      step: "processing",
      address: mockAddress,
      paymentMethod: { type: "card", cardLastFour: "4242" },
      order: null,
      errorMessage: null,
    })
    render(<CheckoutClient />)
    expect(screen.getByTestId("payment-processing")).toBeDefined()
  })

  it("shows error state when errorMessage is set", () => {
    useCheckoutStore.setState({
      step: "error",
      address: mockAddress,
      paymentMethod: { type: "card", cardLastFour: "4242" },
      order: null,
      errorMessage: "Payment declined",
    })
    render(<CheckoutClient />)
    expect(screen.getByTestId("payment-error")).toBeDefined()
    expect(screen.getByText(/payment declined/i)).toBeDefined()
  })
})
