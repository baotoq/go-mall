const KEYCLOAK_INTERNAL_URL = process.env.KEYCLOAK_INTERNAL_URL!
const KEYCLOAK_CLIENT_ID = process.env.KEYCLOAK_CLIENT_ID!
const KEYCLOAK_CLIENT_SECRET = process.env.KEYCLOAK_CLIENT_SECRET!
const REALM = "gomall"

const TOKEN_URL = `${KEYCLOAK_INTERNAL_URL}/realms/${REALM}/protocol/openid-connect/token`
const ADMIN_USERS_URL = `${KEYCLOAK_INTERNAL_URL}/admin/realms/${REALM}/users`

function decodeJwtSub(token: string): string {
  const payloadBase64 = token.split(".")[1]
  const decoded = JSON.parse(atob(payloadBase64.replace(/-/g, "+").replace(/_/g, "/")))
  return decoded.sub as string
}

export async function getKeycloakToken(email: string, password: string) {
  const res = await fetch(TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "password",
      client_id: KEYCLOAK_CLIENT_ID,
      client_secret: KEYCLOAK_CLIENT_SECRET,
      username: email,
      password,
      scope: "openid email profile",
    }),
  })
  if (!res.ok) return null
  const data = await res.json()
  return {
    access_token: data.access_token as string,
    refresh_token: data.refresh_token as string,
    expires_in: data.expires_in as number,
    sub: decodeJwtSub(data.access_token as string),
  }
}

export async function refreshKeycloakToken(refreshToken: string) {
  const res = await fetch(TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "refresh_token",
      client_id: KEYCLOAK_CLIENT_ID,
      client_secret: KEYCLOAK_CLIENT_SECRET,
      refresh_token: refreshToken,
    }),
  })
  if (!res.ok) throw new Error("RefreshTokenError")
  const data = await res.json()
  return {
    access_token: data.access_token as string,
    refresh_token: data.refresh_token as string,
    expires_in: data.expires_in as number,
  }
}

async function getAdminToken(): Promise<string> {
  const res = await fetch(TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "client_credentials",
      client_id: KEYCLOAK_CLIENT_ID,
      client_secret: KEYCLOAK_CLIENT_SECRET,
    }),
  })
  if (!res.ok) throw new Error("AdminTokenError")
  const data = await res.json()
  return data.access_token as string
}

export async function createKeycloakUser(email: string, password: string): Promise<void> {
  const adminToken = await getAdminToken()
  const createRes = await fetch(ADMIN_USERS_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${adminToken}`,
    },
    body: JSON.stringify({ email, username: email, enabled: true, emailVerified: false }),
  })
  if (createRes.status === 409) throw new Error("UserAlreadyExists")
  if (!createRes.ok) throw new Error("CreateUserError")
  const location = createRes.headers.get("Location")
  if (!location) throw new Error("NoLocationHeader")
  const userId = location.split("/").pop()!
  try {
    const pwRes = await fetch(`${ADMIN_USERS_URL}/${userId}/reset-password`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${adminToken}` },
      body: JSON.stringify({ type: "password", temporary: false, value: password }),
    })
    if (!pwRes.ok) throw new Error("SetPasswordError")
  } catch (e) {
    await fetch(`${ADMIN_USERS_URL}/${userId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${adminToken}` },
    })
    throw e
  }
}
