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

interface CheckoutResponse {
  checkout_url?: string;
  order?: any;
}

export function CheckoutButton({ total }: CheckoutButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const { clearCart } = useCartStore();

  const handleCheckout = async () => {
    setIsLoading(true);
    console.log("Initiating checkout...");
    try {
      const response = await api.post<CheckoutResponse>("/orders/checkout", {});
      console.log("Checkout response:", response);

      // response.data is the T (CheckoutResponse) because api.ts unwraps the backend's "data" field
      const checkoutData = response.data;
      const checkoutUrl = checkoutData.checkout_url;

      console.log("Extracted checkout URL:", checkoutUrl);

      if (checkoutUrl) {
        console.log("Redirecting to Stripe:", checkoutUrl);
        window.location.href = checkoutUrl;
      } else {
        console.log("No checkout URL found, assuming free enrollment");
        // Free course - direct enrollment
        clearCart();
        toast.success("Enrollment successful!");
        router.push("/dashboard/my-courses");
      }
    } catch (error: unknown) {
      console.error("Checkout error:", error);
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
