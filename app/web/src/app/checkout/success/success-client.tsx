"use client";

import Link from "next/link";
import { useEffect } from "react";
import { buttonVariants } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import { useCartStore } from "@/store/cart";
import type { CheckoutSession } from "@/lib/ucp/types/checkout";

const SHIPPING_CENTS = 599;

interface SuccessClientProps {
  session: CheckoutSession | null;
}

export function SuccessClient({ session }: SuccessClientProps) {
  const clearCart = useCartStore((s) => s.clearCart);

  useEffect(() => {
    clearCart();
  }, [clearCart]);

  if (!session) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-4 py-24 px-4">
        <h1 className="text-2xl font-bold">No order found</h1>
        <p className="text-muted-foreground">
          This order may have expired or the link is invalid.
        </p>
        <Link
          href="/products"
          className={cn(buttonVariants({ size: "lg" }), "mt-2")}
        >
          Continue shopping
        </Link>
      </div>
    );
  }

  const orderId = `ord_${session.id.replace(/-/g, "").slice(0, 8)}`;
  const totalCents = session.totals.subtotal_cents + SHIPPING_CENTS;
  const addr = session.shipping_address;

  return (
    <div className="flex-1 py-12 px-4">
      <div className="mx-auto max-w-xl space-y-8">
        <div className="text-center space-y-2">
          <div className="text-4xl">&#10003;</div>
          <h1 className="text-3xl font-bold">Order placed!</h1>
          <p className="text-muted-foreground text-sm">
            Order #{orderId}
          </p>
        </div>

        {addr && (
          <div className="rounded-xl border bg-card p-6 space-y-2">
            <h2 className="font-semibold text-sm">Ships to</h2>
            <p className="text-sm text-muted-foreground">
              {addr.line1}
              {addr.line2 ? `, ${addr.line2}` : ""}
            </p>
            <p className="text-sm text-muted-foreground">
              {addr.city}, {addr.state} {addr.postal_code}
            </p>
            <p className="text-sm text-muted-foreground">{addr.country}</p>
          </div>
        )}

        <div className="rounded-xl border bg-card p-6 space-y-4">
          <h2 className="font-semibold text-sm">Summary</h2>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Subtotal</span>
            <span>${(session.totals.subtotal_cents / 100).toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Shipping</span>
            <span>${(SHIPPING_CENTS / 100).toFixed(2)}</span>
          </div>
          <Separator />
          <div className="flex justify-between font-bold">
            <span>Total</span>
            <span>${(totalCents / 100).toFixed(2)}</span>
          </div>
        </div>

        <Link
          href="/products"
          className={cn(buttonVariants({ size: "lg" }), "w-full")}
        >
          Continue shopping
        </Link>
      </div>
    </div>
  );
}
