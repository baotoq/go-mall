import React from "react"
import { render, screen, fireEvent } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { ProductCard } from "@/components/product-card"
import { useCartStore } from "@/store/cart"

const product = {
  id: 1,
  name: "Wireless Headphones",
  price: 79.99,
  category: "electronics",
  description: "Great headphones",
  emoji: "🎧",
  rating: 4.5,
  reviews: 128,
  badge: "Best Seller",
}

beforeEach(() => {
  useCartStore.setState({ items: [] })
})

describe("ProductCard", () => {
  it("renders product name and price", () => {
    render(<ProductCard product={product} />)
    expect(screen.getByText("Wireless Headphones")).toBeDefined()
    expect(screen.getByText("$79.99")).toBeDefined()
  })

  it("renders badge when present", () => {
    render(<ProductCard product={product} />)
    expect(screen.getByText("Best Seller")).toBeDefined()
  })

  it("does not render badge when absent", () => {
    const noBadge = { ...product, badge: undefined }
    render(<ProductCard product={noBadge} />)
    expect(screen.queryByText("Best Seller")).toBeNull()
  })

  it("adds item to cart on click", () => {
    render(<ProductCard product={product} />)
    fireEvent.click(screen.getByRole("button", { name: /add/i }))
    expect(useCartStore.getState().items).toHaveLength(1)
    expect(useCartStore.getState().items[0].id).toBe(1)
  })
})
