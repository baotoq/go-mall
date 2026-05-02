export const runtime = "nodejs";

import { corsHeaders, withCors } from "@/lib/ucp/cors";
import {
  completeCheckout,
  getCheckoutSession,
  updateCheckout,
} from "@/lib/ucp/handlers/checkout";
import { negotiateCapabilities, parseUCPAgent } from "@/lib/ucp/negotiation";
import { errorResponse, wrapResponse } from "@/lib/ucp/response";
import { UpdateCheckoutInputSchema } from "@/lib/ucp/schemas/checkout";

function isUcpEnabled(): boolean {
  return process.env.UCP_ENABLED === "true";
}

function getSessionFromHeader(req: Request): string | null {
  return req.headers.get("X-UCP-Session");
}

// Bearer-style check: token equals session id (UUIDv4, ~122 bits entropy).
// Disclosure of the id is full takeover — callers must keep it out of logs/URLs.
function authorize(
  req: Request,
  id: string,
): { ok: true } | { ok: false; res: Response } {
  const sessionToken = getSessionFromHeader(req);
  if (!sessionToken) {
    return {
      ok: false,
      res: withCors(
        errorResponse(
          401,
          "missing_session",
          "X-UCP-Session header is required",
        ),
        req,
      ),
    };
  }
  if (sessionToken !== id) {
    return {
      ok: false,
      res: withCors(
        errorResponse(403, "session_mismatch", "X-UCP-Session does not match"),
        req,
      ),
    };
  }
  return { ok: true };
}

export async function OPTIONS(req: Request) {
  return new Response(null, { status: 204, headers: corsHeaders(req) });
}

export async function GET(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  if (!isUcpEnabled())
    return withCors(
      errorResponse(503, "ucp_disabled", "UCP is not enabled"),
      req,
    );

  const { id } = await params;
  const auth = authorize(req, id);
  if (!auth.ok) return auth.res;

  const session = await getCheckoutSession(id);
  if (!session)
    return withCors(
      errorResponse(404, "not_found", "Checkout session not found"),
      req,
    );
  const negotiation = await negotiateCapabilities();
  return withCors(
    wrapResponse(session as unknown as Record<string, unknown>, negotiation),
    req,
  );
}

export async function PATCH(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  if (!isUcpEnabled())
    return withCors(
      errorResponse(503, "ucp_disabled", "UCP is not enabled"),
      req,
    );

  const { id } = await params;
  const auth = authorize(req, id);
  if (!auth.ok) return auth.res;

  const body = await req.json().catch(() => null);
  const parsed = UpdateCheckoutInputSchema.safeParse(body ?? {});
  if (!parsed.success)
    return withCors(
      errorResponse(400, "invalid_request", parsed.error.issues[0].message),
      req,
    );

  const ucpAgent = parseUCPAgent(req.headers.get("UCP-Agent"));
  const negotiation = await negotiateCapabilities(ucpAgent?.profile);

  const session = await updateCheckout(id, parsed.data);
  if (!session)
    return withCors(
      errorResponse(404, "not_found", "Checkout session not found"),
      req,
    );
  return withCors(
    wrapResponse(session as unknown as Record<string, unknown>, negotiation),
    req,
  );
}

export async function POST(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  if (!isUcpEnabled())
    return withCors(
      errorResponse(503, "ucp_disabled", "UCP is not enabled"),
      req,
    );

  const { id } = await params;
  const auth = authorize(req, id);
  if (!auth.ok) return auth.res;

  const url = new URL(req.url);
  const action = url.searchParams.get("action");
  if (action !== "complete")
    return withCors(
      errorResponse(400, "invalid_action", "Only action=complete is supported"),
      req,
    );

  const idempotencyKey = req.headers.get("Idempotency-Key");
  const ucpAgent = parseUCPAgent(req.headers.get("UCP-Agent"));
  const negotiation = await negotiateCapabilities(ucpAgent?.profile);

  const result = await completeCheckout(id, { idempotencyKey });
  if ("error" in result) {
    return withCors(
      errorResponse(result.status, result.code, result.content),
      req,
    );
  }
  return withCors(
    wrapResponse(
      result.session as unknown as Record<string, unknown>,
      negotiation,
    ),
    req,
  );
}
