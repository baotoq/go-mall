import Link from "next/link";
import { ProductCard } from "@/components/product-card";
import { listProducts, listCategories } from "@/lib/api";
import { cn } from "@/lib/utils";

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<{ category?: string }>;
}) {
  const { category } = await searchParams;
  const [{ products }, categories] = await Promise.all([
    listProducts({ categoryId: category }),
    listCategories(),
  ]);

  return (
    <div className="flex-1 py-8 px-4">
      <div className="mx-auto max-w-6xl">
        <h1 className="text-3xl font-bold mb-8">Products</h1>
        <div className="flex gap-8">
          {/* Sidebar */}
          <aside className="w-44 shrink-0">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
              Categories
            </p>
            <nav className="space-y-1">
              <Link
                href="/products"
                className={cn(
                  "block px-3 py-2 rounded-lg text-sm transition-colors",
                  !category
                    ? "bg-primary text-primary-foreground font-medium"
                    : "hover:bg-muted",
                )}
              >
                All Products
              </Link>
              {categories.map((cat) => (
                <Link
                  key={cat.id}
                  href={`/products?category=${cat.id}`}
                  className={cn(
                    "block px-3 py-2 rounded-lg text-sm transition-colors",
                    category === cat.id
                      ? "bg-primary text-primary-foreground font-medium"
                      : "hover:bg-muted",
                  )}
                >
                  {cat.name}
                </Link>
              ))}
            </nav>
          </aside>

          {/* Grid */}
          <div className="flex-1 min-w-0">
            <p className="text-sm text-muted-foreground mb-4">
              {products.length} product{products.length !== 1 ? "s" : ""}
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {products.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
