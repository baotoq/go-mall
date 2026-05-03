import type { CheckoutSession, IdempotencyEntry } from "./types/checkout";

// Production guard runs lazily on first store access, not at module load:
// `next build` evaluates route modules during page-data collection under
// NODE_ENV=production, so an eager throw would break the build. This still
// fails fast on the first real request in a prod runtime.
let prodGuardChecked = false;
export function initStore(): void {
  if (prodGuardChecked) return;
  prodGuardChecked = true;
  if (process.env.NODE_ENV === "production") {
    throw new Error("UCP in-memory store must not be used in production");
  }
}

// Survive Next.js dev HMR: route module reloads must not duplicate the
// session/idempotency Maps or stack a new setInterval on every reload.
type StoreSingleton = {
  sessions: Map<string, CheckoutSession>;
  idempotency: Map<string, IdempotencyEntry>;
  intervalId?: NodeJS.Timeout;
};

declare global {
  // eslint-disable-next-line no-var
  var __UCP_STORE__: StoreSingleton | undefined;
}

if (!globalThis.__UCP_STORE__) {
  globalThis.__UCP_STORE__ = { sessions: new Map(), idempotency: new Map() };
}
const store: StoreSingleton = globalThis.__UCP_STORE__;

const sessions = store.sessions;
const idempotency = store.idempotency;

export function getSession(id: string): CheckoutSession | null {
  initStore();
  const session = sessions.get(id);
  if (!session) return null;
  if (new Date(session.expires_at).getTime() < Date.now()) {
    sessions.delete(id);
    return null;
  }
  return session;
}

export function setSession(id: string, session: CheckoutSession): void {
  initStore();
  sessions.set(id, session);
}

export function getIdempotency(key: string): IdempotencyEntry | null {
  initStore();
  const entry = idempotency.get(key);
  if (!entry) return null;
  if (entry.expires_at < Date.now()) {
    idempotency.delete(key);
    return null;
  }
  return entry;
}

export function setIdempotency(key: string, entry: IdempotencyEntry): void {
  initStore();
  idempotency.set(key, entry);
}

// Periodic eviction every 60s, only registered once per process even
// across HMR reloads.
if (!store.intervalId) {
  const id = setInterval(() => {
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
  }, 60_000);
  id.unref();
  store.intervalId = id;
}
