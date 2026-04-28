"use client";

import { Loader2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { useCart } from "@/lib/store/cart";
import { Button } from "./ui/button";

interface AddToCartButtonProps {
  productId: string;
  name: string;
}

export function AddToCartButton({ productId, name }: AddToCartButtonProps) {
  const addItem = useCart((state) => state.addItem);
  const [loading, setLoading] = useState(false);

  const handleAdd = async () => {
    setLoading(true);
    try {
      await addItem(productId, 1);
      toast.success(`Added ${name} to your bag.`);
    } catch (_e) {
      toast.error("Failed to add to bag.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button
      className="w-full rounded-full h-14 text-base font-medium transition-all active:scale-[0.98]"
      onClick={handleAdd}
      disabled={loading}
    >
      {loading ? (
        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
      ) : (
        "Add to Bag"
      )}
    </Button>
  );
}
