import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch } from "@/lib/server-api";
import {
  DollarSign,
  TrendingUp,
  Download,
  ArrowUpRight,
  Calendar,
  CreditCard,
} from "lucide-react";
import Link from "next/link";

interface EarningsStats {
  total_earnings: number;
  available_balance: number;
  pending_payout: number;
  this_month: number;
  last_month: number;
  change_percent: number;
  recent_transactions: {
    id: string;
    type: "sale" | "payout" | "refund";
    amount: number;
    course_title?: string;
    created_at: string;
    status: "completed" | "pending" | "failed";
  }[];
  monthly_earnings: {
    month: string;
    amount: number;
  }[];
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

// Server Component
export default async function TutorEarningsPage() {
  const stats = await authServerFetch<EarningsStats>("/tutors/earnings");

  const changePercent = stats?.change_percent || 0;
  const isPositive = changePercent >= 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Earnings</h2>
          <p className="text-muted-foreground">
            Track your revenue and payouts
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline">
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
          <Button>
            <CreditCard className="mr-2 h-4 w-4" />
            Request Payout
          </Button>
        </div>
      </div>

      {/* Earnings Stats */}
      <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Earnings
            </CardTitle>
            <DollarSign className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(stats?.total_earnings || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">Lifetime</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Available Balance
            </CardTitle>
            <CreditCard className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              ${(stats?.available_balance || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">Ready for payout</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              This Month
            </CardTitle>
            <TrendingUp className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(stats?.this_month || 0).toLocaleString()}
            </div>
            <div className="flex items-center text-xs mt-1">
              {isPositive ? (
                <ArrowUpRight className="h-4 w-4 text-green-600" />
              ) : (
                <ArrowUpRight className="h-4 w-4 text-red-600 rotate-180" />
              )}
              <span className={isPositive ? "text-green-600" : "text-red-600"}>
                {Math.abs(changePercent)}%
              </span>
              <span className="text-muted-foreground ml-1">vs last month</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Pending Payout
            </CardTitle>
            <Calendar className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(stats?.pending_payout || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">Processing</p>
          </CardContent>
        </Card>
      </div>

      {/* Earnings Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Monthly Earnings</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center bg-muted/30 rounded-lg">
            {stats?.monthly_earnings?.length ? (
              <div className="w-full h-full flex items-end justify-around p-4 gap-2">
                {stats.monthly_earnings.map((data, index) => {
                  const maxAmount = Math.max(
                    ...stats.monthly_earnings.map((d) => d.amount)
                  );
                  const height =
                    maxAmount > 0 ? (data.amount / maxAmount) * 100 : 0;

                  return (
                    <div
                      key={index}
                      className="flex flex-col items-center gap-1"
                    >
                      <div
                        className="w-8 bg-green-600 rounded-t transition-all"
                        style={{ height: `${Math.max(height, 5)}%` }}
                        title={`$${data.amount.toLocaleString()}`}
                      />
                      <span className="text-xs text-muted-foreground">
                        {data.month.slice(0, 3)}
                      </span>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="text-center text-muted-foreground">
                <TrendingUp className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No earnings data yet</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Recent Transactions */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Transactions</CardTitle>
        </CardHeader>
        <CardContent>
          {stats?.recent_transactions?.length ? (
            <div className="space-y-3">
              {stats.recent_transactions.map((tx) => (
                <div
                  key={tx.id}
                  className="flex items-center justify-between p-3 bg-muted/50 rounded-lg"
                >
                  <div className="flex items-center gap-4">
                    <div
                      className={`h-10 w-10 rounded-full flex items-center justify-center ${
                        tx.type === "sale"
                          ? "bg-green-100 text-green-600"
                          : tx.type === "payout"
                          ? "bg-blue-100 text-blue-600"
                          : "bg-red-100 text-red-600"
                      }`}
                    >
                      <DollarSign className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="font-medium capitalize">{tx.type}</p>
                      {tx.course_title && (
                        <p className="text-sm text-muted-foreground truncate max-w-[200px]">
                          {tx.course_title}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="text-right">
                    <div
                      className={`font-semibold ${
                        tx.type === "refund" ? "text-red-600" : ""
                      }`}
                    >
                      {tx.type === "refund" ? "-" : "+"}$
                      {Math.abs(tx.amount).toFixed(2)}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {formatDate(tx.created_at)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <DollarSign className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No transactions yet</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
