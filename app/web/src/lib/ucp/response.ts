import type { NegotiationResult } from "./negotiation";

export function wrapResponse(
  data: Record<string, unknown>,
  negotiation: NegotiationResult,
  status = 200,
): Response {
  const body = {
    ...data,
    ucp: {
      version: negotiation.version,
      capabilities: negotiation.capabilities,
    },
  };
  return Response.json(body, { status });
}

// Spec-confirmed error envelope for protocol errors (400, 409, 502, 503)
export function errorResponse(
  status: number,
  code: string,
  content: string,
): Response {
  return Response.json({ code, content }, { status });
}
