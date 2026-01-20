import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { authServerFetch } from "@/lib/server-api";
import {
  Users,
  BookOpen,
  DollarSign,
  TrendingUp,
  ArrowUpRight,
  ArrowDownRight,
} from "lucide-react";

interface DashboardStats {
  total_users: number;
  total_courses: number;
  total_revenue: number;
  total_enrollments: number;
  users_change: number;
  courses_change: number;
  revenue_change: number;
  enrollments_change: number;
}

// Server Component
export default async function AdminDashboardPage() {
  const stats = await authServerFetch<DashboardStats>("/admin/dashboard");

  const cards = [
    {
      title: "Total Users",
      value: stats?.total_users?.toLocaleString() || "0",
      change: stats?.users_change || 0,
      icon: Users,
    },
    {
      title: "Total Courses",
      value: stats?.total_courses?.toLocaleString() || "0",
      change: stats?.courses_change || 0,
      icon: BookOpen,
    },
    {
      title: "Revenue",
      value: `$${(stats?.total_revenue || 0).toLocaleString()}`,
      change: stats?.revenue_change || 0,
      icon: DollarSign,
    },
    {
      title: "Enrollments",
      value: stats?.total_enrollments?.toLocaleString() || "0",
      change: stats?.enrollments_change || 0,
      icon: TrendingUp,
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Dashboard</h2>
        <p className="text-muted-foreground">
          Overview of your platform&apos;s performance
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {cards.map((card) => (
          <Card key={card.title}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {card.title}
              </CardTitle>
              <card.icon className="h-5 w-5 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{card.value}</div>
              <div className="flex items-center text-xs mt-1">
                {card.change >= 0 ? (
                  <ArrowUpRight className="h-4 w-4 text-green-600" />
                ) : (
                  <ArrowDownRight className="h-4 w-4 text-red-600" />
                )}
                <span
                  className={
                    card.change >= 0 ? "text-green-600" : "text-red-600"
                  }
                >
                  {Math.abs(card.change)}%
                </span>
                <span className="text-muted-foreground ml-1">
                  vs last month
                </span>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Recent Activity */}
      <div className="grid lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Recent Users</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8 text-muted-foreground">
              <Users className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No recent users</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Recent Enrollments</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8 text-muted-foreground">
              <BookOpen className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No recent enrollments</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
