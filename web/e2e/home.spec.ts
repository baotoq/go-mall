import { test, expect } from "@playwright/test";

test.describe("Home page", () => {
  test("renders hero section", async ({ page }) => {
    await page.goto("/");
    await expect(
      page.getByRole("heading", { name: /shop the best deals/i }),
    ).toBeVisible();
    await expect(page.getByRole("link", { name: /shop now/i })).toBeVisible();
  });

  test("renders all 4 categories", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByText("Electronics", { exact: true })).toBeVisible();
    await expect(page.getByText("Clothing", { exact: true })).toBeVisible();
    await expect(page.getByText("Books", { exact: true })).toBeVisible();
    await expect(
      page.getByText("Home & Garden", { exact: true }),
    ).toBeVisible();
  });

  test("renders featured products", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByText("Featured Products")).toBeVisible();
    const cards = page.locator(".rounded-xl.border").filter({ hasText: "$" });
    await expect(cards).toHaveCount(4);
  });

  test("category link navigates to filtered products", async ({ page }) => {
    await page.goto("/");
    await page
      .getByRole("link", { name: /electronics/i })
      .first()
      .click();
    await expect(page).toHaveURL(/category=electronics/);
    await expect(page.getByRole("heading", { name: "Products" })).toBeVisible();
  });
});
