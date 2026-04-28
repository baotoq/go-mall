"use client";

import { Loader2, Minus, Plus } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { getProduct, type ProductInfo } from "@/lib/api/catalog";
import { useCart } from "@/lib/store/cart";

export default function CartPage() {
  const { items, updateQuantity, removeItem, initialized } = useCart();
  const [products, setProducts] = useState<Record<string, ProductInfo>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!initialized) return;

    const fetchProducts = async () => {
      setLoading(true);
      const toFetch = items.filter((item) => !products[item.productId]);

      if (toFetch.length > 0) {
        try {
          const fetched = await Promise.all(
            toFetch.map((item) => getProduct(item.productId).catch(() => null)),
          );
          const newProducts = { ...products };
          fetched.forEach((p) => {
            if (p) newProducts[p.id] = p;
          });
          setProducts(newProducts);
        } catch (e) {
          console.error(e);
        }
      }
      setLoading(false);
    };

    fetchProducts();
  }, [items, products, initialized]);

  if (
    !initialized ||
    (loading && items.length > 0 && Object.keys(products).length === 0)
  ) {
    return (
      <div className="flex h-[60vh] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-ink/40" />
      </div>
    );
  }

  const subtotal = items.reduce((acc, item) => {
    const product = products[item.productId];
    return acc + (product ? product.price * item.quantity : 0);
  }, 0);

  const tax = subtotal * 0.08;
  const total = subtotal + tax;

  if (items.length === 0) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-24 text-center">
        <h1 className="text-4xl font-semibold tracking-tight text-ink mb-6">
          Your bag is empty.
        </h1>
        <p className="text-ink-2 mb-8">Free delivery and free returns.</p>
        <Link href="/shop">
          <Button className="rounded-full h-12 px-8 text-base font-medium">
            Continue Shopping
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8 py-16 md:py-24">
      <div className="mb-12">
        <h1 className="text-4xl md:text-5xl font-semibold tracking-tighter text-ink mb-4">
          Review your bag.
        </h1>
        <p className="text-lg text-ink-2">Free delivery and free returns.</p>
      </div>

      <div className="flex flex-col lg:flex-row gap-16">
        {/* Cart Items */}
        <div className="flex-1 flex flex-col gap-8">
          {items.map((item) => {
            const product = products[item.productId];
            if (!product) return null;

            return (
              <div
                key={item.productId}
                className="flex flex-col sm:flex-row gap-6 pb-8 border-b border-ink/10"
              >
                <div className="relative aspect-square w-full sm:w-40 sm:h-40 shrink-0 overflow-hidden rounded-2xl bg-surface-2">
                  <Image
                    src={product.imageUrl}
                    alt={product.name}
                    fill
                    className="object-cover"
                    sizes="(max-width: 640px) 100vw, 160px"
                  />
                </div>

                <div className="flex flex-1 flex-col justify-between">
                  <div className="flex flex-col sm:flex-row sm:justify-between gap-4">
                    <div>
                      <h3 className="text-xl font-semibold tracking-tight text-ink">
                        {product.name}
                      </h3>
                      <p className="text-sm text-ink-2 mt-1">In Stock</p>
                    </div>
                    <div className="text-xl font-medium text-ink tabular-nums">
                      ${(product.price * item.quantity).toFixed(2)}
                    </div>
                  </div>

                  <div className="flex items-center justify-between mt-6 sm:mt-0">
                    <div className="flex items-center gap-4 bg-surface-2 rounded-full px-2 py-1">
                      <button
                        type="button"
                        className="flex h-8 w-8 items-center justify-center rounded-full bg-background text-ink shadow-sm hover:scale-95 transition-transform"
                        onClick={() =>
                          updateQuantity(
                            item.productId,
                            Math.max(1, item.quantity - 1),
                          )
                        }
                        disabled={item.quantity <= 1}
                      >
                        <Minus className="h-4 w-4" />
                      </button>
                      <span className="w-6 text-center text-sm font-medium tabular-nums">
                        {item.quantity}
                      </span>
                      <button
                        type="button"
                        className="flex h-8 w-8 items-center justify-center rounded-full bg-background text-ink shadow-sm hover:scale-95 transition-transform"
                        onClick={() =>
                          updateQuantity(item.productId, item.quantity + 1)
                        }
                      >
                        <Plus className="h-4 w-4" />
                      </button>
                    </div>

                    <button
                      type="button"
                      className="text-sm font-medium text-ink hover:text-accent transition-colors underline underline-offset-4 decoration-ink/20"
                      onClick={() => removeItem(item.productId)}
                    >
                      Remove
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Order Summary */}
        <div className="w-full lg:w-[380px] shrink-0">
          <div className="bg-surface-2 rounded-3xl p-8 sticky top-24">
            <h2 className="text-2xl font-semibold tracking-tight text-ink mb-6">
              Order Summary
            </h2>

            <div className="flex flex-col gap-4 text-base mb-6 pb-6 border-b border-ink/10">
              <div className="flex justify-between text-ink-2">
                <span>Subtotal</span>
                <span className="text-ink tabular-nums">
                  ${subtotal.toFixed(2)}
                </span>
              </div>
              <div className="flex justify-between text-ink-2">
                <span>Shipping</span>
                <span className="text-ink">Free</span>
              </div>
              <div className="flex justify-between text-ink-2">
                <span>Estimated Tax</span>
                <span className="text-ink tabular-nums">${tax.toFixed(2)}</span>
              </div>
            </div>

            <div className="flex justify-between text-xl font-semibold text-ink mb-8">
              <span>Total</span>
              <span className="tabular-nums">${total.toFixed(2)}</span>
            </div>

            <Link href="/checkout">
              <Button className="w-full rounded-full h-14 text-base font-medium transition-all active:scale-[0.98]">
                Check Out
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
