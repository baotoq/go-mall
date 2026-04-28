"use client";

import { useEffect } from "react";
import { useCart } from "@/lib/store/cart";

export function CartProvider({ children }: { children: React.ReactNode }) {
  const init = useCart((state) => state.init);

  useEffect(() => {
    init();
  }, [init]);

  return <>{children}</>;
}
