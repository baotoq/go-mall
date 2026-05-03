export type CheckoutStatus =
  | "incomplete"
  | "requires_escalation"
  | "ready_for_complete"
  | "complete_in_progress"
  | "completed"
  | "canceled";

export interface CheckoutMessage {
  type: "error" | "warning" | "info";
  code: string;
  path?: string;
  content: string;
  severity: "recoverable" | "unrecoverable";
}

export interface CheckoutTotals {
  subtotal_cents: number;
  currency: string;
}

export interface BuyerInfo {
  email?: string;
  name?: string;
  phone?: string;
}

export interface ShippingAddress {
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postal_code: string;
  country: string; // ISO 3166-1 alpha-2
}

export interface PaymentSummary {
  brand: string; // 'visa' | 'mastercard' | 'amex' | 'unknown'
  last4: string; // 4 digits
}

export interface CheckoutSession {
  id: string;
  status: CheckoutStatus;
  currency: string;
  cart_session_id: string;
  user_id: string;
  buyer?: BuyerInfo;
  shipping_address?: ShippingAddress;
  payment?: PaymentSummary;
  totals: CheckoutTotals;
  messages: CheckoutMessage[];
  expires_at: string; // ISO 8601
  created_at: string;
  updated_at: string;
}

export interface IdempotencyEntry {
  response: unknown;
  hash: string;
  expires_at: number; // ms timestamp
}
