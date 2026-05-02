import type { CheckoutSession, IdempotencyEntry } from "./types/checkout";

export function initStore(): void {
  if (process.env.NODE_ENV === "production") {
    throw new Error("UCP in-memory store must not be used in production");
  }
}

// Eagerly initialize on module load so dynamic import() rejects in production
initStore();

const sessions = new Map<string, CheckoutSession>();
const idempotency = new Map<string, IdempotencyEntry>();

export function getSession(id: string): CheckoutSession | null {
  const session = sessions.get(id);
  if (!session) return null;
  if (new Date(session.expires_at).getTime() < Date.now()) {
    sessions.delete(id);
    return null;
  }
  return session;
}

export function setSession(id: string, session: CheckoutSession): void {
  sessions.set(id, session);
}

export function getIdempotency(key: string): IdempotencyEntry | null {
  const entry = idempotency.get(`idempotency:${key}`);
  if (!entry) return null;
  if (entry.expires_at < Date.now()) {
    idempotency.delete(`idempotency:${key}`);
    return null;
  }
  return entry;
}

export function setIdempotency(key: string, entry: IdempotencyEntry): void {
  idempotency.set(`idempotency:${key}`, entry);
}

// Periodic eviction every 60s (only in non-production)
setInterval(() => {
  const now = Date.now();
  for (const [id, session] of sessions) {
    if (new Date(session.expires_at).getTime() < now) {
      sessions.delete(id);
    }
  }
  for (const [key, entry] of idempotency) {
    if (entry.expires_at < now) {
      idempotency.delete(key);
    }
  }
}, 60_000).unref();
