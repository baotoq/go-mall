import { test, expect } from "@playwright/test"

test.describe("Auth flows", () => {
  test("unauthenticated /cart redirects to /signin with callbackUrl", async ({ page }) => {
    await page.goto("/cart")
    await expect(page).toHaveURL(/\/signin/)
    expect(page.url()).toContain("callbackUrl=%2Fcart")
  })

  test("sign-in page renders email and password fields", async ({ page }) => {
    await page.goto("/signin")
    await expect(page.getByLabel("Email")).toBeVisible()
    await expect(page.getByLabel("Password")).toBeVisible()
    await expect(page.getByRole("button", { name: /sign in/i })).toBeVisible()
  })

  test("sign-up page renders email and password fields", async ({ page }) => {
    await page.goto("/signup")
    await expect(page.getByLabel("Email")).toBeVisible()
    await expect(page.getByLabel("Password")).toBeVisible()
    await expect(page.getByRole("button", { name: /create account/i })).toBeVisible()
  })

  test("wrong credentials shows inline error", async ({ page }) => {
    await page.goto("/signin")
    await page.getByLabel("Email").fill("wrong@example.com")
    await page.getByLabel("Password").fill("wrongpassword123")
    await page.getByRole("button", { name: /sign in/i }).click()
    await expect(page.getByRole("alert")).toContainText("Invalid email or password")
  })

  test("header shows Sign in link when unauthenticated", async ({ page }) => {
    await page.goto("/")
    await expect(page.getByRole("link", { name: /sign in/i })).toBeVisible()
  })

  test("sign-up links to sign-in", async ({ page }) => {
    await page.goto("/signup")
    await page.getByRole("link", { name: /sign in/i }).click()
    await expect(page).toHaveURL(/\/signin/)
  })

  test("sign-in links to sign-up", async ({ page }) => {
    await page.goto("/signin")
    await page.getByRole("link", { name: /sign up/i }).click()
    await expect(page).toHaveURL(/\/signup/)
  })
})
