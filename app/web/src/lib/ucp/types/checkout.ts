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
}

export interface CheckoutSession {
  id: string;
  status: CheckoutStatus;
  currency: string;
  cart_session_id: string;
  user_id: string;
  buyer?: BuyerInfo;
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
