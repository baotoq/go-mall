import { expect, test } from "@playwright/test";

test.describe("Product detail page", () => {
  test("renders product info", async ({ page }) => {
    await page.goto("/products/1");
    await expect(
      page.getByRole("heading", { name: "Wireless Headphones" }),
    ).toBeVisible();
    await expect(page.getByText("$79.99")).toBeVisible();
    await expect(page.getByText("Electronics")).toBeVisible();
  });

  test("renders Add to Cart button", async ({ page }) => {
    await page.goto("/products/1");
    await expect(
      page.getByRole("button", { name: /add to cart/i }),
    ).toBeVisible();
  });

  test("shows Added to Cart feedback after click", async ({ page }) => {
    await page.goto("/products/1");
    await page.getByRole("button", { name: /add to cart/i }).click();
    await expect(page.getByText("Added to Cart")).toBeVisible();
  });

  test("cart badge increments after adding item", async ({ page }) => {
    await page.goto("/products/1");
    await page.getByRole("button", { name: /add to cart/i }).click();
    await expect(page.locator("header").getByText("1")).toBeVisible();
  });

  test("back link returns to products", async ({ page }) => {
    await page.goto("/products/1");
    await page.getByRole("link", { name: /back to products/i }).click();
    await expect(page).toHaveURL("/products");
  });

  test("returns 404 for unknown product", async ({ page }) => {
    const res = await page.goto("/products/9999");
    expect(res?.status()).toBe(404);
  });
});
