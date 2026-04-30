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
