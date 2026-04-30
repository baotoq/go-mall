import { beforeEach, describe, expect, it } from "vitest"
import { useCartStore } from "@/store/cart"

const sampleItem = {
  id: "abc-123",
  name: "Wireless Headphones",
  priceCents: 7999,
  imageUrl: "",
}

beforeEach(() => {
  useCartStore.setState({ items: [] })
})

describe("addItem", () => {
  it("adds new item with quantity 1", () => {
    useCartStore.getState().addItem(sampleItem)
    expect(useCartStore.getState().items).toHaveLength(1)
    expect(useCartStore.getState().items[0].quantity).toBe(1)
  })

  it("increments quantity for existing item", () => {
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().addItem(sampleItem)
    expect(useCartStore.getState().items).toHaveLength(1)
    expect(useCartStore.getState().items[0].quantity).toBe(2)
  })
})

describe("removeItem", () => {
  it("removes item by id", () => {
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().removeItem(sampleItem.id)
    expect(useCartStore.getState().items).toHaveLength(0)
  })
})

describe("updateQuantity", () => {
  it("updates quantity", () => {
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().updateQuantity(sampleItem.id, 5)
    expect(useCartStore.getState().items[0].quantity).toBe(5)
  })

  it("removes item when quantity set to 0", () => {
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().updateQuantity(sampleItem.id, 0)
    expect(useCartStore.getState().items).toHaveLength(0)
  })
})

describe("totals", () => {
  it("totalItems sums quantities", () => {
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().addItem({ id: "def-456", name: "Watch", priceCents: 5000, imageUrl: "" })
    expect(useCartStore.getState().totalItems()).toBe(3)
  })

  it("totalPrice sums priceCents/100 × quantity", () => {
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().addItem(sampleItem)
    expect(useCartStore.getState().totalPrice()).toBeCloseTo(159.98)
  })
})

describe("clearCart", () => {
  it("empties the cart", () => {
    useCartStore.getState().addItem(sampleItem)
    useCartStore.getState().clearCart()
    expect(useCartStore.getState().items).toHaveLength(0)
  })
})
