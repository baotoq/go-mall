"use client"

import Link from "next/link"
import { ShoppingCart, Store } from "lucide-react"
import { useCartStore } from "@/store/cart"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function Header() {
  const totalItems = useCartStore((state) => state.totalItems())

  return (
    <header className="sticky top-0 z-50 border-b bg-background/80 backdrop-blur-sm">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        <Link
          href="/"
          className="flex items-center gap-2 font-bold text-lg"
        >
          <Store className="size-5" />
          GoMall
        </Link>

        <nav className="flex items-center gap-6 text-sm font-medium">
          <Link
            href="/"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Home
          </Link>
          <Link
            href="/products"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Products
          </Link>
        </nav>

        <Link
          href="/cart"
          className={cn(
            buttonVariants({ variant: "outline", size: "sm" }),
            "relative gap-2",
          )}
        >
          <ShoppingCart className="size-4" />
          Cart
          {totalItems > 0 && (
            <span className="absolute -top-1.5 -right-1.5 size-4 rounded-full bg-primary text-primary-foreground text-[10px] flex items-center justify-center font-bold">
              {totalItems > 9 ? "9+" : totalItems}
            </span>
          )}
        </Link>
      </div>
    </header>
  )
}
