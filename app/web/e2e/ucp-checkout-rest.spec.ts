import { test, expect } from "@playwright/test";

// Requires: make dev (cart service at :8002) + UCP_ENABLED=true npm run dev
// Assumes a cart exists with session ID "e2e-test-session" and at least one item.
// Pre-condition: add a product to the cart via the UI or API before running.

test.describe("UCP Checkout REST flow", () => {
  const CART_SESSION = "e2e-test-session";
  let sessionId: string;

  test("full checkout flow: create → update buyer → complete", async ({
    request,
  }) => {
    // 1. Create checkout
    const createRes = await request.post("/api/ucp/checkout", {
      data: { cart_session_id: CART_SESSION, currency: "USD" },
      headers: { "Content-Type": "application/json" },
    });
    expect(createRes.status()).toBe(201);

    const createBody = await createRes.json();
    expect(createBody.session_id).toBeTruthy();
    expect(createBody.status).toBe("incomplete");
    sessionId = createBody.session_id;

    // 2. PATCH buyer email → ready_for_complete
    const patchRes = await request.patch(`/api/ucp/checkout/${sessionId}`, {
      data: { buyer: { email: "e2e-test@example.com" } },
      headers: {
        "Content-Type": "application/json",
        "X-UCP-Session": sessionId,
      },
    });
    expect(patchRes.status()).toBe(200);
    const patchBody = await patchRes.json();
    expect(patchBody.status).toBe("ready_for_complete");

    // 3. Complete checkout → creates Go order
    const completeRes = await request.post(
      `/api/ucp/checkout/${sessionId}?action=complete`,
      {
        data: {},
        headers: {
          "Content-Type": "application/json",
          "X-UCP-Session": sessionId,
        },
      },
    );
    expect(completeRes.status()).toBe(200);
    const completeBody = await completeRes.json();
    expect(completeBody.status).toBe("completed");
  });

  test("returns 503 when UCP_ENABLED is not set", async ({ request }) => {
    // This test only applies if UCP_ENABLED env is false (integration check).
    // When running with UCP_ENABLED=true the route returns normally.
    // Skip if the server is running with UCP enabled.
    const res = await request.get("/.well-known/ucp");
    if (res.status() === 200) {
      test.skip();
    }
    expect(res.status()).toBe(503);
  });
});
