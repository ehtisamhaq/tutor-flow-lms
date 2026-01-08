"use client";

import { useState } from "react";
import { ShoppingCart, Check } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import api from "@/lib/api";
import { useCartStore } from "@/store/cart-store";

interface AddToCartButtonProps {
  courseId: string;
}

export function AddToCartButton({ courseId }: AddToCartButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [isAdded, setIsAdded] = useState(false);
  const { addItem } = useCartStore();

  const handleAddToCart = async () => {
    setIsLoading(true);
    try {
      const response = await api.post("/cart/items", { course_id: courseId });
      const item = response.data.data;
      addItem(item);
      setIsAdded(true);
      toast.success("Added to cart!");
    } catch (error: unknown) {
      const err = error as {
        response?: { data?: { error?: { message?: string } } };
      };
      toast.error(
        err.response?.data?.error?.message || "Failed to add to cart"
      );
    } finally {
      setIsLoading(false);
    }
  };

  if (isAdded) {
    return (
      <Button className="w-full" size="lg" variant="secondary" disabled>
        <Check className="mr-2 h-5 w-5" />
        Added to Cart
      </Button>
    );
  }

  return (
    <Button
      className="w-full"
      size="lg"
      onClick={handleAddToCart}
      disabled={isLoading}
    >
      <ShoppingCart className="mr-2 h-5 w-5" />
      {isLoading ? "Adding..." : "Add to Cart"}
    </Button>
  );
}
