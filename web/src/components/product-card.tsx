"use client"

import Link from "next/link"
import { Star, ShoppingCart } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useCartStore } from "@/store/cart"
import type { Product } from "@/lib/mock-data"

export function ProductCard({ product }: { product: Product }) {
  const addItem = useCartStore((state) => state.addItem)

  return (
    <div className="group rounded-xl border bg-card overflow-hidden hover:shadow-md transition-shadow">
      <Link href={`/products/${product.id}`}>
        <div className="aspect-square bg-muted flex items-center justify-center text-7xl select-none">
          {product.emoji}
        </div>
      </Link>
      <div className="p-4 space-y-2">
        {product.badge && (
          <span className="inline-block text-xs font-semibold bg-primary text-primary-foreground px-2 py-0.5 rounded-full">
            {product.badge}
          </span>
        )}
        <Link href={`/products/${product.id}`}>
          <h3 className="font-medium leading-tight hover:underline line-clamp-1">
            {product.name}
          </h3>
        </Link>
        <div className="flex items-center gap-1 text-xs text-muted-foreground">
          <Star className="size-3 fill-amber-400 text-amber-400" />
          <span className="font-medium">{product.rating}</span>
          <span>({product.reviews})</span>
        </div>
        <div className="flex items-center justify-between pt-1">
          <span className="font-bold text-lg">${product.price}</span>
          <Button
            size="sm"
            onClick={() =>
              addItem({
                id: product.id,
                name: product.name,
                price: product.price,
                emoji: product.emoji,
              })
            }
          >
            <ShoppingCart className="size-3.5" />
            Add
          </Button>
        </div>
      </div>
    </div>
  )
}
