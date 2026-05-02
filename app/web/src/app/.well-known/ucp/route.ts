export const runtime = "nodejs";

import { generateProfile } from "@/lib/ucp/profile";

export function GET() {
  const profile = generateProfile();
  return Response.json(profile, {
    headers: {
      "Cache-Control": "public, max-age=3600",
      Vary: "Accept",
    },
  });
}
