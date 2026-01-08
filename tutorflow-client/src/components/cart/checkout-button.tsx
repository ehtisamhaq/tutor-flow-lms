"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import api from "@/lib/api";
import { useCartStore } from "@/store/cart-store";

interface CheckoutButtonProps {
  total: number;
}

export function CheckoutButton({ total }: CheckoutButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const { clearCart } = useCartStore();

  const handleCheckout = async () => {
    setIsLoading(true);
    try {
      const response = await api.post("/orders/checkout");
      const { checkout_url } = response.data.data;

      if (checkout_url) {
        // Redirect to Stripe checkout
        window.location.href = checkout_url;
      } else {
        // Free course - direct enrollment
        clearCart();
        toast.success("Enrollment successful!");
        router.push("/dashboard/my-courses");
      }
    } catch (error: unknown) {
      const err = error as {
        response?: { data?: { error?: { message?: string } } };
      };
      toast.error(err.response?.data?.error?.message || "Checkout failed");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Button
      className="w-full"
      size="lg"
      onClick={handleCheckout}
      disabled={isLoading}
    >
      {isLoading ? "Processing..." : `Checkout - $${total.toFixed(2)}`}
    </Button>
  );
}
