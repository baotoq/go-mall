export const runtime = "nodejs";

import { auth } from "@/auth";
import { createCheckout } from "@/lib/ucp/handlers/checkout";
import { negotiateCapabilities, parseUCPAgent } from "@/lib/ucp/negotiation";
import {
  CreateCheckoutInputSchema,
  validateIdempotencyKey,
} from "@/lib/ucp/schemas/checkout";
import { errorResponse, wrapResponse } from "@/lib/ucp/response";
import { corsHeaders, withCors } from "@/lib/ucp/cors";

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
      errorResponse(400, "invalid_idempotency_key", keyValidation.error!),
      req,
    );

  let userId: string | undefined;
  try {
    const session = await auth();
    userId = session?.user?.id ?? undefined;
  } catch (err) {
    // Misconfigured auth must not silently fall through to anonymous.
    // Log the cause so an operator can spot a broken next-auth setup.
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

  const responseBody = { ...result.session, session_id: result.session.id };
  return withCors(
    wrapResponse(responseBody as Record<string, unknown>, negotiation, 201),
    req,
  );
}
