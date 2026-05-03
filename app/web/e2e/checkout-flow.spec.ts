import { expect, test } from "@playwright/test";

// End-to-end checkout flow against live services:
// Keycloak :8080, Frontend :3000, Catalog :8001, Cart :8002.
// `MOCK_ORDER_SERVICE=true` (per .env.local) skips the Go order service
// and returns a deterministic mock order id.

function uniqueEmail(label: string): string {
  return `e2e-${label}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}@example.com`;
}

test.describe("Checkout flow (signed-in)", () => {
  test("happy path: register → sign in → add to cart → checkout → success", async ({
    page,
    request,
  }) => {
    const email = uniqueEmail("checkout");
    const password = "TestPass123!";

    // 1. Register a fresh user via the live endpoint
    const register = await request.post("/api/auth/register", {
      headers: { Origin: "http://localhost:3000" },
      data: { email, password },
    });
    expect(register.ok()).toBe(true);

    // 2. Sign in via the UI so NextAuth sets the JWT cookie
    await page.goto("/signin");
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Password").fill(password);
    await page
      .locator("form")
      .getByRole("button", { name: /^sign in$/i })
      .click();
    await expect(page).toHaveURL("/");

    // 3. Discover a real product id from the catalog
    const catalog = await request.get(
      "http://localhost:8001/v1/products?page_size=1",
    );
    expect(catalog.ok()).toBe(true);
    const { products } = (await catalog.json()) as {
      products: Array<{ id: string; name: string }>;
    };
    expect(products.length).toBeGreaterThan(0);
    const product = products[0];

    // 4. Add to cart from the product detail page
    await page.goto(`/products/${product.id}`);
    await page.getByRole("button", { name: /add to cart/i }).click();
    await expect(page.locator("header").getByText(/cart\s*1/i)).toBeVisible();

    // 5. Cart → Checkout
    await page.goto("/cart");
    await page.getByRole("link", { name: /^checkout$/i }).click();
    await expect(page).toHaveURL("/checkout");

    // 6. Fill checkout form. Email is pre-filled from session.
    await page.getByLabel("Name").fill("Test User");
    await page.getByLabel("Address Line 1").fill("123 Main St");
    await page.getByLabel("City").fill("Springfield");
    await page.getByLabel("State").fill("IL");
    await page.getByLabel("Postal Code").fill("62701");
    await page.getByLabel("Country").fill("US");
    await page.getByLabel("Card Number").fill("4242 4242 4242 4242");
    await page.getByLabel("Expiry (MM/YY)").fill("12/30");
    await page.getByLabel("CVC").fill("123");

    // 7. Place order → success page with id
    await page.getByRole("button", { name: /place order/i }).click();
    await expect(page).toHaveURL(/\/checkout\/success\?id=/);
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  });
});
