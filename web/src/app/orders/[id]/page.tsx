import { auth } from "@/auth"
import { redirect } from "next/navigation"
import { OrderDetailClient } from "./order-detail-client"

export const metadata = { title: "Order Details — GoMall" }

export default async function OrderDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const session = await auth()
  if (!session) {
    redirect("/signin?callbackUrl=/orders")
  }
  const { id } = await params
  return <OrderDetailClient id={id} />
}
