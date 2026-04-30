import type { Order, OrderItem, Address, PaymentMethod, OrderStatus } from "./types"
import type { CartItem } from "@/store/cart"

const ORDERS_KEY = "go_mall_orders"

function getSessionId(): string {
  if (typeof window === "undefined") return "ssr"
  let id = localStorage.getItem("cart_session_id")
  if (!id) { id = crypto.randomUUID(); localStorage.setItem("cart_session_id", id) }
  return id
}

function loadOrders(): Order[] {
  if (typeof window === "undefined") return []
  try {
    const raw = localStorage.getItem(ORDERS_KEY)
    return raw ? (JSON.parse(raw) as Order[]) : []
  } catch {
    return []
  }
}

function saveOrders(orders: Order[]): void {
  if (typeof window === "undefined") return
  localStorage.setItem(ORDERS_KEY, JSON.stringify(orders))
}

export async function createOrder(params: {
  cartItems: CartItem[]
  address: Address
  paymentMethod: PaymentMethod
  totalCents: number
}): Promise<Order> {
  const sessionId = getSessionId()
  const now = new Date().toISOString()
  const order: Order = {
    id: crypto.randomUUID(),
    sessionId,
    items: params.cartItems.map((item): OrderItem => ({
      productId: item.id,
      name: item.name,
      priceCents: item.priceCents,
      currency: "USD",
      imageUrl: item.imageUrl,
      quantity: item.quantity,
      subtotalCents: item.priceCents * item.quantity,
    })),
    address: params.address,
    paymentMethod: params.paymentMethod,
    status: "pending",
    totalCents: params.totalCents,
    createdAt: now,
    updatedAt: now,
  }
  const orders = loadOrders()
  orders.unshift(order)
  saveOrders(orders)
  return order
}

export async function processPayment(orderId: string): Promise<Order> {
  await new Promise((resolve) => setTimeout(resolve, 1500))

  const orders = loadOrders()
  const idx = orders.findIndex((o) => o.id === orderId)
  if (idx === -1) throw new Error(`Order ${orderId} not found`)

  const order = orders[idx]

  // simulate deterministic failure for specific test card pattern (last four "0000")
  const shouldFail =
    order.paymentMethod.type === "card" && order.paymentMethod.cardLastFour === "0000"

  const updatedOrder: Order = {
    ...order,
    status: (shouldFail ? "failed" : "paid") as OrderStatus,
    updatedAt: new Date().toISOString(),
  }
  orders[idx] = updatedOrder
  saveOrders(orders)
  return updatedOrder
}

export async function listOrders(): Promise<Order[]> {
  const sessionId = getSessionId()
  return loadOrders().filter((o) => o.sessionId === sessionId)
}

export async function getOrder(id: string): Promise<Order | null> {
  const orders = loadOrders()
  return orders.find((o) => o.id === id) ?? null
}

export async function updateOrderStatus(id: string, status: OrderStatus): Promise<Order | null> {
  const orders = loadOrders()
  const idx = orders.findIndex((o) => o.id === id)
  if (idx === -1) return null
  const updated: Order = { ...orders[idx], status, updatedAt: new Date().toISOString() }
  orders[idx] = updated
  saveOrders(orders)
  return updated
}
