"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import type { PaymentMethod, PaymentMethodType } from "@/lib/types"

interface PaymentStepProps {
  onSubmit: (method: PaymentMethod) => void
  onBack: () => void
}

export function PaymentStep({ onSubmit, onBack }: PaymentStepProps) {
  const [methodType, setMethodType] = useState<PaymentMethodType>("card")
  const [cardNumber, setCardNumber] = useState("")
  const [cardError, setCardError] = useState("")

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    if (methodType === "card") {
      const digits = cardNumber.replace(/\s/g, "")
      if (digits.length < 4) {
        setCardError("Please enter at least 4 card digits")
        return
      }
      const lastFour = digits.slice(-4)
      onSubmit({ type: "card", cardLastFour: lastFour, cardBrand: "Visa" })
    } else if (methodType === "cod") {
      onSubmit({ type: "cod" })
    } else {
      onSubmit({ type: "bank_transfer" })
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6" data-testid="payment-form">
      <div className="space-y-3">
        <Label>Payment Method</Label>
        <RadioGroup value={methodType} onValueChange={(v) => setMethodType(v as PaymentMethodType)}>
          <div className="flex items-center gap-3 rounded-lg border p-4 cursor-pointer">
            <RadioGroupItem value="card" id="card" />
            <Label htmlFor="card" className="cursor-pointer flex-1">Credit / Debit Card</Label>
          </div>
          <div className="flex items-center gap-3 rounded-lg border p-4 cursor-pointer">
            <RadioGroupItem value="cod" id="cod" />
            <Label htmlFor="cod" className="cursor-pointer flex-1">Cash on Delivery</Label>
          </div>
          <div className="flex items-center gap-3 rounded-lg border p-4 cursor-pointer">
            <RadioGroupItem value="bank_transfer" id="bank_transfer" />
            <Label htmlFor="bank_transfer" className="cursor-pointer flex-1">Bank Transfer</Label>
          </div>
        </RadioGroup>
      </div>

      {methodType === "card" && (
        <div className="space-y-1">
          <Label htmlFor="cardNumber">Card Number</Label>
          <Input
            id="cardNumber"
            placeholder="1234 5678 9012 3456"
            value={cardNumber}
            onChange={(e) => { setCardNumber(e.target.value); setCardError("") }}
            aria-invalid={!!cardError}
            maxLength={19}
          />
          {cardError && <p className="text-sm text-destructive">{cardError}</p>}
          <p className="text-xs text-muted-foreground">Use card number ending in 0000 to test payment failure.</p>
        </div>
      )}

      <div className="flex gap-3">
        <Button type="button" variant="outline" size="lg" className="flex-1" onClick={onBack}>
          Back
        </Button>
        <Button type="submit" size="lg" className="flex-1">
          Review Order
        </Button>
      </div>
    </form>
  )
}
