import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { ProductCard } from "@/components/product-card";
import { useCartStore } from "@/store/cart";

const product = {
  id: "abc-123",
  name: "Wireless Headphones",
  slug: "wireless-headphones",
  description: "Great headphones",
  priceCents: 7999,
  currency: "USD",
  imageUrl: "",
  theme: "",
  stock: 50,
  categoryId: "cat-1",
};

beforeEach(() => {
  useCartStore.setState({ items: [] });
});

describe("ProductCard", () => {
  it("renders product name and price", () => {
    render(<ProductCard product={product} />);
    expect(screen.getByText("Wireless Headphones")).toBeDefined();
    expect(screen.getByText("$79.99")).toBeDefined();
  });

  it("renders description excerpt", () => {
    render(<ProductCard product={product} />);
    expect(screen.getByText("Great headphones")).toBeDefined();
  });

  it("adds item to cart on click", () => {
    render(<ProductCard product={product} />);
    fireEvent.click(screen.getByRole("button", { name: /add/i }));
    expect(useCartStore.getState().items).toHaveLength(1);
    expect(useCartStore.getState().items[0].id).toBe("abc-123");
  });
});
