import type { NextAuthConfig } from "next-auth"

export const authConfig: NextAuthConfig = {
  pages: {
    signIn: "/signin",
  },
  callbacks: {
    authorized({ auth, request: { nextUrl } }) {
      const isLoggedIn = !!auth?.user
      const isCart = nextUrl.pathname.startsWith("/cart")
      if (isCart && !isLoggedIn) return false
      return true
    },
    session({ session, token }) {
      if (token.access_token) session.access_token = token.access_token as string
      if (token.error) session.error = token.error as "RefreshTokenError"
      return session
    },
  },
  providers: [],
}
