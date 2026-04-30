import React from "react"
import { render, screen, fireEvent } from "@testing-library/react"
import { beforeEach, describe, expect, it } from "vitest"
import { AddToCartButton } from "@/components/add-to-cart-button"
import { useCartStore } from "@/store/cart"

const product = {
  id: 3,
  name: "Laptop Stand",
  price: 39.99,
  category: "electronics",
  description: "Aluminum stand",
  emoji: "💻",
  rating: 4.7,
  reviews: 203,
}

beforeEach(() => {
  useCartStore.setState({ items: [] })
})

describe("AddToCartButton", () => {
  it("renders Add to Cart label", () => {
    render(<AddToCartButton product={product} />)
    expect(screen.getByText("Add to Cart")).toBeDefined()
  })

  it("shows Added to Cart feedback after click", async () => {
    render(<AddToCartButton product={product} />)
    fireEvent.click(screen.getByRole("button"))
    expect(screen.getByText("Added to Cart")).toBeDefined()
  })

  it("adds item to cart store on click", () => {
    render(<AddToCartButton product={product} />)
    fireEvent.click(screen.getByRole("button"))
    expect(useCartStore.getState().items).toHaveLength(1)
    expect(useCartStore.getState().items[0].name).toBe("Laptop Stand")
  })
})
