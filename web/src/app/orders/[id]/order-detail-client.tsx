"use client"

import Link from "next/link"
import { useEffect, useState } from "react"
import { ArrowLeft, CheckCircle, XCircle, Clock, Package } from "lucide-react"
import { getOrder } from "@/lib/orders-api"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { Order } from "@/lib/types"

function StatusIcon({ status }: { status: Order["status"] }) {
  if (status === "paid") return <CheckCircle className="size-12 text-green-500" />
  if (status === "failed") return <XCircle className="size-12 text-red-500" />
  if (status === "cancelled") return <XCircle className="size-12 text-gray-400" />
  return <Clock className="size-12 text-yellow-500" />
}

function statusLabel(status: Order["status"]): string {
  const labels: Record<Order["status"], string> = {
    pending: "Pending",
    processing: "Processing",
    paid: "Order Confirmed",
    failed: "Payment Failed",
    cancelled: "Cancelled",
  }
  return labels[status]
}

function paymentLabel(order: Order): string {
  const { paymentMethod } = order
  if (paymentMethod.type === "card") {
    return `${paymentMethod.cardBrand ?? "Card"} ending in ${paymentMethod.cardLastFour}`
  }
  if (paymentMethod.type === "cod") return "Cash on Delivery"
  return "Bank Transfer"
}

export function OrderDetailClient({ id }: { id: string }) {
  const [order, setOrder] = useState<Order | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getOrder(id).then((o) => {
      setOrder(o)
      setLoading(false)
    })
  }, [id])

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center py-24">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" aria-label="Loading" />
      </div>
    )
  }

  if (!order) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-4 py-24 px-4">
        <Package className="size-16 text-muted-foreground/30" />
        <h1 className="text-2xl font-bold">Order not found</h1>
        <Link href="/orders" className={cn(buttonVariants(), "mt-2")}>
          Back to Orders
        </Link>
      </div>
    )
  }

  return (
    <div className="flex-1 py-8 px-4">
      <div className="mx-auto max-w-2xl">
        <Link
          href="/orders"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground mb-8 transition-colors"
        >
          <ArrowLeft className="size-4" />
          Back to Orders
        </Link>

        {/* Status banner */}
        <div className="flex flex-col items-center gap-3 rounded-xl border bg-card p-8 mb-6 text-center">
          <StatusIcon status={order.status} />
          <h1 className="text-2xl font-bold" data-testid="order-status-heading">
            {statusLabel(order.status)}
          </h1>
          <p className="text-sm text-muted-foreground">
            Order #{order.id.slice(0, 8).toUpperCase()}
          </p>
          <p className="text-xs text-muted-foreground">
            Placed on{" "}
            {new Date(order.createdAt).toLocaleDateString(undefined, {
              year: "numeric",
              month: "long",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            })}
          </p>
        </div>

        {/* Items */}
        <div className="rounded-xl border bg-card p-5 mb-4 space-y-4">
          <h2 className="font-semibold">Items</h2>
          {order.items.map((item) => (
            <div key={item.productId} className="flex items-center gap-4">
              <div className="size-12 rounded-lg bg-muted flex items-center justify-center shrink-0 overflow-hidden">
                {item.imageUrl ? (
                  <img src={item.imageUrl} alt={item.name} className="w-full h-full object-cover" />
                ) : (
                  <span className="text-base font-bold text-muted-foreground/30 select-none">
                    {item.name.charAt(0).toUpperCase()}
                  </span>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p className="font-medium text-sm line-clamp-1">{item.name}</p>
                <p className="text-xs text-muted-foreground">
                  {item.quantity} × ${(item.priceCents / 100).toFixed(2)}
                </p>
              </div>
              <p className="font-bold text-sm shrink-0">
                ${(item.subtotalCents / 100).toFixed(2)}
              </p>
            </div>
          ))}
          <div className="border-t pt-3 flex justify-between font-bold">
            <span>Total</span>
            <span data-testid="order-total">${(order.totalCents / 100).toFixed(2)}</span>
          </div>
        </div>

        {/* Shipping address */}
        <div className="rounded-xl border bg-card p-5 mb-4">
          <h2 className="font-semibold mb-3">Shipping Address</h2>
          <address className="not-italic text-sm text-muted-foreground space-y-0.5">
            <p className="font-medium text-foreground">{order.address.fullName}</p>
            <p>{order.address.line1}</p>
            {order.address.line2 && <p>{order.address.line2}</p>}
            <p>
              {order.address.city}, {order.address.state} {order.address.postalCode}
            </p>
            <p>{order.address.country}</p>
          </address>
        </div>

        {/* Payment */}
        <div className="rounded-xl border bg-card p-5 mb-6">
          <h2 className="font-semibold mb-3">Payment</h2>
          <p className="text-sm text-muted-foreground">{paymentLabel(order)}</p>
        </div>

        <div className="flex gap-3">
          <Link href="/orders" className={cn(buttonVariants({ variant: "outline" }), "flex-1")}>
            All Orders
          </Link>
          <Link href="/products" className={cn(buttonVariants(), "flex-1")}>
            Continue Shopping
          </Link>
        </div>
      </div>
    </div>
  )
}
