export interface LineItem {
  productId: string;
  quantity: number;
  price: number;
}

export interface CreatePaymentRequest {
  totalAmount: number;
  currency: string;
  items: LineItem[];
}

export interface PaymentResponse {
  id: string;
  totalAmount: number;
  currency: string;
  status: string;
  createdAt: number;
}

const PAYMENT_URL = process.env.PAYMENT_URL || "http://localhost:8890";

export async function createPayment(
  req: CreatePaymentRequest,
): Promise<PaymentResponse> {
  const res = await fetch(`${PAYMENT_URL}/api/v1/payments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error("Failed to create payment");
  return res.json();
}
