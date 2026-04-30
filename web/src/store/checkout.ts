"use client"

import { create } from "zustand"
import type { Address, PaymentMethod, Order } from "@/lib/types"

export type CheckoutStep = "address" | "payment" | "review" | "processing" | "done" | "error"

interface CheckoutStore {
  step: CheckoutStep
  address: Address | null
  paymentMethod: PaymentMethod | null
  order: Order | null
  errorMessage: string | null

  setAddress: (address: Address) => void
  setPaymentMethod: (method: PaymentMethod) => void
  setOrder: (order: Order) => void
  setStep: (step: CheckoutStep) => void
  setError: (message: string) => void
  reset: () => void
}

const initialState = {
  step: "address" as CheckoutStep,
  address: null,
  paymentMethod: null,
  order: null,
  errorMessage: null,
}

export const useCheckoutStore = create<CheckoutStore>((set) => ({
  ...initialState,

  setAddress: (address) => set({ address, step: "payment" }),

  setPaymentMethod: (paymentMethod) => set({ paymentMethod, step: "review" }),

  setOrder: (order) => set({ order }),

  setStep: (step) => set({ step }),

  setError: (errorMessage) => set({ errorMessage, step: "error" }),

  reset: () => set(initialState),
}))
