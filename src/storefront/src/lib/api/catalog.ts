export interface ProductInfo {
  id: string;
  name: string;
  slug: string;
  description: string;
  imageUrl: string;
  price: number;
  totalStock: number;
  remainingStock: number;
  categoryId: string;
  createdAt: number;
  updatedAt: number;
}

export interface CategoryInfo {
  id: string;
  name: string;
  slug: string;
  description: string;
  createdAt: number;
  updatedAt: number;
}

const CATALOG_URL =
  process.env.CATALOG_SERVICE_URL ||
  process.env.NEXT_PUBLIC_CATALOG_URL ||
  "http://localhost:9001";

export async function listProducts(params?: {
  keyword?: string;
  categoryId?: string;
  page?: number;
  pageSize?: number;
}): Promise<{ products: ProductInfo[]; total: number }> {
  const url = new URL(`${CATALOG_URL}/api/v1/products`);
  if (params?.keyword) url.searchParams.set("keyword", params.keyword);
  if (params?.categoryId) url.searchParams.set("categoryId", params.categoryId);
  if (params?.page) url.searchParams.set("page", params.page.toString());
  if (params?.pageSize)
    url.searchParams.set("pageSize", params.pageSize.toString());

  const res = await fetch(url, {
    next: { tags: ["products"], revalidate: 3600 },
  });
  if (!res.ok) throw new Error("Failed to fetch products");
  return res.json();
}

export async function getProduct(id: string): Promise<ProductInfo> {
  const res = await fetch(`${CATALOG_URL}/api/v1/products/${id}`, {
    next: { tags: [`product-${id}`], revalidate: 3600 },
  });
  if (!res.ok) {
    if (res.status === 404) throw new Error("Product not found");
    throw new Error("Failed to fetch product");
  }
  return res.json();
}

export async function getProductBySlug(slug: string): Promise<ProductInfo> {
  const res = await fetch(`${CATALOG_URL}/api/v1/products/by-slug/${slug}`, {
    next: { tags: [`product-slug-${slug}`], revalidate: 3600 },
  });
  if (!res.ok) {
    if (res.status === 404) return null as unknown as ProductInfo;
    throw new Error("Failed to fetch product by slug");
  }
  return res.json();
}

export async function listCategories(params?: {
  page?: number;
  pageSize?: number;
}): Promise<{ categories: CategoryInfo[]; total: number }> {
  const url = new URL(`${CATALOG_URL}/api/v1/categories`);
  if (params?.page) url.searchParams.set("page", params.page.toString());
  if (params?.pageSize)
    url.searchParams.set("pageSize", params.pageSize.toString());

  const res = await fetch(url, {
    next: { tags: ["categories"], revalidate: 3600 },
  });
  if (!res.ok) throw new Error("Failed to fetch categories");
  return res.json();
}
