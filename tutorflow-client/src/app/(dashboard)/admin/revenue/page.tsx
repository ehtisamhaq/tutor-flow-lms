import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch } from "@/lib/server-api";
import {
  DollarSign,
  TrendingUp,
  Users,
  BookOpen,
  Download,
  ArrowUpRight,
  ArrowDownRight,
} from "lucide-react";
import Link from "next/link";

interface RevenueStats {
  total_revenue: number;
  this_month: number;
  last_month: number;
  change_percent: number;
  total_sales: number;
  avg_order_value: number;
  top_courses: {
    id: string;
    title: string;
    revenue: number;
    sales: number;
  }[];
  monthly_data: {
    month: string;
    revenue: number;
    sales: number;
  }[];
}

// Server Component
export default async function AdminRevenuePage() {
  const stats = await authServerFetch<RevenueStats>("/admin/revenue/stats");

  const changePercent = stats?.change_percent || 0;
  const isPositive = changePercent >= 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Revenue Analytics</h2>
          <p className="text-muted-foreground">
            Track your platform&apos;s financial performance
          </p>
        </div>
        <Button variant="outline">
          <Download className="mr-2 h-4 w-4" />
          Export Report
        </Button>
      </div>

      {/* Revenue Stats */}
      <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Revenue
            </CardTitle>
            <DollarSign className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(stats?.total_revenue || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">Lifetime</p>
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
                <ArrowDownRight className="h-4 w-4 text-red-600" />
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
              Total Sales
            </CardTitle>
            <Users className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats?.total_sales || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">Enrollments</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Order Value
            </CardTitle>
            <BookOpen className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(stats?.avg_order_value || 0).toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">Per transaction</p>
          </CardContent>
        </Card>
      </div>

      {/* Chart Placeholder */}
      <Card>
        <CardHeader>
          <CardTitle>Revenue Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center bg-muted/30 rounded-lg">
            {stats?.monthly_data?.length ? (
              <div className="w-full h-full flex items-end justify-around p-4 gap-2">
                {stats.monthly_data.map((data, index) => {
                  const maxRevenue = Math.max(
                    ...stats.monthly_data.map((d) => d.revenue)
                  );
                  const height =
                    maxRevenue > 0 ? (data.revenue / maxRevenue) * 100 : 0;

                  return (
                    <div
                      key={index}
                      className="flex flex-col items-center gap-1"
                    >
                      <div
                        className="w-8 bg-primary rounded-t transition-all"
                        style={{ height: `${Math.max(height, 5)}%` }}
                        title={`$${data.revenue.toLocaleString()}`}
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
                <p>No revenue data yet</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Top Courses */}
      <Card>
        <CardHeader>
          <CardTitle>Top Performing Courses</CardTitle>
        </CardHeader>
        <CardContent>
          {stats?.top_courses?.length ? (
            <div className="space-y-4">
              {stats.top_courses.map((course, index) => (
                <div
                  key={course.id}
                  className="flex items-center justify-between p-3 bg-muted/50 rounded-lg"
                >
                  <div className="flex items-center gap-4">
                    <span className="text-lg font-semibold text-muted-foreground w-8">
                      #{index + 1}
                    </span>
                    <div>
                      <p className="font-medium">{course.title}</p>
                      <p className="text-sm text-muted-foreground">
                        {course.sales} sales
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">
                      ${course.revenue.toLocaleString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <BookOpen className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No course data available</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
