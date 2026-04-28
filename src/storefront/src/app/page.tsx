export const dynamic = "force-dynamic";

import Link from "next/link";
import { ProductCard } from "@/components/product-card";
import { Button } from "@/components/ui/button";
import { listProducts } from "@/lib/api/catalog";

export default async function Home() {
  const data = await listProducts({ pageSize: 4 });
  const featuredProducts = data.products || [];

  return (
    <div className="flex flex-col">
      {/* Hero Section */}
      <section className="relative flex min-h-[85vh] flex-col items-center justify-center overflow-hidden bg-surface text-center">
        <div className="absolute inset-0 z-0 opacity-40 bg-gradient-to-b from-surface-2 to-surface" />
        <div className="relative z-10 flex flex-col items-center px-4 max-w-4xl mt-[-8vh]">
          <h1 className="text-5xl md:text-7xl lg:text-[88px] font-semibold tracking-tighter text-ink leading-[1.05] mb-6">
            Pro power.
            <br />
            Everywhere.
          </h1>
          <p className="text-xl md:text-2xl text-ink-2 font-medium mb-10 max-w-2xl">
            The most advanced tech we have ever built. Now available in a
            lighter, faster, and more powerful design.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 items-center">
            <Link href="/shop">
              <Button
                size="lg"
                className="rounded-full h-14 px-8 text-base font-medium"
              >
                Shop all products
              </Button>
            </Link>
            <Link href="/shop?categoryId=mac">
              <Button
                variant="ghost"
                size="lg"
                className="rounded-full h-14 px-8 text-base font-medium border-ink/20 hover:bg-surface-2"
              >
                Learn more {">"}
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Featured Products */}
      <section className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8 py-24 md:py-32 border-t border-ink/10">
        <div className="flex items-end justify-between mb-12">
          <h2 className="text-3xl md:text-4xl font-semibold tracking-tight text-ink">
            New Arrivals
          </h2>
          <Link
            href="/shop"
            className="text-ink hover:text-ink/70 font-medium underline underline-offset-4 decoration-ink/30 transition-colors"
          >
            Shop the collection {">"}
          </Link>
        </div>
        <div className="grid grid-cols-1 gap-x-8 gap-y-12 sm:grid-cols-2 lg:grid-cols-4">
          {featuredProducts.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      </section>

      {/* Promo Section */}
      <section className="bg-surface-2 py-24 border-y border-ink/10">
        <div className="mx-auto w-full max-w-4xl px-4 text-center">
          <h2 className="text-3xl md:text-5xl font-semibold tracking-tight text-ink mb-6">
            Trade in your old device for credit.
          </h2>
          <p className="text-lg text-ink-2 mb-10">
            Get up to $500 toward your new purchase when you trade in an
            eligible device. Terms apply.
          </p>
          <Button variant="outline" className="rounded-full h-12 px-6">
            See how it works
          </Button>
        </div>
      </section>
    </div>
  );
}
