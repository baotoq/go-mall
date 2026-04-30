import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { createOrder, listOrders, getOrder, processPayment, updateOrderStatus } from "./orders-api"
import type { Address, PaymentMethod } from "./types"
import type { CartItem } from "@/store/cart"

// Provide a real localStorage implementation for tests (jsdom without URL gives a stub)
const localStorageStore: Record<string, string> = {}
const localStorageMock = {
  getItem: (key: string) => localStorageStore[key] ?? null,
  setItem: (key: string, value: string) => { localStorageStore[key] = value },
  removeItem: (key: string) => { delete localStorageStore[key] },
  clear: () => { Object.keys(localStorageStore).forEach((k) => delete localStorageStore[k]) },
  get length() { return Object.keys(localStorageStore).length },
  key: (i: number) => Object.keys(localStorageStore)[i] ?? null,
}
vi.stubGlobal("localStorage", localStorageMock)

const mockAddress: Address = {
  fullName: "Jane Doe",
  line1: "123 Main St",
  line2: "",
  city: "New York",
  state: "NY",
  postalCode: "10001",
  country: "US",
}

const mockPaymentCard: PaymentMethod = {
  type: "card",
  cardLastFour: "4242",
  cardBrand: "Visa",
}

const mockPaymentFail: PaymentMethod = {
  type: "card",
  cardLastFour: "0000",
  cardBrand: "Visa",
}

const mockCartItems: CartItem[] = [
  { id: "prod-1", name: "Widget", priceCents: 1000, imageUrl: "", quantity: 2 },
  { id: "prod-2", name: "Gadget", priceCents: 2500, imageUrl: "", quantity: 1 },
]

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

describe("createOrder", () => {
  it("creates an order and stores it in localStorage", async () => {
    const order = await createOrder({
      cartItems: mockCartItems,
      address: mockAddress,
      paymentMethod: mockPaymentCard,
      totalCents: 4500,
    })
    expect(order.id).toBeDefined()
    expect(order.status).toBe("pending")
    expect(order.items).toHaveLength(2)
    expect(order.totalCents).toBe(4500)
    expect(order.address.city).toBe("New York")
  })

  it("calculates subtotalCents per item", async () => {
    const order = await createOrder({
      cartItems: mockCartItems,
      address: mockAddress,
      paymentMethod: mockPaymentCard,
      totalCents: 4500,
    })
    expect(order.items[0].subtotalCents).toBe(2000)
    expect(order.items[1].subtotalCents).toBe(2500)
  })
})

describe("listOrders", () => {
  it("returns orders for the current session", async () => {
    await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentCard, totalCents: 4500 })
    await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentCard, totalCents: 4500 })
    const orders = await listOrders()
    expect(orders).toHaveLength(2)
  })

  it("returns empty array when no orders", async () => {
    const orders = await listOrders()
    expect(orders).toHaveLength(0)
  })

  it("returns newest orders first", async () => {
    const first = await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentCard, totalCents: 1000 })
    const second = await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentCard, totalCents: 2000 })
    const orders = await listOrders()
    expect(orders[0].id).toBe(second.id)
    expect(orders[1].id).toBe(first.id)
  })
})

describe("getOrder", () => {
  it("returns the order by id", async () => {
    const created = await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentCard, totalCents: 4500 })
    const found = await getOrder(created.id)
    expect(found?.id).toBe(created.id)
  })

  it("returns null for unknown id", async () => {
    const found = await getOrder("nonexistent-id")
    expect(found).toBeNull()
  })
})

describe("processPayment", () => {
  beforeEach(() => { vi.useFakeTimers() })
  afterEach(() => { vi.useRealTimers() })

  it("sets status to paid on success", async () => {
    const order = await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentCard, totalCents: 4500 })
    const promise = processPayment(order.id)
    await vi.runAllTimersAsync()
    const result = await promise
    expect(result.status).toBe("paid")
  })

  it("sets status to failed for card last four 0000", async () => {
    const order = await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentFail, totalCents: 4500 })
    const promise = processPayment(order.id)
    await vi.runAllTimersAsync()
    const result = await promise
    expect(result.status).toBe("failed")
  })

  it("throws for unknown order id", async () => {
    // attach rejection handler before advancing timers to avoid unhandled rejection
    const promise = processPayment("bad-id")
    promise.catch(() => {})
    await vi.runAllTimersAsync()
    await expect(promise).rejects.toThrow("bad-id")
  })
})

describe("updateOrderStatus", () => {
  it("updates the order status", async () => {
    const order = await createOrder({ cartItems: mockCartItems, address: mockAddress, paymentMethod: mockPaymentCard, totalCents: 4500 })
    const updated = await updateOrderStatus(order.id, "cancelled")
    expect(updated?.status).toBe("cancelled")
  })

  it("returns null for unknown id", async () => {
    const result = await updateOrderStatus("nope", "cancelled")
    expect(result).toBeNull()
  })
})
