import NextAuth from "next-auth"
import Credentials from "next-auth/providers/credentials"
import { authConfig } from "./auth.config"
import { getKeycloakToken, refreshKeycloakToken } from "@/lib/keycloak"

declare module "next-auth" {
  interface Session {
    access_token: string
    error?: "RefreshTokenError"
  }
  interface User {
    access_token: string
    refresh_token: string
    expires_at: number
  }
}

declare module "@auth/core/jwt" {
  interface JWT {
    access_token: string
    refresh_token: string
    expires_at: number
    error?: "RefreshTokenError"
  }
}

export const { handlers, auth, signIn, signOut } = NextAuth({
  ...authConfig,
  providers: [
    Credentials({
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        const { email, password } = credentials as { email: string; password: string }
        const tokens = await getKeycloakToken(email, password)
        if (!tokens) return null
        return {
          id: tokens.sub,
          email,
          access_token: tokens.access_token,
          refresh_token: tokens.refresh_token,
          expires_at: Math.floor(Date.now() / 1000) + tokens.expires_in,
        }
      },
    }),
  ],
  callbacks: {
    ...authConfig.callbacks,
    session({ session, token }) {
      if (token.access_token) session.access_token = token.access_token
      if (token.error) session.error = token.error
      return session
    },
    jwt({ token, user }) {
      if (user) {
        return {
          ...token,
          access_token: user.access_token,
          refresh_token: user.refresh_token,
          expires_at: user.expires_at,
          error: undefined,
        }
      }
      if (!token.expires_at || Date.now() < token.expires_at * 1000) {
        return token
      }
      return refreshKeycloakToken(token.refresh_token)
        .then((tokens) => ({
          ...token,
          access_token: tokens.access_token,
          refresh_token: tokens.refresh_token,
          expires_at: Math.floor(Date.now() / 1000) + tokens.expires_in,
          error: undefined,
        }))
        .catch(() => ({ ...token, error: "RefreshTokenError" as const }))
    },
  },
})
