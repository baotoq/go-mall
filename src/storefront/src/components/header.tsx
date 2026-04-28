export const dynamic = "force-dynamic";

import { Search, User } from "lucide-react";
import Link from "next/link";
import { Suspense } from "react";
import { listCategories } from "@/lib/api/catalog";
import { CartIcon } from "./cart-icon";

export async function Header() {
  let categories: { id: string; name: string }[] = [];
  try {
    const data = await listCategories();
    categories = data.categories || [];
  } catch (_e) {
    // Ignore, fallback to empty
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b border-white/10 bg-surface/80 backdrop-blur-md saturate-[180%]">
      <div className="mx-auto flex h-12 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <Link href="/" className="font-semibold tracking-tighter text-ink">
          Apple
        </Link>
        <nav className="hidden md:flex gap-6">
          <Link
            href="/shop"
            className="text-xs font-medium text-ink/80 hover:text-ink transition-colors"
          >
            Shop
          </Link>
          {categories.map((c) => (
            <Link
              key={c.id}
              href={`/shop?categoryId=${c.id}`}
              className="text-xs font-medium text-ink/80 hover:text-ink transition-colors"
            >
              {c.name}
            </Link>
          ))}
        </nav>
        <div className="flex items-center gap-4">
          <button
            type="button"
            aria-label="Search"
            className="text-ink/80 hover:text-ink transition-colors"
          >
            <Search className="h-4 w-4" />
          </button>
          <Link
            href="/account"
            aria-label="Account"
            className="text-ink/80 hover:text-ink transition-colors"
          >
            <User className="h-4 w-4" />
          </Link>
          <Suspense fallback={<div className="h-4 w-4" />}>
            <CartIcon />
          </Suspense>
        </div>
      </div>
    </header>
  );
}
