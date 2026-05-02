import { z } from "zod";

export const CreateCheckoutInputSchema = z.object({
  cart_session_id: z.string().min(1),
  currency: z.string().length(3),
  buyer: z.object({ email: z.string().email().optional() }).optional(),
});
export type CreateCheckoutInput = z.infer<typeof CreateCheckoutInputSchema>;

export const UpdateCheckoutInputSchema = z.object({
  buyer: z.object({ email: z.string().email() }).optional(),
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
