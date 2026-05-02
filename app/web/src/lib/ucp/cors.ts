// Shared CORS helpers for UCP routes.
// Strict policy:
//   - Reflect Origin only when it appears in UCP_ALLOWED_ORIGINS allow-list.
//   - On mismatch or missing Origin, omit Access-Control-Allow-Origin entirely
//     so browsers block the request rather than disclosing a configured origin.
//   - Always include Vary: Origin so caches do not collapse responses across
//     different requesters.

const DEFAULT_ALLOWED_METHODS = "GET,POST,PATCH,OPTIONS";
const DEFAULT_ALLOWED_HEADERS =
  "Content-Type,X-UCP-Session,Idempotency-Key,UCP-Agent";
const DEFAULT_MAX_AGE_SECONDS = "600";

function getAllowedOrigins(): string[] {
  return (process.env.UCP_ALLOWED_ORIGINS ?? "http://localhost:3000")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

export function corsHeaders(req: Request): Record<string, string> {
  const origins = getAllowedOrigins();
  const origin = req.headers.get("Origin") ?? "";
  const headers: Record<string, string> = {
    Vary: "Origin",
    "Access-Control-Allow-Methods": DEFAULT_ALLOWED_METHODS,
    "Access-Control-Allow-Headers": DEFAULT_ALLOWED_HEADERS,
    "Access-Control-Max-Age": DEFAULT_MAX_AGE_SECONDS,
  };
  if (origin && origins.includes(origin)) {
    headers["Access-Control-Allow-Origin"] = origin;
  }
  return headers;
}

export function withCors(res: Response, req: Request): Response {
  const cors = corsHeaders(req);
  for (const [k, v] of Object.entries(cors)) {
    res.headers.set(k, v);
  }
  return res;
}
