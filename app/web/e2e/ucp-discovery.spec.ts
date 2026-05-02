import { expect, test } from "@playwright/test";

test.describe("UCP Discovery Profile", () => {
  test("GET /.well-known/ucp returns valid profile", async ({ request }) => {
    const res = await request.get("/.well-known/ucp");
    expect(res.status()).toBe(200);

    const body = await res.json();
    expect(body.ucp).toBeDefined();
    expect(
      body.ucp.capabilities.map((c: { name: string }) => c.name),
    ).toContain("dev.ucp.shopping.checkout");
  });

  test("profile has Cache-Control header", async ({ request }) => {
    const res = await request.get("/.well-known/ucp");
    expect(res.headers()["cache-control"]).toContain("public");
    expect(res.headers()["cache-control"]).toContain("max-age=3600");
  });

  test("profile lists REST and MCP transports", async ({ request }) => {
    const res = await request.get("/.well-known/ucp");
    const body = await res.json();
    const service = body.ucp.services["dev.ucp.shopping"];
    expect(service.rest).toBeDefined();
    expect(service.mcp).toBeDefined();
  });
});
