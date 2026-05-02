import type { Category, CartData, CartItemData, Product } from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8001";

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
  };
}

function mapCategory(raw: Record<string, unknown>): Category {
  return {
    id: String(raw.id ?? ""),
    name: String(raw.name ?? ""),
    slug: String(raw.slug ?? ""),
    description: String(raw.description ?? ""),
  };
}

export async function listProducts(params?: {
  q?: string;
  categoryId?: string;
  page?: number;
  pageSize?: number;
}): Promise<{ products: Product[]; total: number }> {
  try {
    const url = new URL(`${API_URL}/v1/products`);
    if (params?.q) url.searchParams.set("q", params.q);
    if (params?.categoryId)
      url.searchParams.set("category_id", params.categoryId);
    if (params?.page) url.searchParams.set("page", String(params.page));
    if (params?.pageSize)
      url.searchParams.set("page_size", String(params.pageSize));

    const res = await fetch(url.toString(), { cache: "no-store" });
    if (!res.ok) return { products: [], total: 0 };
    const data = await res.json();
    return {
      products: (data.products ?? []).map(mapProduct),
      total: Number(data.total ?? 0),
    };
  } catch {
    return { products: [], total: 0 };
  }
}

export async function getProduct(id: string): Promise<Product | null> {
  try {
    const res = await fetch(`${API_URL}/v1/products/${id}`, {
      cache: "no-store",
    });
    if (!res.ok) return null;
    return mapProduct(await res.json());
  } catch {
    return null;
  }
}

export async function listCategories(): Promise<Category[]> {
  try {
    const res = await fetch(`${API_URL}/v1/categories?page_size=100`, {
      cache: "no-store",
    });
    if (!res.ok) return [];
    const data = await res.json();
    return (data.categories ?? []).map(mapCategory);
  } catch {
    return [];
  }
}

const CART_API_URL =
  process.env.NEXT_PUBLIC_CART_API_URL ?? "http://localhost:8002";

function mapCartItem(raw: Record<string, unknown>): CartItemData {
  return {
    id: String(raw.id ?? ""),
    productId: String(raw.productId ?? ""),
    name: String(raw.name ?? ""),
    priceCents: Number(raw.priceCents ?? 0),
    currency: String(raw.currency ?? "USD"),
    imageUrl: String(raw.imageUrl ?? ""),
    quantity: Number(raw.quantity ?? 0),
    subtotalCents: Number(raw.subtotalCents ?? 0),
  };
}

function mapCart(raw: Record<string, unknown>): CartData {
  return {
    id: String(raw.id ?? ""),
    sessionId: String(raw.sessionId ?? ""),
    items: ((raw.items ?? []) as Record<string, unknown>[]).map(mapCartItem),
    totalCents: Number(raw.totalCents ?? 0),
  };
}

export async function getCart(sessionId: string): Promise<CartData | null> {
  try {
    const res = await fetch(`${CART_API_URL}/v1/carts/${sessionId}`, {
      cache: "no-store",
    });
    if (!res.ok) return null;
    return mapCart(await res.json());
  } catch {
    return null;
  }
}

export async function addCartItem(
  sessionId: string,
  item: {
    productId: string;
    name: string;
    priceCents: number;
    currency: string;
    imageUrl: string;
    quantity: number;
  },
): Promise<CartData | null> {
  try {
    const res = await fetch(`${CART_API_URL}/v1/carts/${sessionId}/items`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        product_id: item.productId,
        name: item.name,
        price_cents: item.priceCents,
        currency: item.currency,
        image_url: item.imageUrl,
        quantity: item.quantity,
      }),
    });
    if (!res.ok) return null;
    return mapCart(await res.json());
  } catch {
    return null;
  }
}

export async function updateCartItem(
  sessionId: string,
  productId: string,
  quantity: number,
): Promise<CartData | null> {
  try {
    const res = await fetch(
      `${CART_API_URL}/v1/carts/${sessionId}/items/${productId}`,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ quantity }),
      },
    );
    if (!res.ok) return null;
    return mapCart(await res.json());
  } catch {
    return null;
  }
}

export async function removeCartItem(
  sessionId: string,
  productId: string,
): Promise<CartData | null> {
  try {
    const res = await fetch(
      `${CART_API_URL}/v1/carts/${sessionId}/items/${productId}`,
      { method: "DELETE" },
    );
    if (!res.ok) return null;
    return mapCart(await res.json());
  } catch {
    return null;
  }
}

export async function clearCart(sessionId: string): Promise<void> {
  try {
    await fetch(`${CART_API_URL}/v1/carts/${sessionId}`, { method: "DELETE" });
  } catch {}
}

export async function seedData(): Promise<{
  categoriesSeeded: number;
  productsSeeded: number;
} | null> {
  try {
    const res = await fetch(`${API_URL}/v1/admin/seed`, { method: "POST" });
    if (!res.ok) return null;
    const data = await res.json();
    return {
      categoriesSeeded: Number(data.categories_seeded ?? 0),
      productsSeeded: Number(data.products_seeded ?? 0),
    };
  } catch {
    return null;
  }
}

export async function cleanData(): Promise<{
  categoriesDeleted: number;
  productsDeleted: number;
} | null> {
  try {
    const res = await fetch(`${API_URL}/v1/admin/clean`, { method: "DELETE" });
    if (!res.ok) return null;
    const data = await res.json();
    return {
      categoriesDeleted: Number(data.categories_deleted ?? 0),
      productsDeleted: Number(data.products_deleted ?? 0),
    };
  } catch {
    return null;
  }
}
