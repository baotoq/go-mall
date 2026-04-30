import { test, expect } from "@playwright/test"

test.describe("Cart page", () => {
  test("shows empty state when no items", async ({ page }) => {
    await page.goto("/cart")
    await expect(page.getByText("Your cart is empty")).toBeVisible()
    await expect(page.getByRole("link", { name: /browse products/i })).toBeVisible()
  })

  test("shows items added from product page", async ({ page }) => {
    await page.goto("/products/1")
    await page.getByRole("button", { name: /add to cart/i }).click()
    await page.getByRole("link", { name: /cart/i }).click()
    await expect(page).toHaveURL("/cart")
    await expect(page.getByText("Wireless Headphones")).toBeVisible()
    await expect(page.getByText("$79.99").first()).toBeVisible()
  })

  test("order summary shows correct total", async ({ page }) => {
    await page.goto("/products/1")
    await page.getByRole("button", { name: /add to cart/i }).click()
    await page.getByRole("link", { name: /cart/i }).click()
    await expect(page).toHaveURL("/cart")
    await expect(page.getByText("Order Summary")).toBeVisible()
    await expect(page.getByText("Free")).toBeVisible()
  })

  test("remove item clears cart", async ({ page }) => {
    await page.goto("/products/1")
    await page.getByRole("button", { name: /add to cart/i }).click()
    await page.getByRole("link", { name: /cart/i }).click()
    await expect(page).toHaveURL("/cart")
    await page.getByTestId("remove-item").click()
    await expect(page.getByText("Your cart is empty")).toBeVisible()
  })
})
