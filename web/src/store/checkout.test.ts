import { beforeEach, describe, expect, it } from "vitest"
import { useCheckoutStore } from "@/store/checkout"
import type { Address, PaymentMethod, Order } from "@/lib/types"

const mockAddress: Address = {
  fullName: "Jane Doe",
  line1: "123 Main St",
  line2: "",
  city: "New York",
  state: "NY",
  postalCode: "10001",
  country: "US",
}

const mockPayment: PaymentMethod = {
  type: "card",
  cardLastFour: "4242",
  cardBrand: "Visa",
}

const mockOrder: Order = {
  id: "order-abc",
  sessionId: "session-xyz",
  items: [],
  address: mockAddress,
  paymentMethod: mockPayment,
  status: "pending",
  totalCents: 4500,
  createdAt: "2026-01-01T00:00:00.000Z",
  updatedAt: "2026-01-01T00:00:00.000Z",
}

beforeEach(() => {
  useCheckoutStore.setState({
    step: "address",
    address: null,
    paymentMethod: null,
    order: null,
    errorMessage: null,
  })
})

describe("initial state", () => {
  it("starts at address step with no data", () => {
    const state = useCheckoutStore.getState()
    expect(state.step).toBe("address")
    expect(state.address).toBeNull()
    expect(state.paymentMethod).toBeNull()
    expect(state.order).toBeNull()
    expect(state.errorMessage).toBeNull()
  })
})

describe("setAddress", () => {
  it("saves address and advances to payment step", () => {
    useCheckoutStore.getState().setAddress(mockAddress)
    const state = useCheckoutStore.getState()
    expect(state.address).toEqual(mockAddress)
    expect(state.step).toBe("payment")
  })
})

describe("setPaymentMethod", () => {
  it("saves payment method and advances to review step", () => {
    useCheckoutStore.getState().setAddress(mockAddress)
    useCheckoutStore.getState().setPaymentMethod(mockPayment)
    const state = useCheckoutStore.getState()
    expect(state.paymentMethod).toEqual(mockPayment)
    expect(state.step).toBe("review")
  })
})

describe("setStep", () => {
  it("navigates to given step", () => {
    useCheckoutStore.getState().setStep("processing")
    expect(useCheckoutStore.getState().step).toBe("processing")
  })
})

describe("setOrder", () => {
  it("stores the created order", () => {
    useCheckoutStore.getState().setOrder(mockOrder)
    expect(useCheckoutStore.getState().order?.id).toBe("order-abc")
  })
})

describe("setError", () => {
  it("sets error message and switches to error step", () => {
    useCheckoutStore.getState().setError("Payment declined")
    const state = useCheckoutStore.getState()
    expect(state.errorMessage).toBe("Payment declined")
    expect(state.step).toBe("error")
  })
})

describe("reset", () => {
  it("returns to initial state", () => {
    useCheckoutStore.getState().setAddress(mockAddress)
    useCheckoutStore.getState().setPaymentMethod(mockPayment)
    useCheckoutStore.getState().setOrder(mockOrder)
    useCheckoutStore.getState().reset()
    const state = useCheckoutStore.getState()
    expect(state.step).toBe("address")
    expect(state.address).toBeNull()
    expect(state.paymentMethod).toBeNull()
    expect(state.order).toBeNull()
    expect(state.errorMessage).toBeNull()
  })
})
