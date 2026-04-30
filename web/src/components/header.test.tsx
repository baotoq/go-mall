import { render, screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { Header } from "@/components/header"
import { useCartStore } from "@/store/cart"

vi.mock("next-auth/react", () => ({
  useSession: () => ({ data: null, status: "unauthenticated" }),
  signOut: vi.fn(),
}))

beforeEach(() => {
  useCartStore.setState({ items: [] })
})

describe("Header", () => {
  it("renders site name", () => {
    render(<Header />)
    expect(screen.getByText("GoMall")).toBeDefined()
  })

  it("renders nav links", () => {
    render(<Header />)
    expect(screen.getByText("Home")).toBeDefined()
    expect(screen.getByText("Products")).toBeDefined()
  })

  it("does not show badge when cart is empty", () => {
    render(<Header />)
    const badge = screen.queryByText(/^\d+$/)
    expect(badge).toBeNull()
  })

  it("shows item count badge when cart has items", () => {
    useCartStore.setState({
      items: [{ id: "test-1", name: "Test", priceCents: 1000, imageUrl: "", quantity: 2 }],
    })
    render(<Header />)
    expect(screen.getByText("2")).toBeDefined()
  })
})
