import type { Category, Product } from "./types"

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8001"

function mapProduct(raw: Record<string, unknown>): Product {
  return {
    id: String(raw.id ?? ""),
    name: String(raw.name ?? ""),
    slug: String(raw.slug ?? ""),
    description: String(raw.description ?? ""),
    priceCents: Number(raw.priceCents ?? 0),
    currency: String(raw.currency ?? "USD"),
    imageUrl: String(raw.imageUrl ?? ""),
    theme: String(raw.theme ?? ""),
    stock: Number(raw.stock ?? 0),
    categoryId: String(raw.categoryId ?? ""),
  }
}

function mapCategory(raw: Record<string, unknown>): Category {
  return {
    id: String(raw.id ?? ""),
    name: String(raw.name ?? ""),
    slug: String(raw.slug ?? ""),
    description: String(raw.description ?? ""),
  }
}

export async function listProducts(params?: {
  q?: string
  categoryId?: string
  page?: number
  pageSize?: number
}): Promise<{ products: Product[]; total: number }> {
  try {
    const url = new URL(`${API_URL}/v1/products`)
    if (params?.q) url.searchParams.set("q", params.q)
    if (params?.categoryId) url.searchParams.set("category_id", params.categoryId)
    if (params?.page) url.searchParams.set("page", String(params.page))
    if (params?.pageSize) url.searchParams.set("page_size", String(params.pageSize))

    const res = await fetch(url.toString(), { cache: "no-store" })
    if (!res.ok) return { products: [], total: 0 }
    const data = await res.json()
    return {
      products: (data.products ?? []).map(mapProduct),
      total: Number(data.total ?? 0),
    }
  } catch {
    return { products: [], total: 0 }
  }
}

export async function getProduct(id: string): Promise<Product | null> {
  try {
    const res = await fetch(`${API_URL}/v1/products/${id}`, { cache: "no-store" })
    if (!res.ok) return null
    return mapProduct(await res.json())
  } catch {
    return null
  }
}

export async function listCategories(): Promise<Category[]> {
  try {
    const res = await fetch(`${API_URL}/v1/categories?page_size=100`, { cache: "no-store" })
    if (!res.ok) return []
    const data = await res.json()
    return (data.categories ?? []).map(mapCategory)
  } catch {
    return []
  }
}
