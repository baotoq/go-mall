import { expect, type Page, test } from "@playwright/test";

// Mock a NextAuth session so the checkout page server-component sees an
// authenticated user without a live Keycloak instance.
async function mockSession(page: Page) {
  await page.route("**/api/auth/session", (route) => {
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        user: { name: "Test User", email: "test@example.com", image: null },
        expires: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
      }),
    });
  });
}

async function fillCheckoutForm(page: Page) {
  // Contact
  await page.getByLabel("Email").fill("test@example.com");
  await page.getByLabel("Name").fill("Test User");
  // Shipping
  await page.getByLabel("Address Line 1").fill("123 Main St");
  await page.getByLabel("City").fill("Springfield");
  await page.getByLabel("State").fill("IL");
  await page.getByLabel("Postal Code").fill("62701");
  await page.getByLabel("Country").fill("US");
  // Payment
  await page.getByLabel("Card Number").fill("4242 4242 4242 4242");
  await page.getByLabel("Expiry (MM/YY)").fill("12/30");
  await page.getByLabel("CVC").fill("123");
}

test.describe("Checkout flow", () => {
  // Requires: `make dev` (cart :8002, catalog :8001, Keycloak :8080) +
  // `UCP_ENABLED=true MOCK_ORDER_SERVICE=true npm run dev` (FE :3000).
  test("happy path: add to cart → checkout → success page shows order id", async ({
    page,
  }) => {
    await mockSession(page);

    // Add a product to cart
    await page.goto("/products/1");
    await page.getByRole("button", { name: /add to cart/i }).click();
    await expect(page.locator("header").getByText("1")).toBeVisible();

    // Navigate to checkout via the cart page Checkout link
    await page.goto("/cart");
    await page.getByRole("link", { name: /checkout/i }).click();
    await expect(page).toHaveURL("/checkout");

    // Fill all form fields
    await fillCheckoutForm(page);

    // Submit
    await page.getByRole("button", { name: /place order/i }).click();

    // Expect success page with order id
    await expect(page).toHaveURL(/\/checkout\/success\?id=/, {
      timeout: 15000,
    });
    await expect(page.getByText(/Order #ord_/i)).toBeVisible();
  });

  // Cart service (loadCart → :8002) must be up for this spec.
  test("empty-cart redirect: signed-in user with no cart items → /cart", async ({
    page,
  }) => {
    await mockSession(page);

    // Visit checkout directly with an empty cart (no products added).
    // The client-side effect pushes to /cart when !isLoading && items.length === 0.
    await page.goto("/checkout");

    await expect(page).toHaveURL("/cart");
  });

  // Pure FE test — no Go backend required. The Next.js server component
  // redirects unauthenticated requests before any cart/order call is made.
  test("unauthenticated redirect: no session → /signin?callbackUrl=%2Fcheckout", async ({
    page,
  }) => {
    // Return an empty session object — NextAuth treats this as unauthenticated.
    await page.route("**/api/auth/session", (route) => {
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({}),
      });
    });

    await page.goto("/checkout");

    await expect(page).toHaveURL(/\/signin\?callbackUrl=(%2F|\/)checkout/);
  });
});
