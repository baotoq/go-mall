"use client";

import Link from "next/link";
import { ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useCartStore } from "@/store/cart";
import type { Product } from "@/lib/types";

export function ProductCard({ product }: { product: Product }) {
  const addItem = useCartStore((state) => state.addItem);

  return (
    <div className="group rounded-xl border bg-card overflow-hidden hover:shadow-md transition-shadow">
      <Link href={`/products/${product.id}`}>
        <div className="aspect-square bg-muted flex items-center justify-center">
          {product.imageUrl ? (
            <img
              src={product.imageUrl}
              alt={product.name}
              className="w-full h-full object-cover"
            />
          ) : (
            <span className="text-5xl font-bold text-muted-foreground/20 select-none">
              {product.name.charAt(0).toUpperCase()}
            </span>
          )}
        </div>
      </Link>
      <div className="p-4 space-y-2">
        <Link href={`/products/${product.id}`}>
          <h3 className="font-medium leading-tight hover:underline line-clamp-1">
            {product.name}
          </h3>
        </Link>
        <p className="text-xs text-muted-foreground line-clamp-2">
          {product.description}
        </p>
        <div className="flex items-center justify-between pt-1">
          <span className="font-bold text-lg">
            ${(product.priceCents / 100).toFixed(2)}
          </span>
          <Button
            size="sm"
            onClick={() =>
              addItem({
                id: product.id,
                name: product.name,
                priceCents: product.priceCents,
                imageUrl: product.imageUrl,
              })
            }
          >
            <ShoppingCart className="size-3.5" />
            Add
          </Button>
        </div>
      </div>
    </div>
  );
}
