import { expect, test } from "@playwright/test";

// Requires: Keycloak (`gomall` realm) on :8080 + `npm run dev` on :3000.
// Each test creates its own user via the live `/api/auth/register` endpoint
// so runs are independent and idempotent against the realm.

function uniqueEmail(label: string): string {
  return `e2e-${label}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}@example.com`;
}

test.describe("Auth", () => {
  test("sign up creates account, signs the user in, redirects to home", async ({
    page,
  }) => {
    const email = uniqueEmail("signup");
    const password = "TestPass123!";

    await page.goto("/signup");
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Password").fill(password);
    await page.getByRole("button", { name: /create account/i }).click();

    await expect(page).toHaveURL("/");
    await expect(page.getByRole("button", { name: /sign out/i })).toBeVisible();
  });

  test("sign in with valid credentials redirects to callbackUrl", async ({
    page,
    request,
  }) => {
    const email = uniqueEmail("signin-ok");
    const password = "TestPass123!";

    const register = await request.post("/api/auth/register", {
      headers: { Origin: "http://localhost:3000" },
      data: { email, password },
    });
    expect(register.ok()).toBe(true);

    await page.goto("/signin?callbackUrl=%2Fcart");
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Password").fill(password);
    await page
      .locator("form")
      .getByRole("button", { name: /^sign in$/i })
      .click();

    await expect(page).toHaveURL("/cart");
    await expect(page.getByRole("button", { name: /sign out/i })).toBeVisible();
  });

  test("sign in with invalid credentials shows error and stays on /signin", async ({
    page,
  }) => {
    await page.goto("/signin");
    await page.getByLabel("Email").fill("nobody@example.com");
    await page.getByLabel("Password").fill("WrongPassword123!");
    await page
      .locator("form")
      .getByRole("button", { name: /^sign in$/i })
      .click();

    await expect(
      page.locator("form").getByText("Invalid email or password"),
    ).toBeVisible();
    await expect(page).toHaveURL(/\/signin/);
  });
});
