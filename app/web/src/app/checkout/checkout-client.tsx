"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { Separator } from "@/components/ui/separator";
import { useCartStore } from "@/store/cart";
import { OrderSummary } from "./order-summary";
import { ContactSection } from "./sections/contact-section";
import { PaymentSection } from "./sections/payment-section";
import { ShippingSection } from "./sections/shipping-section";

const CheckoutFormSchema = z.object({
  // Contact
  email: z.string().email({ message: "Valid email is required" }),
  name: z.string().min(1, { message: "Name is required" }),
  phone: z.string().optional(),
  // Shipping
  line1: z.string().min(1, { message: "Address is required" }),
  line2: z.string().optional(),
  city: z.string().min(1, { message: "City is required" }),
  state: z.string().min(1, { message: "State is required" }),
  postal_code: z.string().min(1, { message: "Postal code is required" }),
  country: z.string().length(2, { message: "2-letter country code required" }),
  // Payment
  card_number: z
    .string()
    .regex(/^[\d\s]{12,19}$/, { message: "Enter a valid card number" }),
  exp: z
    .string()
    .regex(/^(0[1-9]|1[0-2])\/\d{2}$/, { message: "Format: MM/YY" }),
  cvc: z.string().regex(/^\d{3,4}$/, { message: "3 or 4 digits" }),
});

export type CheckoutFormValues = z.infer<typeof CheckoutFormSchema>;

interface CheckoutClientProps {
  defaultEmail: string;
}

function getSessionId(): string {
  if (typeof window === "undefined") return "";
  return localStorage.getItem("cart_session_id") ?? "";
}

export function CheckoutClient({ defaultEmail }: CheckoutClientProps) {
  const router = useRouter();
  const { items, isLoading, loadCart } = useCartStore();
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    loadCart();
  }, [loadCart]);

  useEffect(() => {
    if (!isLoading && items.length === 0) {
      router.push("/cart");
    }
  }, [isLoading, items.length, router]);

  const form = useForm<CheckoutFormValues>({
    resolver: zodResolver(CheckoutFormSchema),
    defaultValues: {
      email: defaultEmail,
      name: "",
      phone: "",
      line1: "",
      line2: "",
      city: "",
      state: "",
      postal_code: "",
      country: "",
      card_number: "",
      exp: "",
      cvc: "",
    },
  });

  const isSubmitting = form.formState.isSubmitting;

  async function onSubmit(values: CheckoutFormValues) {
    setSubmitError(null);
    const idempotencyKey = crypto.randomUUID();
    const cartSessionId = getSessionId();

    // Step 1: Create checkout session
    let sessionId: string;
    try {
      const res = await fetch("/api/ucp/checkout", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": idempotencyKey,
        },
        body: JSON.stringify({
          cart_session_id: cartSessionId,
          currency: "USD",
          buyer: {
            email: values.email,
            name: values.name,
            phone: values.phone || undefined,
          },
          shipping_address: {
            line1: values.line1,
            line2: values.line2 || undefined,
            city: values.city,
            state: values.state,
            postal_code: values.postal_code,
            country: values.country,
          },
          payment: {
            card_number: values.card_number,
            exp: values.exp,
            cvc: values.cvc,
          },
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        const msg =
          data?.error?.content ??
          data?.messages?.[0]?.content ??
          `Checkout failed (${res.status})`;
        setSubmitError(msg);
        return;
      }

      const data = await res.json();
      sessionId = data.session_id ?? data.id;
    } catch {
      setSubmitError("Network error. Please try again.");
      return;
    }

    // Step 2: Complete the session
    try {
      const res = await fetch(
        `/api/ucp/checkout/${sessionId}?action=complete`,
        {
          method: "POST",
          headers: { "X-UCP-Session": sessionId },
        },
      );

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        const msg =
          data?.error?.content ??
          data?.messages?.[0]?.content ??
          `Order completion failed (${res.status})`;
        setSubmitError(msg);
        return;
      }
    } catch {
      setSubmitError("Network error completing order. Please try again.");
      return;
    }

    router.push(`/checkout/success?id=${sessionId}`);
  }

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center py-24">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="flex-1 py-8 px-4">
      <div className="mx-auto max-w-5xl">
        <h1 className="text-3xl font-bold mb-8">Checkout</h1>

        {submitError && (
          <div
            role="alert"
            className="mb-6 rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive"
          >
            {submitError}
          </div>
        )}

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="grid gap-8 md:grid-cols-[1fr_360px]"
          >
            {/* Left column: form sections */}
            <div className="space-y-8">
              <ContactSection control={form.control} />
              <Separator />
              <ShippingSection control={form.control} />
              <Separator />
              <PaymentSection control={form.control} />

              <Button
                type="submit"
                size="lg"
                className="w-full"
                disabled={isSubmitting}
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" />
                    Processing…
                  </>
                ) : (
                  "Place order"
                )}
              </Button>
            </div>

            {/* Right column: order summary */}
            <div>
              <OrderSummary items={items} />
            </div>
          </form>
        </Form>
      </div>
    </div>
  );
}
