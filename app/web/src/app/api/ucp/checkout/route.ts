export const runtime = "nodejs";

import { createHash } from "node:crypto";
import { auth } from "@/auth";
import { corsHeaders, withCors } from "@/lib/ucp/cors";
import { createCheckout } from "@/lib/ucp/handlers/checkout";
import { negotiateCapabilities, parseUCPAgent } from "@/lib/ucp/negotiation";
import { errorResponse, wrapResponse } from "@/lib/ucp/response";
import {
  CreateCheckoutInputSchema,
  validateIdempotencyKey,
} from "@/lib/ucp/schemas/checkout";
import { getIdempotency, setIdempotency } from "@/lib/ucp/store";
import type { CheckoutSession } from "@/lib/ucp/types/checkout";

const IDEMPOTENCY_TTL_MS = 24 * 60 * 60 * 1000;

function isUcpEnabled(): boolean {
  return process.env.UCP_ENABLED === "true";
}

export async function OPTIONS(req: Request) {
  return new Response(null, { status: 204, headers: corsHeaders(req) });
}

export async function POST(req: Request) {
  if (!isUcpEnabled())
    return withCors(
      errorResponse(503, "ucp_disabled", "UCP is not enabled"),
      req,
    );

  const body = await req.json().catch(() => null);
  if (!body)
    return withCors(
      errorResponse(400, "invalid_json", "Request body must be valid JSON"),
      req,
    );

  const parsed = CreateCheckoutInputSchema.safeParse(body);
  if (!parsed.success) {
    return withCors(
      errorResponse(
        400,
        "invalid_request",
        parsed.error.issues[0]?.message ?? "Invalid request",
      ),
      req,
    );
  }

  const idempotencyKey = req.headers.get("Idempotency-Key");
  const keyValidation = validateIdempotencyKey(idempotencyKey);
  if (!keyValidation.valid)
    return withCors(
      errorResponse(400, "invalid_idempotency_key", keyValidation.error ?? ""),
      req,
    );

  // Scope the store key by the full canonical body so:
  //   (a) a leaked key cannot replay a different caller's session (cart isolation)
  //   (b) a different currency with the same key gets a fresh session (correctness)
  // Note: callers must use a new key if external cart contents change — per the
  // IETF Idempotency-Key draft, the key must not be reused for different requests.
  const scopedKey = idempotencyKey
    ? createHash("sha256")
        .update(`${idempotencyKey}:${JSON.stringify(parsed.data)}`)
        .digest("hex")
    : null;

  if (scopedKey) {
    const cached = getIdempotency(scopedKey);
    if (cached) {
      const ucpAgent = parseUCPAgent(req.headers.get("UCP-Agent"));
      const negotiation = await negotiateCapabilities(ucpAgent?.profile);
      const session = structuredClone(cached.response) as CheckoutSession;
      return withCors(
        wrapResponse(
          { ...session, session_id: session.id } as Record<string, unknown>,
          negotiation,
          201,
        ),
        req,
      );
    }
  }

  let userId: string | undefined;
  try {
    const session = await auth();
    userId = session?.user?.id ?? undefined;
  } catch (err) {
    console.error("[ucp] auth() failed; treating as anonymous:", err);
  }

  const ucpAgent = parseUCPAgent(req.headers.get("UCP-Agent"));
  const negotiation = await negotiateCapabilities(ucpAgent?.profile);

  const result = await createCheckout(parsed.data, { userId });
  if ("error" in result) {
    return withCors(
      errorResponse(result.status, result.code, result.content),
      req,
    );
  }

  if (scopedKey) {
    setIdempotency(scopedKey, {
      response: structuredClone(result.session),
      hash: createHash("sha256").update(result.session.id).digest("hex"),
      expires_at: Date.now() + IDEMPOTENCY_TTL_MS,
    });
  }

  const responseBody = { ...result.session, session_id: result.session.id };
  return withCors(
    wrapResponse(responseBody as Record<string, unknown>, negotiation, 201),
    req,
  );
}
