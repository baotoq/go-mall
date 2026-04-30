"use client"

import Link from "next/link"
import { useEffect, useState } from "react"
import { Package, ArrowLeft, ShoppingBag } from "lucide-react"
import { listOrders } from "@/lib/orders-api"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { Order } from "@/lib/types"

function statusLabel(status: Order["status"]): string {
  const labels: Record<Order["status"], string> = {
    pending: "Pending",
    processing: "Processing",
    paid: "Paid",
    failed: "Failed",
    cancelled: "Cancelled",
  }
  return labels[status]
}

function statusClass(status: Order["status"]): string {
  const classes: Record<Order["status"], string> = {
    pending: "bg-yellow-100 text-yellow-800",
    processing: "bg-blue-100 text-blue-800",
    paid: "bg-green-100 text-green-800",
    failed: "bg-red-100 text-red-800",
    cancelled: "bg-gray-100 text-gray-600",
  }
  return classes[status]
}

export function OrdersClient() {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listOrders().then((o) => {
      setOrders(o)
      setLoading(false)
    })
  }, [])

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center py-24">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" aria-label="Loading" />
      </div>
    )
  }

  if (orders.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-4 py-24 px-4">
        <ShoppingBag className="size-16 text-muted-foreground/30" />
        <h1 className="text-2xl font-bold">No orders yet</h1>
        <p className="text-muted-foreground">Place your first order to see it here.</p>
        <Link href="/products" className={cn(buttonVariants({ size: "lg" }), "mt-2")}>
          Browse Products
        </Link>
      </div>
    )
  }

  return (
    <div className="flex-1 py-8 px-4">
      <div className="mx-auto max-w-3xl">
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground mb-8 transition-colors"
        >
          <ArrowLeft className="size-4" />
          Back to Home
        </Link>

        <h1 className="text-3xl font-bold mb-8">Your Orders</h1>

        <div className="space-y-4">
          {orders.map((order) => (
            <Link
              key={order.id}
              href={`/orders/${order.id}`}
              className="block rounded-xl border bg-card p-5 hover:shadow-md transition-shadow"
              data-testid="order-row"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-3">
                  <Package className="size-5 text-muted-foreground shrink-0 mt-0.5" />
                  <div>
                    <p className="font-medium text-sm">Order #{order.id.slice(0, 8).toUpperCase()}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {new Date(order.createdAt).toLocaleDateString(undefined, {
                        year: "numeric",
                        month: "long",
                        day: "numeric",
                      })}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {order.items.length} item{order.items.length !== 1 ? "s" : ""}
                    </p>
                  </div>
                </div>
                <div className="flex flex-col items-end gap-2 shrink-0">
                  <span
                    className={cn("rounded-full px-2.5 py-0.5 text-xs font-medium", statusClass(order.status))}
                    data-testid="order-status"
                  >
                    {statusLabel(order.status)}
                  </span>
                  <span className="font-bold">${(order.totalCents / 100).toFixed(2)}</span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  )
}
