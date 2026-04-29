import { revalidateTag } from "next/cache";
import { NextRequest, NextResponse } from "next/server";

const VALID_TAGS = new Set(["products", "categories"]);

export async function POST(req: NextRequest) {
  const secret = req.headers.get("x-revalidate-secret");
  if (secret !== process.env.REVALIDATE_SECRET) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const { tag } = await req.json();
  if (!tag || !VALID_TAGS.has(tag)) {
    return NextResponse.json({ error: "invalid tag" }, { status: 400 });
  }

  revalidateTag(tag, "max");
  return NextResponse.json({ revalidated: tag });
}
