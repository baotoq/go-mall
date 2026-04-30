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
