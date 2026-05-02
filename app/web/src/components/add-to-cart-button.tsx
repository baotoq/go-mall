"use client";

import { Check, ShoppingCart } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import type { Product } from "@/lib/types";
import { useCartStore } from "@/store/cart";

export function AddToCartButton({ product }: { product: Product }) {
  const [added, setAdded] = useState(false);
  const addItem = useCartStore((state) => state.addItem);

  function handleAdd() {
    addItem({
      id: product.id,
      name: product.name,
      priceCents: product.priceCents,
      imageUrl: product.imageUrl,
    });
    setAdded(true);
    setTimeout(() => setAdded(false), 1500);
  }

  return (
    <Button size="lg" onClick={handleAdd} className="w-full sm:w-auto gap-2">
      {added ? (
        <>
          <Check className="size-4" />
          Added to Cart
        </>
      ) : (
        <>
          <ShoppingCart className="size-4" />
          Add to Cart
        </>
      )}
    </Button>
  );
}
