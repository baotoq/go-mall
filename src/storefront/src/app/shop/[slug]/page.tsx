export const dynamic = "force-dynamic";

import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { AddToCartButton } from "@/components/add-to-cart-button";
import { getProduct, listCategories } from "@/lib/api/catalog";

export default async function ProductPage(props: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await props.params;

  let product:
    | {
        id: string;
        name: string;
        price: number;
        description: string;
        imageUrl: string;
        categoryId: string;
      }
    | undefined;
  let categories: { id: string; name: string }[] = [];

  try {
    const [productData, categoriesData] = await Promise.all([
      getProduct(slug),
      listCategories(),
    ]);
    product = productData;
    categories = categoriesData.categories || [];
  } catch (_e) {
    notFound();
  }

  const category = categories.find(
    (c: { id: string; name: string }) => c.id === product.categoryId,
  );
  const categoryName = category ? category.name : "Products";

  return (
    <div className="flex flex-col min-h-screen bg-background">
      {/* Product Top Nav */}
      <div className="border-b border-ink/10 bg-surface/80 backdrop-blur-md sticky top-12 z-40">
        <div className="mx-auto flex h-12 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <h1 className="text-sm font-medium tracking-tight text-ink">
            {product.name}
          </h1>
          <div className="flex items-center gap-6">
            <span className="text-sm font-medium text-ink-2 hidden sm:inline-block">
              ${product.price.toFixed(2)}
            </span>
          </div>
        </div>
      </div>

      {/* Breadcrumbs */}
      <div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
        <div className="text-xs font-medium text-ink-2 flex items-center gap-2">
          <Link href="/" className="hover:text-ink transition-colors">
            Home
          </Link>
          <span>›</span>
          <Link
            href={`/shop?categoryId=${product.categoryId}`}
            className="hover:text-ink transition-colors"
          >
            {categoryName}
          </Link>
          <span>›</span>
          <span className="text-ink">{product.name}</span>
        </div>
      </div>

      <div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8 pb-24 md:pb-32">
        <div className="flex flex-col lg:flex-row gap-12 lg:gap-24">
          {/* Gallery Area */}
          <div className="flex-1 flex flex-col gap-6">
            <div className="relative aspect-square w-full overflow-hidden rounded-3xl bg-surface-2 md:aspect-[4/3] lg:aspect-square">
              <Image
                src={product.imageUrl}
                alt={product.name}
                fill
                priority
                className="object-cover"
                sizes="(max-width: 1024px) 100vw, 50vw"
              />
            </div>
            {/* Optional mini gallery thumbnails could go here */}
            <div className="grid grid-cols-4 gap-4">
              {[0, 1, 2].map((i) => (
                <div
                  key={i}
                  className="relative aspect-square rounded-xl bg-surface-2 overflow-hidden border-2 border-transparent hover:border-ink/20 transition-colors cursor-pointer"
                >
                  <Image
                    src={product.imageUrl}
                    alt=""
                    fill
                    className="object-cover opacity-60"
                    sizes="25vw"
                  />
                </div>
              ))}
            </div>
          </div>

          {/* Buy Box Area */}
          <div className="w-full lg:w-[380px] xl:w-[420px] flex-shrink-0 flex flex-col gap-8 lg:sticky lg:top-32 lg:self-start">
            <div>
              <p className="text-sm font-medium text-accent tracking-wide uppercase mb-2">
                New
              </p>
              <h1 className="text-3xl md:text-4xl font-semibold tracking-tight text-ink mb-4">
                {product.name}
              </h1>
              <p className="text-2xl md:text-3xl font-medium text-ink">
                ${product.price.toFixed(2)}
              </p>
            </div>

            <div className="prose prose-sm text-ink-2 max-w-none">
              <p className="leading-relaxed text-base">{product.description}</p>
            </div>

            {/* Dummy Configurator */}
            <div className="flex flex-col gap-4 border-t border-ink/10 pt-8">
              <p className="text-sm font-semibold text-ink">Color</p>
              <div className="flex gap-4">
                {["#1a1a1a", "#e0dcd1", "#2a5a8a"].map((color, i) => (
                  <button
                    type="button"
                    key={color}
                    className={`h-8 w-8 rounded-full border-2 focus:outline-none focus:ring-2 focus:ring-offset-2 transition-all ${i === 0 ? "border-accent ring-2 ring-offset-2 ring-accent/20" : "border-transparent ring-0 hover:scale-110"}`}
                    style={{ backgroundColor: color }}
                    aria-label={`Color option ${i}`}
                  />
                ))}
              </div>
            </div>

            <div className="border-t border-ink/10 pt-8 flex flex-col gap-4">
              <AddToCartButton productId={product.id} name={product.name} />

              <div className="bg-surface-2 rounded-2xl p-6 text-sm text-ink-2 space-y-4">
                <div className="flex justify-between">
                  <span className="font-medium text-ink">Free delivery</span>
                  <span>Delivers in 2-3 days</span>
                </div>
                <div className="flex justify-between">
                  <span className="font-medium text-ink">Free returns</span>
                  <span>Within 30 days</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Highlights / Features (Placeholder) */}
        <div className="mt-24 md:mt-32 pt-16 border-t border-ink/10">
          <h2 className="text-2xl md:text-3xl font-semibold tracking-tight text-ink mb-12 text-center">
            Every detail considered.
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex flex-col gap-4">
                <div className="h-48 bg-surface-2 rounded-2xl mb-2" />
                <h3 className="text-lg font-semibold text-ink">Feature {i}</h3>
                <p className="text-ink-2 text-sm">
                  Designed specifically for this product to provide the absolute
                  best experience possible in every way.
                </p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
