"use client"

import { Button } from "@/components/ui/button"
import type { Address, PaymentMethod } from "@/lib/types"
import type { CartItem } from "@/store/cart"

interface ReviewStepProps {
  items: CartItem[]
  address: Address
  paymentMethod: PaymentMethod
  totalPrice: number
  onConfirm: () => void
  onBack: () => void
}

function formatPayment(method: PaymentMethod): string {
  if (method.type === "card") return `Card ending in ${method.cardLastFour}`
  if (method.type === "cod") return "Cash on Delivery"
  return "Bank Transfer"
}

export function ReviewStep({ items, address, paymentMethod, totalPrice, onConfirm, onBack }: ReviewStepProps) {
  return (
    <div className="space-y-6" data-testid="review-step">
      <section className="rounded-lg border p-4 space-y-2">
        <h3 className="font-semibold">Shipping Address</h3>
        <p className="text-sm">{address.fullName}</p>
        <p className="text-sm text-muted-foreground">
          {address.line1}{address.line2 ? `, ${address.line2}` : ""}
        </p>
        <p className="text-sm text-muted-foreground">
          {address.city}, {address.state} {address.postalCode}, {address.country}
        </p>
      </section>

      <section className="rounded-lg border p-4 space-y-2">
        <h3 className="font-semibold">Payment</h3>
        <p className="text-sm text-muted-foreground">{formatPayment(paymentMethod)}</p>
      </section>

      <section className="rounded-lg border p-4 space-y-3">
        <h3 className="font-semibold">Order Items</h3>
        {items.map((item) => (
          <div key={item.id} className="flex justify-between text-sm">
            <span>{item.name} × {item.quantity}</span>
            <span>${((item.priceCents / 100) * item.quantity).toFixed(2)}</span>
          </div>
        ))}
        <div className="border-t pt-3 flex justify-between font-bold">
          <span>Total</span>
          <span>${totalPrice.toFixed(2)}</span>
        </div>
      </section>

      <div className="flex gap-3">
        <Button type="button" variant="outline" size="lg" className="flex-1" onClick={onBack}>
          Back
        </Button>
        <Button type="button" size="lg" className="flex-1" onClick={onConfirm} data-testid="place-order">
          Place Order
        </Button>
      </div>
    </div>
  )
}
