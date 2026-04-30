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
  },
  providers: [],
}
