export const dynamic = "force-dynamic";

import Link from "next/link";
import { ProductCard } from "@/components/product-card";
import { Button } from "@/components/ui/button";
import {
  listCategories,
  listProducts,
  type ProductInfo,
} from "@/lib/api/catalog";

export default async function ShopPage(props: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const searchParams = await props.searchParams;
  const categoryId = searchParams.categoryId as string | undefined;

  let products: ProductInfo[] = [];
  let categoryName = "All Products";
  let categories: { id: string; name: string }[] = [];

  try {
    const [productsData, categoriesData] = await Promise.all([
      listProducts({ categoryId: categoryId, pageSize: 100 }),
      listCategories(),
    ]);
    products = productsData.products || [];
    categories = categoriesData.categories || [];

    if (categoryId) {
      const cat = categories.find((c) => c.id === categoryId);
      if (cat) categoryName = cat.name;
    }
  } catch (e) {
    console.error("Failed to fetch shop data", e);
  }

  return (
    <div className="flex flex-col min-h-screen bg-background">
      <div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8 py-12 md:py-20">
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12">
          <div>
            <div className="text-sm font-medium text-ink-2 mb-4">
              <Link href="/" className="hover:text-ink transition-colors">
                Home
              </Link>
              <span className="mx-2">›</span>
              <span className="text-ink">{categoryName}</span>
            </div>
            <h1 className="text-4xl md:text-5xl lg:text-6xl font-semibold tracking-tighter text-ink">
              {categoryName}.
            </h1>
          </div>
          <p className="text-base text-ink-2 font-medium">
            {products.length} products
          </p>
        </div>

        {/* Filter Bar (UI only) */}
        <div className="flex flex-wrap items-center gap-3 py-6 border-y border-ink/10 mb-12">
          {categories.map((c) => (
            <Link key={c.id} href={`/shop?categoryId=${c.id}`}>
              <Button
                variant={categoryId === c.id ? "default" : "outline"}
                className={`rounded-full h-10 px-5 text-sm font-medium transition-all ${categoryId === c.id ? "bg-ink text-surface hover:bg-ink/90 border-transparent" : "bg-transparent border-ink/20 text-ink hover:border-ink/40"}`}
              >
                {c.name}
              </Button>
            </Link>
          ))}
          {categoryId && (
            <Link href="/shop">
              <Button
                variant="ghost"
                className="rounded-full h-10 px-5 text-sm font-medium text-ink-2 hover:text-ink"
              >
                Clear Filters
              </Button>
            </Link>
          )}
        </div>

        {/* Product Grid */}
        {products.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-x-8 gap-y-16">
            {products.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        ) : (
          <div className="py-32 text-center">
            <h3 className="text-2xl font-semibold text-ink mb-4">
              No products found
            </h3>
            <p className="text-ink-2 mb-8">
              We couldn't find anything matching your criteria.
            </p>
            <Link href="/shop">
              <Button className="rounded-full h-12 px-8">
                View all products
              </Button>
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}
