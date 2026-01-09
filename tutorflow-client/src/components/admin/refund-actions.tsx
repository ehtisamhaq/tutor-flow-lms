"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Check, X } from "lucide-react";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

interface RefundActionsProps {
  refundId: string;
}

export function RefundActions({ refundId }: RefundActionsProps) {
  const router = useRouter();
  const [loading, setLoading] = useState<"approve" | "reject" | null>(null);

  const handleAction = async (action: "approve" | "reject") => {
    setLoading(action);
    try {
      await api.post(`/admin/refunds/${refundId}/${action}`);
      toast.success(`Refund ${action}d successfully`);
      router.refresh();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error?.message || `Failed to ${action} refund`
      );
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="flex gap-2">
      <Button
        type="button"
        size="sm"
        className="bg-green-600 hover:bg-green-700"
        onClick={() => handleAction("approve")}
        disabled={loading !== null}
      >
        <Check className="h-4 w-4 mr-1" />
        {loading === "approve" ? "Approving..." : "Approve"}
      </Button>
      <Button
        type="button"
        size="sm"
        variant="destructive"
        onClick={() => handleAction("reject")}
        disabled={loading !== null}
      >
        <X className="h-4 w-4 mr-1" />
        {loading === "reject" ? "Rejecting..." : "Reject"}
      </Button>
    </div>
  );
}
