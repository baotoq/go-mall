"use client";

import { ArrowLeft, ShoppingBag, Trash2 } from "lucide-react";
import Link from "next/link";
import { useEffect } from "react";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useCartStore } from "@/store/cart";

export function CartClient() {
  const {
    items,
    isLoading,
    removeItem,
    updateQuantity,
    totalItems,
    totalPrice,
    loadCart,
  } = useCartStore();

  useEffect(() => {
    loadCart();
  }, [loadCart]);

  if (isLoading) {
    return (
      <div className="flex-1 py-8 px-4">
        <div className="mx-auto max-w-3xl space-y-4">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="flex items-center gap-4 rounded-xl border bg-card p-4 animate-pulse"
            >
              <div className="size-16 rounded-lg bg-muted shrink-0" />
              <div className="flex-1 space-y-2">
                <div className="h-4 w-40 rounded bg-muted" />
                <div className="h-3 w-20 rounded bg-muted" />
              </div>
              <div className="h-8 w-24 rounded bg-muted" />
              <div className="h-4 w-16 rounded bg-muted" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-4 py-24 px-4">
        <ShoppingBag className="size-16 text-muted-foreground/30" />
        <h1 className="text-2xl font-bold">Your cart is empty</h1>
        <p className="text-muted-foreground">
          Add some products to get started.
        </p>
        <Link
          href="/products"
          className={cn(buttonVariants({ size: "lg" }), "mt-2")}
        >
          Browse Products
        </Link>
      </div>
    );
  }

  return (
    <div className="flex-1 py-8 px-4">
      <div className="mx-auto max-w-3xl">
        <Link
          href="/products"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground mb-8 transition-colors"
        >
          <ArrowLeft className="size-4" />
          Continue Shopping
        </Link>

        <h1 className="text-3xl font-bold mb-8">
          Cart ({totalItems()} item{totalItems() !== 1 ? "s" : ""})
        </h1>

        <div className="space-y-4 mb-8">
          {items.map((item) => (
            <div
              key={item.id}
              className="flex items-center gap-4 rounded-xl border bg-card p-4"
            >
              <div className="size-16 rounded-lg bg-muted flex items-center justify-center shrink-0 overflow-hidden">
                {item.imageUrl ? (
                  <img
                    src={item.imageUrl}
                    alt={item.name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <span className="text-xl font-bold text-muted-foreground/30 select-none">
                    {item.name.charAt(0).toUpperCase()}
                  </span>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <Link
                  href={`/products/${item.id}`}
                  className="font-medium hover:underline line-clamp-1"
                >
                  {item.name}
                </Link>
                <p className="text-sm text-muted-foreground">
                  ${(item.priceCents / 100).toFixed(2)} each
                </p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <Button
                  variant="outline"
                  size="icon-sm"
                  onClick={() => updateQuantity(item.id, item.quantity - 1)}
                >
                  −
                </Button>
                <span className="w-6 text-center text-sm font-medium">
                  {item.quantity}
                </span>
                <Button
                  variant="outline"
                  size="icon-sm"
                  onClick={() => updateQuantity(item.id, item.quantity + 1)}
                >
                  +
                </Button>
              </div>
              <p className="w-20 text-right font-bold shrink-0">
                ${((item.priceCents / 100) * item.quantity).toFixed(2)}
              </p>
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => removeItem(item.id)}
                className="text-muted-foreground hover:text-destructive shrink-0"
                data-testid="remove-item"
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
          ))}
        </div>

        {/* Summary */}
        <div className="rounded-xl border bg-card p-6 space-y-4">
          <h2 className="font-semibold text-lg">Order Summary</h2>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">
              Subtotal ({totalItems()} items)
            </span>
            <span>${totalPrice().toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Shipping</span>
            <span className="text-green-600 font-medium">Free</span>
          </div>
          <div className="border-t pt-4 flex justify-between font-bold text-lg">
            <span>Total</span>
            <span>${totalPrice().toFixed(2)}</span>
          </div>
          <Link
            href="/checkout"
            className={cn(buttonVariants({ size: "lg" }), "w-full")}
          >
            Checkout
          </Link>
        </div>
      </div>
    </div>
  );
}
