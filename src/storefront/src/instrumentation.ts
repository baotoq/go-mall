import { registerOTel } from "@vercel/otel";

export function register() {
  registerOTel({ serviceName: "storefront" });

  if (process.env.NODE_ENV === "development") {
    const _fetch = globalThis.fetch;
    globalThis.fetch = (url, opts) =>
      _fetch(url, { cache: "no-store", ...opts });
  }
}
