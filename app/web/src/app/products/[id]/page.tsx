import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { getProduct, listCategories } from "@/lib/api";
import { AddToCartButton } from "@/components/add-to-cart-button";

export default async function ProductPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const [product, categories] = await Promise.all([
    getProduct(id),
    listCategories(),
  ]);
  if (!product) notFound();

  const category = categories.find((c) => c.id === product.categoryId);

  return (
    <div className="flex-1 py-8 px-4">
      <div className="mx-auto max-w-6xl">
        <Link
          href="/products"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground mb-8 transition-colors"
        >
          <ArrowLeft className="size-4" />
          Back to Products
        </Link>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-12">
          {/* Image */}
          <div className="aspect-square rounded-2xl bg-muted flex items-center justify-center overflow-hidden">
            {product.imageUrl ? (
              <img
                src={product.imageUrl}
                alt={product.name}
                className="w-full h-full object-cover"
              />
            ) : (
              <span className="text-[8rem] font-bold text-muted-foreground/20 select-none">
                {product.name.charAt(0).toUpperCase()}
              </span>
            )}
          </div>

          {/* Info */}
          <div className="space-y-5">
            <div className="space-y-1">
              {category && (
                <p className="text-sm text-muted-foreground">{category.name}</p>
              )}
              <h1 className="text-3xl font-bold">{product.name}</h1>
            </div>

            <p className="text-4xl font-bold">
              ${(product.priceCents / 100).toFixed(2)}
            </p>

            {product.stock > 0 ? (
              <p className="text-sm text-green-600 font-medium">
                In stock ({product.stock} available)
              </p>
            ) : (
              <p className="text-sm text-destructive font-medium">
                Out of stock
              </p>
            )}

            <p className="text-muted-foreground leading-relaxed">
              {product.description}
            </p>

            <div className="pt-2">
              <AddToCartButton product={product} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
