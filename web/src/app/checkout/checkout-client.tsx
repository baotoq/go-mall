"use client"

import { useEffect } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { ArrowLeft } from "lucide-react"
import { useCheckoutStore } from "@/store/checkout"
import { useCartStore } from "@/store/cart"
import { createOrder, processPayment } from "@/lib/orders-api"
import { AddressStep } from "./steps/address-step"
import { PaymentStep } from "./steps/payment-step"
import { ReviewStep } from "./steps/review-step"
import { ProcessingStep } from "./steps/processing-step"
import type { Address, PaymentMethod } from "@/lib/types"

const STEP_LABELS = ["Address", "Payment", "Review", "Processing"]
const STEP_INDICES: Record<string, number> = {
  address: 0,
  payment: 1,
  review: 2,
  processing: 3,
  done: 3,
  error: 3,
}

export function CheckoutClient() {
  const router = useRouter()
  const { step, address, paymentMethod, order, errorMessage, setAddress, setPaymentMethod, setOrder, setStep, setError, reset } =
    useCheckoutStore()
  const { items, totalPrice, loadCart, clearCart } = useCartStore()

  useEffect(() => {
    loadCart()
  }, [loadCart])

  useEffect(() => {
    if (step === "address" && items.length === 0) {
      // Cart is empty — nothing to checkout
    }
  }, [step, items])

  async function handleConfirmOrder() {
    if (!address || !paymentMethod) return
    setStep("processing")

    try {
      const newOrder = await createOrder({
        cartItems: items,
        address,
        paymentMethod,
        totalCents: Math.round(totalPrice() * 100),
      })
      setOrder(newOrder)

      const processed = await processPayment(newOrder.id)
      setOrder(processed)

      if (processed.status === "paid") {
        clearCart()
      } else {
        setError("Your payment was declined. Please try a different payment method.")
      }
    } catch {
      setError("An unexpected error occurred. Please try again.")
    }
  }

  function handleRetry() {
    setStep("review")
  }

  const currentStepIndex = STEP_INDICES[step] ?? 0

  if (items.length === 0 && step === "address") {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-4 py-24 px-4">
        <h1 className="text-2xl font-bold">Your cart is empty</h1>
        <p className="text-muted-foreground">Add some products before checking out.</p>
        <Link href="/products" className="text-primary underline">Browse Products</Link>
      </div>
    )
  }

  return (
    <div className="flex-1 py-8 px-4">
      <div className="mx-auto max-w-xl">
        <Link
          href="/cart"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground mb-8 transition-colors"
        >
          <ArrowLeft className="size-4" />
          Back to Cart
        </Link>

        <h1 className="text-3xl font-bold mb-6">Checkout</h1>

        {/* Step indicator */}
        {step !== "processing" && step !== "done" && step !== "error" && (
          <div className="flex items-center gap-2 mb-8">
            {STEP_LABELS.slice(0, 3).map((label, i) => (
              <div key={label} className="flex items-center gap-2">
                <div
                  className={`flex items-center justify-center rounded-full text-sm font-medium size-7 ${
                    i < currentStepIndex
                      ? "bg-primary text-primary-foreground"
                      : i === currentStepIndex
                        ? "bg-primary text-primary-foreground ring-2 ring-primary/30"
                        : "bg-muted text-muted-foreground"
                  }`}
                >
                  {i < currentStepIndex ? "✓" : i + 1}
                </div>
                <span className={`text-sm ${i === currentStepIndex ? "font-semibold" : "text-muted-foreground"}`}>
                  {label}
                </span>
                {i < 2 && <div className="h-px w-8 bg-border" />}
              </div>
            ))}
          </div>
        )}

        {step === "address" && (
          <AddressStep
            defaultValues={address ?? undefined}
            onSubmit={(addr: Address) => setAddress(addr)}
          />
        )}

        {step === "payment" && (
          <PaymentStep
            onSubmit={(method: PaymentMethod) => setPaymentMethod(method)}
            onBack={() => setStep("address")}
          />
        )}

        {step === "review" && address && paymentMethod && (
          <ReviewStep
            items={items}
            address={address}
            paymentMethod={paymentMethod}
            totalPrice={totalPrice()}
            onConfirm={handleConfirmOrder}
            onBack={() => setStep("payment")}
          />
        )}

        {(step === "processing" || step === "done" || step === "error") && (
          <ProcessingStep
            order={order}
            errorMessage={errorMessage}
            onRetry={handleRetry}
          />
        )}
      </div>
    </div>
  )
}
