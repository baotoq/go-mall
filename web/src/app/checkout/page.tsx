import { auth } from "@/auth"
import { redirect } from "next/navigation"
import { CheckoutClient } from "./checkout-client"

export default async function CheckoutPage() {
  const session = await auth()
  if (!session) {
    redirect("/signin?callbackUrl=/checkout")
  }
  return <CheckoutClient />
}
