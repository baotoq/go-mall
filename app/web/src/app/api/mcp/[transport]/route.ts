export const runtime = "nodejs";

import { createMcpHandler } from "mcp-handler";
import { z } from "zod";
import {
  createCheckout,
  getCheckoutSession,
  updateCheckout,
  completeCheckout,
} from "@/lib/ucp/handlers/checkout";
import { generateProfile } from "@/lib/ucp/profile";
import { negotiateCapabilities } from "@/lib/ucp/negotiation";
import ucpConfig from "@/../ucp.config.json";

const handler = createMcpHandler(
  (server) => {
    server.registerTool(
      "ucp_get_profile",
      {
        title: "Get UCP Profile",
        description: "Get the UCP discovery profile and supported capabilities",
        inputSchema: {},
      },
      async () => {
        const profile = generateProfile();
        return {
          content: [{ type: "text", text: JSON.stringify(profile, null, 2) }],
        };
      },
    );

    server.registerTool(
      "ucp_create_checkout",
      {
        title: "Create Checkout",
        description:
          "Create a UCP checkout session from a cart. Returns session_id for subsequent calls.",
        inputSchema: {
          cart_session_id: z.string().describe("Cart session ID to checkout"),
          currency: z.string().length(3).describe("ISO 4217 currency code"),
          buyer_email: z
            .string()
            .email()
            .optional()
            .describe("Buyer email (optional, can PATCH later)"),
          platform_profile_url: z
            .string()
            .url()
            .optional()
            .describe("Platform UCP profile URL for capability negotiation"),
        },
      },
      async ({
        cart_session_id,
        currency,
        buyer_email,
        platform_profile_url,
      }) => {
        await negotiateCapabilities(platform_profile_url);
        const result = await createCheckout(
          {
            cart_session_id,
            currency,
            buyer: buyer_email ? { email: buyer_email } : undefined,
          },
          {},
        );
        if ("error" in result) {
          return {
            content: [
              {
                type: "text",
                text: JSON.stringify({
                  error: result.code,
                  content: result.content,
                }),
              },
            ],
            isError: true,
          };
        }
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(
                { ...result.session, session_id: result.session.id },
                null,
                2,
              ),
            },
          ],
        };
      },
    );

    server.registerTool(
      "ucp_get_checkout",
      {
        title: "Get Checkout",
        description: "Get a checkout session by ID.",
        inputSchema: {
          session_id: z.string().describe("Checkout session ID"),
        },
      },
      async ({ session_id }) => {
        const session = await getCheckoutSession(session_id);
        if (!session) {
          return {
            content: [
              {
                type: "text",
                text: JSON.stringify({
                  error: "not_found",
                  content: "Session not found",
                }),
              },
            ],
            isError: true,
          };
        }
        return {
          content: [{ type: "text", text: JSON.stringify(session, null, 2) }],
        };
      },
    );

    server.registerTool(
      "ucp_update_checkout",
      {
        title: "Update Checkout",
        description:
          "Update checkout buyer info. Setting buyer.email transitions to ready_for_complete.",
        inputSchema: {
          session_id: z.string().describe("Checkout session ID"),
          buyer_email: z.string().email().describe("Buyer email address"),
        },
      },
      async ({ session_id, buyer_email }) => {
        const session = await updateCheckout(session_id, {
          buyer: { email: buyer_email },
        });
        if (!session) {
          return {
            content: [
              {
                type: "text",
                text: JSON.stringify({
                  error: "not_found",
                  content: "Session not found",
                }),
              },
            ],
            isError: true,
          };
        }
        return {
          content: [{ type: "text", text: JSON.stringify(session, null, 2) }],
        };
      },
    );

    server.registerTool(
      "ucp_complete_checkout",
      {
        title: "Complete Checkout",
        description:
          "Complete a checkout session. Session must be in ready_for_complete status.",
        inputSchema: {
          session_id: z.string().describe("Checkout session ID"),
          idempotency_key: z
            .string()
            .optional()
            .describe("Optional idempotency key for safe retries"),
        },
      },
      async ({ session_id, idempotency_key }) => {
        const result = await completeCheckout(session_id, {
          idempotencyKey: idempotency_key ?? null,
        });
        if ("error" in result) {
          return {
            content: [
              {
                type: "text",
                text: JSON.stringify({
                  error: result.code,
                  content: result.content,
                }),
              },
            ],
            isError: true,
          };
        }
        return {
          content: [
            { type: "text", text: JSON.stringify(result.session, null, 2) },
          ],
        };
      },
    );
  },
  { serverInfo: { name: "ucp-shopping", version: ucpConfig.ucp_version } },
  {
    basePath: "/api/mcp",
    maxDuration: 60,
    verboseLogs: process.env.NODE_ENV === "development",
  },
);

export { handler as GET, handler as POST };
