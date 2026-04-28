import { v4 as uuidv4 } from "uuid";
import { create } from "zustand";
import { persist } from "zustand/middleware";
import {
  addToCart,
  getCart,
  removeCartItem,
  updateCartItem,
} from "../api/cart";

export interface CartItem {
  productId: string;
  quantity: number;
}

interface CartState {
  sessionId: string;
  items: CartItem[];
  initialized: boolean;
  init: () => Promise<void>;
  addItem: (productId: string, quantity?: number) => Promise<void>;
  updateQuantity: (productId: string, quantity: number) => Promise<void>;
  removeItem: (productId: string) => Promise<void>;
  clear: () => void;
}

export const useCart = create<CartState>()(
  persist(
    (set, get) => ({
      sessionId: "",
      items: [],
      initialized: false,
      init: async () => {
        let sid = get().sessionId;
        if (!sid) {
          sid = uuidv4();
          set({ sessionId: sid });
        }
        try {
          const cart = await getCart(sid);
          set({
            items: cart.items.map((i) => ({
              productId: i.productId,
              quantity: i.quantity,
            })),
            initialized: true,
          });
        } catch (e) {
          console.error("Failed to init cart", e);
          set({ initialized: true });
        }
      },
      addItem: async (productId, quantity = 1) => {
        const sid = get().sessionId;
        if (!sid) return;

        // Optimistic update
        set((state) => {
          const existing = state.items.find(
            (item) => item.productId === productId,
          );
          if (existing) {
            return {
              items: state.items.map((item) =>
                item.productId === productId
                  ? { ...item, quantity: item.quantity + quantity }
                  : item,
              ),
            };
          }
          return { items: [...state.items, { productId, quantity }] };
        });

        try {
          await addToCart(sid, productId, quantity);
        } catch (e) {
          console.error("Failed to add to cart on backend", e);
          // Optional: revert optimistic update on fail
        }
      },
      updateQuantity: async (productId, quantity) => {
        const sid = get().sessionId;
        if (!sid) return;

        set((state) => ({
          items: state.items.map((item) =>
            item.productId === productId ? { ...item, quantity } : item,
          ),
        }));

        try {
          await updateCartItem(sid, productId, quantity);
        } catch (e) {
          console.error("Failed to update cart on backend", e);
        }
      },
      removeItem: async (productId) => {
        const sid = get().sessionId;
        if (!sid) return;

        set((state) => ({
          items: state.items.filter((item) => item.productId !== productId),
        }));

        try {
          await removeCartItem(sid, productId);
        } catch (e) {
          console.error("Failed to remove cart item on backend", e);
        }
      },
      clear: () => set({ items: [] }),
    }),
    {
      name: "shopping-cart",
      partialize: (state) => ({
        sessionId: state.sessionId,
        items: state.items,
      }), // persist sessionId and items
    },
  ),
);
