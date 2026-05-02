import { NextRequest, NextResponse } from "next/server";
import { createKeycloakUser } from "@/lib/keycloak";

export async function POST(req: NextRequest) {
  const origin = req.headers.get("origin") ?? "";
  const host = req.headers.get("host") ?? "";
  const expectedOrigin = `${req.nextUrl.protocol}//${host}`;
  if (origin !== expectedOrigin) {
    return NextResponse.json({ error: "Forbidden" }, { status: 403 });
  }

  let email: string, password: string;
  try {
    ({ email, password } = await req.json());
  } catch {
    return NextResponse.json(
      { error: "Invalid request body" },
      { status: 400 },
    );
  }

  if (!email || !password) {
    return NextResponse.json(
      { error: "Email and password required" },
      { status: 400 },
    );
  }
  if (password.length < 8) {
    return NextResponse.json(
      { error: "Password must be at least 8 characters" },
      { status: 400 },
    );
  }

  try {
    await createKeycloakUser(email, password);
    return NextResponse.json({ success: true });
  } catch (e) {
    if (e instanceof Error && e.message === "UserAlreadyExists") {
      return NextResponse.json(
        {
          message: "Check your inbox — we'll send a link if an account exists.",
        },
        { status: 200 },
      );
    }
    console.error("Register error:", e);
    return NextResponse.json(
      { error: "Service unavailable. Try again." },
      { status: 503 },
    );
  }
}
