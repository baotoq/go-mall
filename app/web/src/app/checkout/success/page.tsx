import { getCheckoutSession } from "@/lib/ucp/handlers/checkout";
import { SuccessClient } from "./success-client";

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ id?: string }>;
}) {
  const { id } = await searchParams;
  if (!id) return <SuccessClient session={null} />;
  const session = await getCheckoutSession(id);
  return <SuccessClient session={session} />;
}
