import { auth } from "@/auth";
import { redirect } from "next/navigation";
import { CartClient } from "./cart-client";

export default async function CartPage() {
  const session = await auth();
  if (!session) {
    redirect("/signin?callbackUrl=/cart");
  }
  return <CartClient />;
}
