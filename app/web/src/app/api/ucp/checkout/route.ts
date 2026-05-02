export const runtime = "nodejs";

import { auth } from "@/auth";
import { createCheckout } from "@/lib/ucp/handlers/checkout";
import { negotiateCapabilities, parseUCPAgent } from "@/lib/ucp/negotiation";
import {
  CreateCheckoutInputSchema,
  validateIdempotencyKey,
} from "@/lib/ucp/schemas/checkout";
import { errorResponse, wrapResponse } from "@/lib/ucp/response";

function isUcpEnabled(): boolean {
  return process.env.UCP_ENABLED === "true";
}

function corsHeaders(req: Request): Record<string, string> {
  const origins = (process.env.UCP_ALLOWED_ORIGINS ?? "http://localhost:3000")
    .split(",")
    .map((s) => s.trim());
  const origin = req.headers.get("Origin") ?? "";
  const allowed = origins.includes(origin) ? origin : origins[0];
  return {
    "Access-Control-Allow-Origin": allowed,
    "Access-Control-Allow-Methods": "GET,POST,PATCH,OPTIONS",
    "Access-Control-Allow-Headers":
      "Content-Type,X-UCP-Session,Idempotency-Key,UCP-Agent",
  };
}

export async function OPTIONS(req: Request) {
  return new Response(null, { status: 204, headers: corsHeaders(req) });
}

export async function POST(req: Request) {
  if (!isUcpEnabled())
    return errorResponse(503, "ucp_disabled", "UCP is not enabled");

  const body = await req.json().catch(() => null);
  if (!body)
    return errorResponse(
      400,
      "invalid_json",
      "Request body must be valid JSON",
    );

  const parsed = CreateCheckoutInputSchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse(
      400,
      "invalid_request",
      parsed.error.issues[0]?.message ?? "Invalid request",
    );
  }

  const idempotencyKey = req.headers.get("Idempotency-Key");
  const keyValidation = validateIdempotencyKey(idempotencyKey);
  if (!keyValidation.valid)
    return errorResponse(400, "invalid_idempotency_key", keyValidation.error!);

  let userId: string | undefined;
  try {
    const session = await auth();
    userId = session?.user?.id ?? undefined;
  } catch {
    /* anonymous */
  }

  const ucpAgent = parseUCPAgent(req.headers.get("UCP-Agent"));
  const negotiation = await negotiateCapabilities(ucpAgent?.profile);

  const result = await createCheckout(parsed.data, { userId });
  if ("error" in result) {
    return errorResponse(result.status, result.code, result.content);
  }

  const responseBody = { ...result.session, session_id: result.session.id };
  return wrapResponse(
    responseBody as Record<string, unknown>,
    negotiation,
    201,
  );
}
