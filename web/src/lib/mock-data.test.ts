import { describe, expect, it } from "vitest"
import {
  getProductById,
  getProductsByCategory,
  getFeaturedProducts,
  products,
} from "@/lib/mock-data"

describe("getProductById", () => {
  it("returns product for valid id", () => {
    const product = getProductById(1)
    expect(product).toBeDefined()
    expect(product?.id).toBe(1)
    expect(product?.name).toBe("Wireless Headphones")
  })

  it("returns undefined for unknown id", () => {
    expect(getProductById(9999)).toBeUndefined()
  })
})

describe("getProductsByCategory", () => {
  it("returns all products when no category given", () => {
    expect(getProductsByCategory()).toHaveLength(products.length)
  })

  it("returns all products for 'all'", () => {
    expect(getProductsByCategory("all")).toHaveLength(products.length)
  })

  it("filters by category", () => {
    const electronics = getProductsByCategory("electronics")
    expect(electronics.length).toBeGreaterThan(0)
    expect(electronics.every((p) => p.category === "electronics")).toBe(true)
  })

  it("returns empty array for unknown category", () => {
    expect(getProductsByCategory("unicorns")).toHaveLength(0)
  })
})

describe("getFeaturedProducts", () => {
  it("returns only products with a badge", () => {
    const featured = getFeaturedProducts()
    expect(featured.length).toBeGreaterThan(0)
    expect(featured.every((p) => p.badge)).toBe(true)
  })
})
