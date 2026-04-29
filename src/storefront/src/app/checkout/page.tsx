"use client";

import { Loader2 } from "lucide-react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { getProduct, type ProductInfo } from "@/lib/api/catalog";
import { useCart } from "@/lib/store/cart";
import { processCheckout } from "../actions";

export default function CheckoutPage() {
  const { items, clear, initialized } = useCart();
  const [products, setProducts] = useState<Record<string, ProductInfo>>({});
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const fetchedIds = useRef<Set<string>>(new Set());
  const router = useRouter();

  useEffect(() => {
    if (!initialized) return;
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
        } catch (_e) {}
      }
      setLoading(false);
    };
    fetchProducts();
  }, [items, initialized]);

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

  if (items.length === 0) {
    router.push("/cart");
    return null;
  }

  const subtotal = items.reduce((acc, item) => {
    const product = products[item.productId];
    return acc + (product ? product.price * item.quantity : 0);
  }, 0);
  const tax = subtotal * 0.08;
  const total = subtotal + tax;

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setProcessing(true);

    const lineItems = items.map((item) => ({
      productId: item.productId,
      quantity: item.quantity,
      price: products[item.productId]?.price || 0,
    }));

    const result = await processCheckout({
      totalAmount: total,
      currency: "USD",
      items: lineItems,
    });

    setProcessing(false);

    if (result.success) {
      toast.success("Order placed successfully!");
      clear();
      router.push("/account");
    } else {
      toast.error(result.error || "Failed to place order.");
    }
  };

  return (
    <div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8 py-12 md:py-20">
      <div className="mb-10 text-center">
        <h1 className="text-3xl md:text-5xl font-semibold tracking-tighter text-ink mb-4">
          Checkout.
        </h1>
        <p className="text-base text-ink-2">Complete your order securely.</p>
      </div>

      <div className="flex flex-col lg:flex-row gap-12 lg:gap-24">
        {/* Checkout Form */}
        <div className="flex-1 max-w-2xl">
          <form onSubmit={handleSubmit} className="flex flex-col gap-10">
            {/* Contact */}
            <div className="flex flex-col gap-6">
              <h2 className="text-2xl font-semibold tracking-tight text-ink">
                Contact Information
              </h2>
              <div className="grid gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="you@example.com"
                    required
                    className="h-12 rounded-xl"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="phone">Phone</Label>
                  <Input
                    id="phone"
                    type="tel"
                    placeholder="(555) 123-4567"
                    required
                    className="h-12 rounded-xl"
                  />
                </div>
              </div>
            </div>

            {/* Shipping */}
            <div className="flex flex-col gap-6">
              <h2 className="text-2xl font-semibold tracking-tight text-ink">
                Shipping Address
              </h2>
              <div className="grid gap-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="firstName">First Name</Label>
                    <Input
                      id="firstName"
                      required
                      className="h-12 rounded-xl"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="lastName">Last Name</Label>
                    <Input id="lastName" required className="h-12 rounded-xl" />
                  </div>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="address">Address</Label>
                  <Input
                    id="address"
                    placeholder="123 Apple Park Way"
                    required
                    className="h-12 rounded-xl"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="city">City</Label>
                    <Input
                      id="city"
                      placeholder="Cupertino"
                      required
                      className="h-12 rounded-xl"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="state">State / Province</Label>
                    <Input
                      id="state"
                      placeholder="CA"
                      required
                      className="h-12 rounded-xl"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="zip">ZIP / Postal Code</Label>
                    <Input
                      id="zip"
                      placeholder="95014"
                      required
                      className="h-12 rounded-xl"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="country">Country</Label>
                    <Input
                      id="country"
                      defaultValue="United States"
                      required
                      className="h-12 rounded-xl"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Payment Info */}
            <div className="flex flex-col gap-6">
              <h2 className="text-2xl font-semibold tracking-tight text-ink">
                Payment Method
              </h2>
              <div className="grid gap-4 bg-surface-2 p-6 rounded-2xl">
                <div className="grid gap-2">
                  <Label htmlFor="card">Card Number</Label>
                  <Input
                    id="card"
                    placeholder="0000 0000 0000 0000"
                    required
                    className="h-12 rounded-xl bg-background"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="exp">Expiration</Label>
                    <Input
                      id="exp"
                      placeholder="MM/YY"
                      required
                      className="h-12 rounded-xl bg-background"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="cvv">CVV</Label>
                    <Input
                      id="cvv"
                      placeholder="123"
                      required
                      className="h-12 rounded-xl bg-background"
                    />
                  </div>
                </div>
              </div>
            </div>

            <Button
              type="submit"
              disabled={processing}
              className="w-full rounded-full h-14 text-lg font-medium transition-all active:scale-[0.98]"
            >
              {processing ? (
                <Loader2 className="mr-2 h-5 w-5 animate-spin" />
              ) : (
                `Pay $${total.toFixed(2)}`
              )}
            </Button>
          </form>
        </div>

        {/* Order Summary sidebar */}
        <div className="w-full lg:w-[380px] shrink-0">
          <div className="bg-surface-2 rounded-3xl p-8 sticky top-24">
            <h2 className="text-2xl font-semibold tracking-tight text-ink mb-6">
              In Your Bag
            </h2>

            <div className="flex flex-col gap-4 mb-6 pb-6 border-b border-ink/10">
              {items.map((item) => {
                const product = products[item.productId];
                if (!product) return null;
                return (
                  <div key={item.productId} className="flex gap-4">
                    <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-xl bg-background">
                      <Image
                        src={product.imageUrl}
                        alt={product.name}
                        fill
                        className="object-cover"
                        sizes="64px"
                      />
                    </div>
                    <div className="flex flex-1 flex-col justify-center">
                      <h4 className="text-sm font-medium text-ink line-clamp-1">
                        {product.name}
                      </h4>
                      <p className="text-xs text-ink-2">Qty {item.quantity}</p>
                    </div>
                    <div className="text-sm font-medium text-ink flex items-center">
                      ${(product.price * item.quantity).toFixed(2)}
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="flex flex-col gap-3 text-sm mb-6 pb-6 border-b border-ink/10">
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

            <div className="flex justify-between text-lg font-semibold text-ink">
              <span>Total</span>
              <span className="tabular-nums">${total.toFixed(2)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
