"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import api from "@/lib/api";
import { useCartStore } from "@/store/cart-store";

interface BuyNowButtonProps {
  courseId: string;
}

export function BuyNowButton({ courseId }: BuyNowButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const { addItem } = useCartStore();

  const handleBuyNow = async () => {
    setIsLoading(true);
    try {
      // 1. Add to cart
      const response = await api.post("/cart/items", { course_id: courseId });
      const item = response.data.data;
      addItem(item);

      // 2. Redirect to cart page (which has checkout button)
      router.push("/cart");
    } catch (error: unknown) {
      const err = error as {
        response?: { data?: { error?: { message?: string } } };
      };
      toast.error(
        err.response?.data?.error?.message || "Failed to initiate purchase"
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Button
      variant="outline"
      className="w-full"
      size="lg"
      onClick={handleBuyNow}
      disabled={isLoading}
    >
      {isLoading ? "Processing..." : "Buy Now"}
    </Button>
  );
}
