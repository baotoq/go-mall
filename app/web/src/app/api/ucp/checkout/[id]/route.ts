export const runtime = "nodejs";

import {
  getCheckoutSession,
  updateCheckout,
  completeCheckout,
} from "@/lib/ucp/handlers/checkout";
import { UpdateCheckoutInputSchema } from "@/lib/ucp/schemas/checkout";
import { negotiateCapabilities, parseUCPAgent } from "@/lib/ucp/negotiation";
import { errorResponse, wrapResponse } from "@/lib/ucp/response";

function isUcpEnabled(): boolean {
  return process.env.UCP_ENABLED === "true";
}

function getSessionFromHeader(req: Request): string | null {
  return req.headers.get("X-UCP-Session");
}

export async function GET(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  if (!isUcpEnabled())
    return errorResponse(503, "ucp_disabled", "UCP is not enabled");
  const { id } = await params;
  const session = await getCheckoutSession(id);
  if (!session)
    return errorResponse(404, "not_found", "Checkout session not found");
  const negotiation = await negotiateCapabilities();
  return wrapResponse(
    session as unknown as Record<string, unknown>,
    negotiation,
  );
}

export async function PATCH(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  if (!isUcpEnabled())
    return errorResponse(503, "ucp_disabled", "UCP is not enabled");
  const sessionToken = getSessionFromHeader(req);
  if (!sessionToken)
    return errorResponse(
      401,
      "missing_session",
      "X-UCP-Session header is required",
    );

  const { id } = await params;
  if (sessionToken !== id)
    return errorResponse(
      403,
      "session_mismatch",
      "X-UCP-Session does not match",
    );

  const body = await req.json().catch(() => null);
  const parsed = UpdateCheckoutInputSchema.safeParse(body ?? {});
  if (!parsed.success)
    return errorResponse(
      400,
      "invalid_request",
      parsed.error.issues[0].message,
    );

  const ucpAgent = parseUCPAgent(req.headers.get("UCP-Agent"));
  const negotiation = await negotiateCapabilities(ucpAgent?.profile);

  const session = await updateCheckout(id, parsed.data);
  if (!session)
    return errorResponse(404, "not_found", "Checkout session not found");
  return wrapResponse(
    session as unknown as Record<string, unknown>,
    negotiation,
  );
}

export async function POST(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  if (!isUcpEnabled())
    return errorResponse(503, "ucp_disabled", "UCP is not enabled");
  const sessionToken = getSessionFromHeader(req);
  if (!sessionToken)
    return errorResponse(
      401,
      "missing_session",
      "X-UCP-Session header is required",
    );

  const { id } = await params;
  if (sessionToken !== id)
    return errorResponse(
      403,
      "session_mismatch",
      "X-UCP-Session does not match",
    );

  const url = new URL(req.url);
  const action = url.searchParams.get("action");
  if (action !== "complete")
    return errorResponse(
      400,
      "invalid_action",
      "Only action=complete is supported",
    );

  const idempotencyKey = req.headers.get("Idempotency-Key");
  const ucpAgent = parseUCPAgent(req.headers.get("UCP-Agent"));
  const negotiation = await negotiateCapabilities(ucpAgent?.profile);

  const result = await completeCheckout(id, { idempotencyKey });
  if ("error" in result) {
    return errorResponse(result.status, result.code, result.content);
  }
  return wrapResponse(
    result.session as unknown as Record<string, unknown>,
    negotiation,
  );
}
