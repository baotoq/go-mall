import { z } from "zod";
import type { PaymentSummary } from "@/lib/ucp/types/checkout";

export const ShippingAddressSchema = z.object({
  line1: z.string().min(1).max(200),
  line2: z.string().max(200).optional(),
  city: z.string().min(1).max(100),
  state: z.string().min(1).max(100),
  postal_code: z.string().min(1).max(20),
  country: z.string().length(2),
});

export const PaymentInputSchema = z.object({
  card_number: z.string().regex(/^[\d\s]{12,19}$/),
  exp: z.string().regex(/^(0[1-9]|1[0-2])\/\d{2}$/),
  cvc: z.string().regex(/^\d{3,4}$/),
});

const BuyerInputSchema = z
  .object({
    email: z.string().email().optional(),
    name: z.string().min(1).max(120).optional(),
    phone: z.string().min(3).max(30).optional(),
  })
  .optional();

export const CreateCheckoutInputSchema = z.object({
  cart_session_id: z.string().min(1),
  currency: z.string().length(3),
  buyer: BuyerInputSchema,
  shipping_address: ShippingAddressSchema.optional(),
  payment: PaymentInputSchema.optional(),
});
export type CreateCheckoutInput = z.infer<typeof CreateCheckoutInputSchema>;

export const UpdateCheckoutInputSchema = z.object({
  buyer: BuyerInputSchema,
  shipping_address: ShippingAddressSchema.optional(),
  payment: PaymentInputSchema.optional(),
});
export type UpdateCheckoutInput = z.infer<typeof UpdateCheckoutInputSchema>;

// Idempotency-Key: 1–64 chars, alphanumeric + `-`
const IDEMPOTENCY_KEY_RE = /^[a-zA-Z0-9-]{1,64}$/;
export function validateIdempotencyKey(key: string | null | undefined): {
  valid: boolean;
  error?: string;
} {
  if (key === null || key === undefined || key === "") return { valid: true }; // optional
  if (!IDEMPOTENCY_KEY_RE.test(key))
    return {
      valid: false,
      error: "Idempotency-Key must be 1–64 alphanumeric chars or hyphens",
    };
  return { valid: true };
}

export function summarizePayment(p: { card_number: string }): PaymentSummary {
  const digits = p.card_number.replace(/\s/g, "");
  const last4 = digits.slice(-4);
  const brand = digits.startsWith("4")
    ? "visa"
    : /^(5[1-5]|2[2-7])/.test(digits)
      ? "mastercard"
      : /^3[47]/.test(digits)
        ? "amex"
        : "unknown";
  return { brand, last4 };
}
