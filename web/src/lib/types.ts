export interface Product {
  id: string
  name: string
  slug: string
  description: string
  priceCents: number
  currency: string
  imageUrl: string
  theme: string
  stock: number
  categoryId: string
}

export interface Category {
  id: string
  name: string
  slug: string
  description: string
}

export interface CartItemData {
  id: string
  productId: string
  name: string
  priceCents: number
  currency: string
  imageUrl: string
  quantity: number
  subtotalCents: number
}

export interface CartData {
  id: string
  sessionId: string
  items: CartItemData[]
  totalCents: number
}

export interface Address {
  fullName: string
  line1: string
  line2: string
  city: string
  state: string
  postalCode: string
  country: string
}

export type PaymentMethodType = "card" | "cod" | "bank_transfer"

export interface PaymentMethod {
  type: PaymentMethodType
  cardLastFour?: string
  cardBrand?: string
}

export type OrderStatus = "pending" | "processing" | "paid" | "failed" | "cancelled"

export interface OrderItem {
  productId: string
  name: string
  priceCents: number
  currency: string
  imageUrl: string
  quantity: number
  subtotalCents: number
}

export interface Order {
  id: string
  sessionId: string
  items: OrderItem[]
  address: Address
  paymentMethod: PaymentMethod
  status: OrderStatus
  totalCents: number
  createdAt: string
  updatedAt: string
}
