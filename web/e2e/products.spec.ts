import { test, expect } from "@playwright/test"

test.describe("Products page", () => {
  test("shows all products by default", async ({ page }) => {
    await page.goto("/products")
    await expect(page.getByText("12 products")).toBeVisible()
  })

  test("filters by electronics category", async ({ page }) => {
    await page.goto("/products?category=electronics")
    await expect(page.getByText("4 products")).toBeVisible()
  })

  test("category sidebar links work", async ({ page }) => {
    await page.goto("/products")
    await page.getByRole("link", { name: /clothing/i }).click()
    await expect(page).toHaveURL(/category=clothing/)
    await expect(page.getByText("4 products")).toBeVisible()
  })

  test("all products link resets filter", async ({ page }) => {
    await page.goto("/products?category=books")
    await page.getByRole("link", { name: "All Products" }).click()
    await expect(page).toHaveURL("/products")
    await expect(page.getByText("12 products")).toBeVisible()
  })

  test("product card links to detail page", async ({ page }) => {
    await page.goto("/products")
    await page.getByRole("link", { name: "Wireless Headphones" }).first().click()
    await expect(page).toHaveURL("/products/1")
  })
})
