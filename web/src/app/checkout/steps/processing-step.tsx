"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { Loader2, CheckCircle2, XCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { Order } from "@/lib/types"

interface ProcessingStepProps {
  order: Order | null
  errorMessage: string | null
  onRetry: () => void
}

export function ProcessingStep({ order, errorMessage, onRetry }: ProcessingStepProps) {
  const router = useRouter()

  useEffect(() => {
    if (order?.status === "paid") {
      const timer = setTimeout(() => {
        router.push(`/orders/${order.id}`)
      }, 1500)
      return () => clearTimeout(timer)
    }
  }, [order, router])

  if (errorMessage) {
    return (
      <div className="flex flex-col items-center gap-4 py-12 text-center" data-testid="payment-error">
        <XCircle className="size-16 text-destructive" />
        <h2 className="text-xl font-bold">Payment Failed</h2>
        <p className="text-muted-foreground max-w-sm">{errorMessage}</p>
        <Button onClick={onRetry} size="lg">
          Try Again
        </Button>
      </div>
    )
  }

  if (order?.status === "paid") {
    return (
      <div className="flex flex-col items-center gap-4 py-12 text-center" data-testid="payment-success">
        <CheckCircle2 className="size-16 text-green-600" />
        <h2 className="text-xl font-bold">Order Confirmed!</h2>
        <p className="text-muted-foreground">Order #{order.id.slice(0, 8)} placed successfully.</p>
        <p className="text-sm text-muted-foreground">Redirecting to your order...</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col items-center gap-4 py-12 text-center" data-testid="payment-processing">
      <Loader2 className="size-16 animate-spin text-primary" />
      <h2 className="text-xl font-bold">Processing Payment</h2>
      <p className="text-muted-foreground">Please wait while we process your order...</p>
    </div>
  )
}
