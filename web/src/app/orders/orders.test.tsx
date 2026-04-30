import { render, screen, waitFor } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { OrdersClient } from "./orders-client"
import { OrderDetailClient } from "./[id]/order-detail-client"
import type { Order } from "@/lib/types"

vi.mock("@/lib/orders-api", () => ({
  listOrders: vi.fn(),
  getOrder: vi.fn(),
}))

import { listOrders, getOrder } from "@/lib/orders-api"

const mockOrder: Order = {
  id: "order-uuid-1234-5678",
  sessionId: "session-abc",
  items: [
    {
      productId: "prod-1",
      name: "Wireless Headphones",
      priceCents: 7999,
      currency: "USD",
      imageUrl: "",
      quantity: 2,
      subtotalCents: 15998,
    },
  ],
  address: {
    fullName: "Jane Doe",
    line1: "123 Main St",
    line2: "",
    city: "Springfield",
    state: "IL",
    postalCode: "62701",
    country: "US",
  },
  paymentMethod: { type: "card", cardLastFour: "4242", cardBrand: "Visa" },
  status: "paid",
  totalCents: 15998,
  createdAt: new Date("2025-01-15T10:00:00Z").toISOString(),
  updatedAt: new Date("2025-01-15T10:01:00Z").toISOString(),
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("OrdersClient", () => {
  it("shows empty state when no orders", async () => {
    vi.mocked(listOrders).mockResolvedValue([])
    render(<OrdersClient />)
    await waitFor(() => expect(screen.getByText("No orders yet")).toBeDefined())
    expect(screen.getByText(/place your first order/i)).toBeDefined()
  })

  it("renders order rows with status and total", async () => {
    vi.mocked(listOrders).mockResolvedValue([mockOrder])
    render(<OrdersClient />)
    await waitFor(() => expect(screen.getByTestId("order-status")).toBeDefined())
    expect(screen.getByTestId("order-status").textContent).toBe("Paid")
    expect(screen.getByText("$159.98")).toBeDefined()
  })

  it("links each row to the order detail page", async () => {
    vi.mocked(listOrders).mockResolvedValue([mockOrder])
    render(<OrdersClient />)
    await waitFor(() => expect(screen.getByTestId("order-status")).toBeDefined())
    const links = screen.getAllByRole("link") as HTMLAnchorElement[]
    const orderLink = links.find((l) => l.href.includes(`/orders/${mockOrder.id}`))
    expect(orderLink).toBeDefined()
  })

  it("shows loading spinner initially", () => {
    vi.mocked(listOrders).mockReturnValue(new Promise(() => {}))
    render(<OrdersClient />)
    expect(screen.getByLabelText("Loading")).toBeDefined()
  })
})

describe("OrderDetailClient", () => {
  it("shows order not found for missing order", async () => {
    vi.mocked(getOrder).mockResolvedValue(null)
    render(<OrderDetailClient id="nonexistent" />)
    await waitFor(() => expect(screen.getByText("Order not found")).toBeDefined())
  })

  it("shows confirmed status for paid order", async () => {
    vi.mocked(getOrder).mockResolvedValue(mockOrder)
    render(<OrderDetailClient id={mockOrder.id} />)
    await waitFor(() => expect(screen.getByTestId("order-status-heading")).toBeDefined())
    expect(screen.getByTestId("order-status-heading").textContent).toBe("Order Confirmed")
  })

  it("renders item list and total", async () => {
    vi.mocked(getOrder).mockResolvedValue(mockOrder)
    render(<OrderDetailClient id={mockOrder.id} />)
    await waitFor(() => expect(screen.getByText("Wireless Headphones")).toBeDefined())
    expect(screen.getByTestId("order-total").textContent).toBe("$159.98")
  })

  it("renders shipping address", async () => {
    vi.mocked(getOrder).mockResolvedValue(mockOrder)
    render(<OrderDetailClient id={mockOrder.id} />)
    await waitFor(() => expect(screen.getByText("Jane Doe")).toBeDefined())
    expect(screen.getByText("123 Main St")).toBeDefined()
  })

  it("renders payment method", async () => {
    vi.mocked(getOrder).mockResolvedValue(mockOrder)
    render(<OrderDetailClient id={mockOrder.id} />)
    await waitFor(() => expect(screen.getByText(/visa ending in 4242/i)).toBeDefined())
  })

  it("shows failed status for failed order", async () => {
    vi.mocked(getOrder).mockResolvedValue({ ...mockOrder, status: "failed" })
    render(<OrderDetailClient id={mockOrder.id} />)
    await waitFor(() => expect(screen.getByTestId("order-status-heading")).toBeDefined())
    expect(screen.getByTestId("order-status-heading").textContent).toBe("Payment Failed")
  })

  it("shows loading spinner initially", () => {
    vi.mocked(getOrder).mockReturnValue(new Promise(() => {}))
    render(<OrderDetailClient id={mockOrder.id} />)
    expect(screen.getByLabelText("Loading")).toBeDefined()
  })
})
