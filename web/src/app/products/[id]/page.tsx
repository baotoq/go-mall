import { notFound } from "next/navigation"
import Link from "next/link"
import { Star, ArrowLeft } from "lucide-react"
import { getProductById, categories } from "@/lib/mock-data"
import { AddToCartButton } from "@/components/add-to-cart-button"
import { cn } from "@/lib/utils"

export default async function ProductPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const product = getProductById(Number(id))
  if (!product) notFound()

  const category = categories.find((c) => c.id === product.category)

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
          <div className="aspect-square rounded-2xl bg-muted flex items-center justify-center text-[8rem] select-none">
            {product.emoji}
          </div>

          {/* Info */}
          <div className="space-y-5">
            <div className="space-y-1">
              {product.badge && (
                <span className="inline-block text-xs font-semibold bg-primary text-primary-foreground px-2.5 py-0.5 rounded-full mb-2">
                  {product.badge}
                </span>
              )}
              <p className="text-sm text-muted-foreground">
                {category?.emoji} {category?.name}
              </p>
              <h1 className="text-3xl font-bold">{product.name}</h1>
            </div>

            <div className="flex items-center gap-2">
              <div className="flex items-center gap-0.5">
                {[...Array(5)].map((_, i) => (
                  <Star
                    key={i}
                    className={cn(
                      "size-4",
                      i < Math.floor(product.rating)
                        ? "fill-amber-400 text-amber-400"
                        : "text-muted-foreground/30 fill-muted-foreground/30",
                    )}
                  />
                ))}
              </div>
              <span className="text-sm font-medium">{product.rating}</span>
              <span className="text-sm text-muted-foreground">
                ({product.reviews} reviews)
              </span>
            </div>

            <p className="text-4xl font-bold">${product.price}</p>

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
  )
}
