import Image from "next/image";
import Link from "next/link";
import type { ProductInfo } from "@/lib/api/catalog";

interface ProductCardProps {
  product: ProductInfo;
}

export function ProductCard({ product }: ProductCardProps) {
  return (
    <Link href={`/shop/${product.id}`} className="group flex flex-col gap-4">
      <div className="relative aspect-[4/5] w-full overflow-hidden rounded-2xl bg-surface-2 transition-transform duration-300 group-hover:scale-[1.02]">
        <Image
          src={product.imageUrl}
          alt={product.name}
          fill
          className="object-cover transition-opacity duration-300"
          sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 25vw"
        />
      </div>
      <div className="flex flex-col gap-1">
        <h3 className="text-sm font-medium text-ink">{product.name}</h3>
        <p className="text-sm text-ink-2">${product.price.toFixed(2)}</p>
      </div>
    </Link>
  );
}
