"use server";

import {
  type CreatePaymentRequest,
  createPayment,
  type PaymentResponse,
} from "@/lib/api/payment";

export async function processCheckout(
  req: CreatePaymentRequest,
): Promise<{ success: boolean; data?: PaymentResponse; error?: string }> {
  try {
    const payment = await createPayment(req);
    return { success: true, data: payment };
  } catch (error: unknown) {
    return {
      success: false,
      error: (error as Error).message || "Payment failed",
    };
  }
}
