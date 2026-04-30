import { auth } from "@/auth"
import { redirect } from "next/navigation"
import { OrdersClient } from "./orders-client"

export const metadata = { title: "Your Orders — GoMall" }

export default async function OrdersPage() {
  const session = await auth()
  if (!session) {
    redirect("/signin?callbackUrl=/orders")
  }
  return <OrdersClient />
}
