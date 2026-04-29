"use client";

import { Minus, Plus, Trash2 } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { getProduct, type ProductInfo } from "@/lib/api/catalog";
import { useCart } from "@/lib/store/cart";
import { Button } from "./ui/button";

interface CartSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CartSheet({ open, onOpenChange }: CartSheetProps) {
  const { items, updateQuantity, removeItem } = useCart();
  const [products, setProducts] = useState<Record<string, ProductInfo>>({});
  const [loading, setLoading] = useState(false);
  const fetchedIds = useRef<Set<string>>(new Set());

  useEffect(() => {
    if (!open) return;
    const fetchProducts = async () => {
      setLoading(true);
      const toFetch = items.filter((item) => !fetchedIds.current.has(item.productId));
      if (toFetch.length > 0) {
        try {
          const fetched = await Promise.all(
            toFetch.map((item) => getProduct(item.productId).catch(() => null)),
          );
          const updates: Record<string, ProductInfo> = {};
          fetched.forEach((p) => {
            if (p) {
              updates[p.id] = p;
              fetchedIds.current.add(p.id);
            }
          });
          setProducts((prev) => ({ ...prev, ...updates }));
        } catch (_e) {
          // ignore
        }
      }
      setLoading(false);
    };
    fetchProducts();
  }, [open, items]);

  const total = items.reduce((acc, item) => {
    const product = products[item.productId];
    return acc + (product ? product.price * item.quantity : 0);
  }, 0);

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex w-full flex-col sm:max-w-md bg-background/80 backdrop-blur-xl border-l border-white/10 p-6">
        <SheetHeader>
          <SheetTitle className="text-2xl font-semibold tracking-tight text-ink">
            Your Bag
          </SheetTitle>
          <SheetDescription className="sr-only">
            Items in your cart
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto py-6">
          {items.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center text-center">
              <p className="text-ink/60 mb-4">Your bag is empty.</p>
              <Button
                variant="outline"
                className="rounded-full px-8 py-2 h-10 border-ink/20 hover:border-ink/40"
                onClick={() => onOpenChange(false)}
              >
                Continue Shopping
              </Button>
            </div>
          ) : (
            <div className="flex flex-col gap-6">
              {items.map((item) => {
                const product = products[item.productId];
                if (!product && loading)
                  return (
                    <div
                      key={item.productId}
                      className="h-24 w-full animate-pulse bg-surface-2 rounded-xl"
                    />
                  );
                if (!product) return null;

                return (
                  <div key={item.productId} className="flex gap-4 group">
                    <div className="relative h-24 w-24 shrink-0 overflow-hidden rounded-xl bg-surface-3">
                      <Image
                        src={product.imageUrl}
                        alt={product.name}
                        fill
                        className="object-cover"
                        sizes="(max-width: 768px) 100vw, 96px"
                      />
                    </div>
                    <div className="flex flex-1 flex-col justify-between">
                      <div className="flex justify-between">
                        <div>
                          <h3 className="font-semibold text-ink text-[15px] leading-tight line-clamp-1">
                            {product.name}
                          </h3>
                          <p className="text-sm text-ink-2 mt-1 tabular-nums">
                            ${product.price.toFixed(2)}
                          </p>
                        </div>
                        <button
                          type="button"
                          aria-label={`Remove ${product.name} from cart`}
                          className="text-ink-3 hover:text-destructive transition-colors opacity-0 group-hover:opacity-100 p-1 -m-1"
                          onClick={() => removeItem(item.productId)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                      <div className="flex items-center gap-3 bg-surface-2 w-fit rounded-full px-2 py-1">
                        <button
                          type="button"
                          className="flex h-6 w-6 items-center justify-center rounded-full bg-background text-ink shadow-sm hover:scale-95 transition-transform"
                          onClick={() =>
                            updateQuantity(
                              item.productId,
                              Math.max(1, item.quantity - 1),
                            )
                          }
                        >
                          <Minus className="h-3 w-3" />
                        </button>
                        <span className="w-4 text-center text-sm font-medium tabular-nums">
                          {item.quantity}
                        </span>
                        <button
                          type="button"
                          className="flex h-6 w-6 items-center justify-center rounded-full bg-background text-ink shadow-sm hover:scale-95 transition-transform"
                          onClick={() =>
                            updateQuantity(item.productId, item.quantity + 1)
                          }
                        >
                          <Plus className="h-3 w-3" />
                        </button>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {items.length > 0 && (
          <div className="border-t border-ink/10 pt-6">
            <div className="flex justify-between text-lg font-semibold text-ink mb-6">
              <span>Estimated Total</span>
              <span className="tabular-nums">${total.toFixed(2)}</span>
            </div>
            <Link href="/cart" onClick={() => onOpenChange(false)}>
              <Button className="w-full rounded-full h-12 text-base font-medium bg-accent hover:bg-accent-hover text-white transition-colors active:scale-[0.96]">
                Review Bag
              </Button>
            </Link>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
