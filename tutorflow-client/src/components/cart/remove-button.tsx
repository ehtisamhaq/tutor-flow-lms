"use client";

import { useState } from "react";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import api from "@/lib/api";
import { useCartStore } from "@/store/cart-store";
import { useRouter } from "next/navigation";

interface RemoveFromCartButtonProps {
  itemId: string;
  courseId: string;
}

export function RemoveFromCartButton({
  itemId,
  courseId,
}: RemoveFromCartButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const { removeItem } = useCartStore();
  const router = useRouter();

  const handleRemove = async () => {
    setIsLoading(true);
    try {
      await api.delete(`/cart/items/${itemId}`);
      removeItem(courseId);
      toast.success("Removed from cart");
      router.refresh();
    } catch {
      toast.error("Failed to remove item");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Button
      variant="ghost"
      size="sm"
      className="text-destructive hover:text-destructive"
      onClick={handleRemove}
      disabled={isLoading}
    >
      <Trash2 className="h-4 w-4" />
    </Button>
  );
}
