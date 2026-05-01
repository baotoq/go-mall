import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { ProductCard } from "@/components/product-card";
import { listProducts, listCategories } from "@/lib/api";
import { cn } from "@/lib/utils";

export default async function Home() {
  const [{ products: featured }, categories] = await Promise.all([
    listProducts({ pageSize: 4 }),
    listCategories(),
  ]);

  return (
    <div className="flex-1">
      {/* Hero */}
      <section className="bg-primary text-primary-foreground py-24 px-4">
        <div className="mx-auto max-w-6xl">
          <div className="max-w-xl">
            <h1 className="text-5xl font-bold tracking-tight mb-4">
              Shop the Best Deals
            </h1>
            <p className="text-lg text-primary-foreground/70 mb-8">
              Discover thousands of products across electronics, clothing,
              books, and more — all in one place.
            </p>
            <Link
              href="/products"
              className={cn(
                buttonVariants({ size: "lg" }),
                "bg-background text-foreground hover:bg-background/90 gap-2",
              )}
            >
              Shop Now
              <ArrowRight className="size-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* Categories */}
      <section className="py-16 px-4">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-2xl font-bold mb-8">Shop by Category</h2>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            {categories.map((cat) => (
              <Link
                key={cat.id}
                href={`/products?category=${cat.id}`}
                className="rounded-xl border bg-card p-6 text-center hover:shadow-md hover:border-primary/30 transition-all"
              >
                <div className="text-4xl font-bold text-muted-foreground/20 mb-3 select-none">
                  {cat.name.charAt(0)}
                </div>
                <div className="font-semibold">{cat.name}</div>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* Featured Products */}
      <section className="py-16 px-4 bg-muted/30">
        <div className="mx-auto max-w-6xl">
          <div className="flex items-center justify-between mb-8">
            <h2 className="text-2xl font-bold">Featured Products</h2>
            <Link
              href="/products"
              className={cn(
                buttonVariants({ variant: "outline", size: "sm" }),
                "gap-1.5",
              )}
            >
              View All
              <ArrowRight className="size-3.5" />
            </Link>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {featured.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-8 px-4 text-center text-sm text-muted-foreground">
        © 2026 GoMall. All rights reserved.
      </footer>
    </div>
  );
}
