import { beforeEach, describe, expect, it, vi } from "vitest";

const mockFetch = vi.fn();
vi.stubGlobal("fetch", mockFetch);

const { getKeycloakToken, refreshKeycloakToken, createKeycloakUser } =
  await import("@/lib/keycloak");

const fakeJwt = [
  btoa(JSON.stringify({ alg: "RS256", typ: "JWT" })),
  btoa(JSON.stringify({ sub: "user-sub-789", email: "user@test.com" })),
  "fakesig",
].join(".");

describe("getKeycloakToken", () => {
  beforeEach(() => mockFetch.mockReset());

  it("returns token data including decoded sub on success", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        access_token: fakeJwt,
        refresh_token: "refresh-456",
        expires_in: 300,
      }),
    });
    const result = await getKeycloakToken("user@test.com", "pass");
    expect(result).toEqual({
      access_token: fakeJwt,
      refresh_token: "refresh-456",
      expires_in: 300,
      sub: "user-sub-789",
    });
  });

  it("returns null on 401", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 401 });
    const result = await getKeycloakToken("user@test.com", "wrong");
    expect(result).toBeNull();
  });
});

describe("refreshKeycloakToken", () => {
  beforeEach(() => mockFetch.mockReset());

  it("returns new tokens on success", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        access_token: "new-access",
        refresh_token: "new-refresh",
        expires_in: 300,
      }),
    });
    const result = await refreshKeycloakToken("old-refresh-token");
    expect(result.access_token).toBe("new-access");
    expect(result.refresh_token).toBe("new-refresh");
  });

  it("throws RefreshTokenError on failure", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 401 });
    await expect(refreshKeycloakToken("bad-token")).rejects.toThrow(
      "RefreshTokenError",
    );
  });
});

describe("createKeycloakUser", () => {
  beforeEach(() => mockFetch.mockReset());

  it("creates user and sets password successfully", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ access_token: "admin-token" }),
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      headers: new Headers({
        Location: "http://kc/admin/realms/gomall/users/user-id-123",
      }),
    });
    mockFetch.mockResolvedValueOnce({ ok: true, status: 204 });
    await expect(
      createKeycloakUser("new@test.com", "password123"),
    ).resolves.toBeUndefined();
  });

  it("throws UserAlreadyExists on 409", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ access_token: "admin-token" }),
    });
    mockFetch.mockResolvedValueOnce({ ok: false, status: 409 });
    await expect(
      createKeycloakUser("exists@test.com", "password123"),
    ).rejects.toThrow("UserAlreadyExists");
  });

  it("deletes orphan user if password set fails", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ access_token: "admin-token" }),
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      headers: new Headers({
        Location: "http://kc/admin/realms/gomall/users/orphan-id",
      }),
    });
    mockFetch.mockResolvedValueOnce({ ok: false, status: 500 });
    mockFetch.mockResolvedValueOnce({ ok: true, status: 204 });
    await expect(
      createKeycloakUser("new@test.com", "password123"),
    ).rejects.toThrow();
    const deleteCall = mockFetch.mock.calls.find(
      ([url, opts]) =>
        String(url).includes("orphan-id") && opts?.method === "DELETE",
    );
    expect(deleteCall).toBeDefined();
  });
});
