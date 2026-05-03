import { Separator } from "@/components/ui/separator";
import type { CartItem } from "@/store/cart";

const SHIPPING_CENTS = 599;

interface OrderSummaryProps {
  items: CartItem[];
}

export function OrderSummary({ items }: OrderSummaryProps) {
  const subtotalCents = items.reduce(
    (acc, item) => acc + item.priceCents * item.quantity,
    0,
  );
  const totalCents = subtotalCents + SHIPPING_CENTS;

  return (
    <div className="rounded-xl border bg-card p-6 space-y-4 sticky top-4">
      <h2 className="font-semibold text-base">Order Summary</h2>
      <div className="space-y-2">
        {items.map((item) => (
          <div key={item.id} className="flex justify-between text-sm">
            <span className="text-muted-foreground line-clamp-1 flex-1 min-w-0 mr-2">
              {item.name}
              {item.quantity > 1 && (
                <span className="ml-1">× {item.quantity}</span>
              )}
            </span>
            <span className="shrink-0 font-medium">
              ${((item.priceCents / 100) * item.quantity).toFixed(2)}
            </span>
          </div>
        ))}
      </div>
      <Separator />
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Subtotal</span>
          <span>${(subtotalCents / 100).toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Shipping</span>
          <span>${(SHIPPING_CENTS / 100).toFixed(2)}</span>
        </div>
      </div>
      <Separator />
      <div className="flex justify-between font-bold text-base">
        <span>Total</span>
        <span>${(totalCents / 100).toFixed(2)}</span>
      </div>
    </div>
  );
}
