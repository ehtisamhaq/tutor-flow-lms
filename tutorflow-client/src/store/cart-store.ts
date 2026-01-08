import { create } from "zustand";

interface CartItem {
  id: string;
  course_id: string;
  course: {
    id: string;
    title: string;
    slug: string;
    thumbnail_url?: string;
    price: number;
    discount_price?: number;
    instructor: {
      id: string;
      first_name: string;
      last_name: string;
    };
  };
}

interface CartState {
  items: CartItem[];
  itemCount: number;
  total: number;
  isLoading: boolean;
  setItems: (items: CartItem[]) => void;
  addItem: (item: CartItem) => void;
  removeItem: (courseId: string) => void;
  clearCart: () => void;
}

export const useCartStore = create<CartState>((set, get) => ({
  items: [],
  itemCount: 0,
  total: 0,
  isLoading: true,

  setItems: (items) => {
    const total = items.reduce((sum, item) => {
      const price = item.course.discount_price ?? item.course.price;
      return sum + price;
    }, 0);

    set({
      items,
      itemCount: items.length,
      total,
      isLoading: false,
    });
  },

  addItem: (item) => {
    const { items } = get();
    const exists = items.some((i) => i.course_id === item.course_id);
    if (!exists) {
      const newItems = [...items, item];
      const total = newItems.reduce((sum, i) => {
        const price = i.course.discount_price ?? i.course.price;
        return sum + price;
      }, 0);

      set({
        items: newItems,
        itemCount: newItems.length,
        total,
      });
    }
  },

  removeItem: (courseId) => {
    const { items } = get();
    const newItems = items.filter((i) => i.course_id !== courseId);
    const total = newItems.reduce((sum, i) => {
      const price = i.course.discount_price ?? i.course.price;
      return sum + price;
    }, 0);

    set({
      items: newItems,
      itemCount: newItems.length,
      total,
    });
  },

  clearCart: () =>
    set({
      items: [],
      itemCount: 0,
      total: 0,
    }),
}));
