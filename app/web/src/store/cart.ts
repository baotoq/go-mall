"use client";

import { create } from "zustand";
import {
  addCartItem,
  clearCart as clearCartAPI,
  getCart,
  removeCartItem,
  updateCartItem,
} from "@/lib/api";
import type { CartData } from "@/lib/types";

export interface CartItem {
  id: string;
  name: string;
  priceCents: number;
  imageUrl: string;
  quantity: number;
}

function getSessionId(): string {
  if (typeof window === "undefined") return "ssr";
  let id = localStorage.getItem("cart_session_id");
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem("cart_session_id", id);
  }
  return id;
}

interface CartStore {
  items: CartItem[];
  isLoading: boolean;
  addItem: (item: Omit<CartItem, "quantity">) => void;
  removeItem: (id: string) => void;
  updateQuantity: (id: string, quantity: number) => void;
  totalItems: () => number;
  totalPrice: () => number;
  clearCart: () => void;
  loadCart: () => Promise<void>;
  syncFromBackend: (cart: CartData) => void;
}

export const useCartStore = create<CartStore>((set, get) => ({
  items: [],
  isLoading: true,
  addItem: (item) => {
    set((state) => {
      const existing = state.items.find((i) => i.id === item.id);
      if (existing) {
        return {
          items: state.items.map((i) =>
            i.id === item.id ? { ...i, quantity: i.quantity + 1 } : i,
          ),
        };
      }
      return { items: [...state.items, { ...item, quantity: 1 }] };
    });
    const sessionId = getSessionId();
    addCartItem(sessionId, {
      productId: item.id,
      name: item.name,
      priceCents: item.priceCents,
      currency: "USD",
      imageUrl: item.imageUrl,
      quantity: 1,
    }).then((cart) => {
      if (cart) get().syncFromBackend(cart);
    });
  },
  removeItem: (id) => {
    set((state) => ({ items: state.items.filter((i) => i.id !== id) }));
    const sessionId = getSessionId();
    removeCartItem(sessionId, id).then((cart) => {
      if (cart) get().syncFromBackend(cart);
    });
  },
  updateQuantity: (id, quantity) => {
    if (quantity <= 0) {
      get().removeItem(id);
      return;
    }
    set((state) => ({
      items: state.items.map((i) => (i.id === id ? { ...i, quantity } : i)),
    }));
    const sessionId = getSessionId();
    updateCartItem(sessionId, id, quantity).then((cart) => {
      if (cart) get().syncFromBackend(cart);
    });
  },
  totalItems: () => get().items.reduce((acc, i) => acc + i.quantity, 0),
  totalPrice: () =>
    get().items.reduce((acc, i) => acc + (i.priceCents / 100) * i.quantity, 0),
  clearCart: () => {
    set({ items: [] });
    const sessionId = getSessionId();
    clearCartAPI(sessionId);
  },
  loadCart: async () => {
    const sessionId = getSessionId();
    try {
      const cart = await getCart(sessionId);
      if (cart) get().syncFromBackend(cart);
    } finally {
      set({ isLoading: false });
    }
  },
  syncFromBackend: (cart: CartData) => {
    set({
      items: cart.items.map((item) => ({
        id: item.productId,
        name: item.name,
        priceCents: item.priceCents,
        imageUrl: item.imageUrl,
        quantity: item.quantity,
      })),
    });
  },
}));
