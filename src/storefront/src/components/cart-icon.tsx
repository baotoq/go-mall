"use client";

import { ShoppingBag } from "lucide-react";
import { useEffect, useState } from "react";
import { useCart } from "@/lib/store/cart";
import { CartSheet } from "./cart-sheet";

export function CartIcon() {
  const [mounted, setMounted] = useState(false);
  const items = useCart((state) => state.items);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const itemCount = mounted
    ? items.reduce((acc, item) => acc + item.quantity, 0)
    : 0;

  return (
    <>
      <button
        type="button"
        aria-label="Cart"
        className="relative text-ink/80 hover:text-ink transition-colors flex items-center justify-center p-1 -m-1"
        onClick={() => setOpen(true)}
      >
        <ShoppingBag className="h-4 w-4" />
        {itemCount > 0 && (
          <span className="absolute -top-1 -right-1 flex h-3 w-3 items-center justify-center rounded-full bg-accent text-[8px] font-bold text-white">
            {itemCount}
          </span>
        )}
      </button>
      <CartSheet open={open} onOpenChange={setOpen} />
    </>
  );
}
