import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { RefundActions } from "@/components/admin/refund-actions";
import { authServerFetch } from "@/lib/server-api";
import {
  RotateCcw,
  Check,
  X,
  Clock,
  DollarSign,
  AlertCircle,
} from "lucide-react";

interface Refund {
  id: string;
  order_id: string;
  user_id: string;
  amount: number;
  reason: string;
  description: string;
  status: "pending" | "approved" | "rejected" | "processed";
  admin_notes: string;
  created_at: string;
  processed_at: string | null;
  user?: {
    name: string;
    email: string;
  };
  order?: {
    order_number: string;
  };
}

interface RefundsResponse {
  items: Refund[];
  total: number;
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function getStatusColor(status: string) {
  switch (status) {
    case "pending":
      return "bg-yellow-100 text-yellow-700";
    case "approved":
    case "processed":
      return "bg-green-100 text-green-700";
    case "rejected":
      return "bg-red-100 text-red-700";
    default:
      return "bg-gray-100 text-gray-700";
  }
}

function getReasonLabel(reason: string) {
  const labels: Record<string, string> = {
    not_as_described: "Not as described",
    duplicate: "Duplicate purchase",
    technical_issue: "Technical issue",
    no_longer_needed: "No longer needed",
    other: "Other",
  };
  return labels[reason] || reason;
}

export default async function RefundsPage() {
  const refunds = await authServerFetch<RefundsResponse>("/admin/refunds");
  const pendingRefunds =
    refunds?.items?.filter((r) => r.status === "pending") || [];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Refund Management</h2>
        <p className="text-muted-foreground">
          Review and process refund requests
        </p>
      </div>

      {/* Stats */}
      <div className="grid sm:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Requests
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{refunds?.total || 0}</div>
          </CardContent>
        </Card>
        <Card className={pendingRefunds.length > 0 ? "border-yellow-300" : ""}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Clock className="h-4 w-4" />
              Pending
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">
              {pendingRefunds.length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Approved
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {refunds?.items?.filter(
                (r) => r.status === "approved" || r.status === "processed"
              ).length || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Refunded
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              $
              {refunds?.items
                ?.filter(
                  (r) => r.status === "approved" || r.status === "processed"
                )
                .reduce((sum, r) => sum + r.amount, 0)
                .toFixed(2) || "0.00"}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Pending Refunds Alert */}
      {pendingRefunds.length > 0 && (
        <Card className="border-yellow-300 bg-yellow-50 dark:bg-yellow-900/20">
          <CardContent className="py-4">
            <div className="flex items-center gap-3">
              <AlertCircle className="h-5 w-5 text-yellow-600" />
              <p className="font-medium text-yellow-800 dark:text-yellow-200">
                {pendingRefunds.length} refund request
                {pendingRefunds.length > 1 ? "s" : ""} awaiting review
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Refunds Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Refund Requests</CardTitle>
        </CardHeader>
        <CardContent>
          {refunds?.items && refunds.items.length > 0 ? (
            <div className="space-y-4">
              {refunds.items.map((refund) => (
                <div
                  key={refund.id}
                  className="flex items-start justify-between p-4 border rounded-lg"
                >
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">
                        {refund.user?.name || "Unknown User"}
                      </span>
                      <span
                        className={`px-2 py-0.5 rounded-full text-xs font-medium ${getStatusColor(
                          refund.status
                        )}`}
                      >
                        {refund.status}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {refund.user?.email}
                    </p>
                    <p className="text-sm">
                      <span className="font-medium">Order:</span>{" "}
                      {refund.order?.order_number || refund.order_id}
                    </p>
                    <p className="text-sm">
                      <span className="font-medium">Reason:</span>{" "}
                      {getReasonLabel(refund.reason)}
                    </p>
                    {refund.description && (
                      <p className="text-sm text-muted-foreground">
                        {refund.description}
                      </p>
                    )}
                    <p className="text-xs text-muted-foreground">
                      Requested: {formatDate(refund.created_at)}
                      {refund.processed_at &&
                        ` • Processed: ${formatDate(refund.processed_at)}`}
                    </p>
                  </div>

                  <div className="text-right space-y-2">
                    <div className="text-xl font-bold">
                      ${refund.amount.toFixed(2)}
                    </div>
                    {refund.status === "pending" && (
                      <RefundActions refundId={refund.id} />
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-12">
              <RotateCcw className="h-12 w-12 mx-auto mb-4 text-muted-foreground opacity-50" />
              <p className="text-muted-foreground">No refund requests yet</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
