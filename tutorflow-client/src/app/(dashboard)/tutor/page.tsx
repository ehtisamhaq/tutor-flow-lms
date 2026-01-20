import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { authServerFetch } from "@/lib/server-api";
import {
  BookOpen,
  DollarSign,
  Users,
  Star,
  TrendingUp,
  ArrowUpRight,
} from "lucide-react";

interface TutorStats {
  total_courses: number;
  total_students: number;
  total_earnings: number;
  average_rating: number;
  this_month_earnings: number;
  earnings_change: number;
}

// Server Component
export default async function TutorDashboardPage() {
  const stats = await authServerFetch<TutorStats>("/tutors/dashboard/stats");

  const cards = [
    {
      title: "My Courses",
      value: stats?.total_courses?.toString() || "0",
      icon: BookOpen,
      description: "Published courses",
    },
    {
      title: "Total Students",
      value: stats?.total_students?.toLocaleString() || "0",
      icon: Users,
      description: "Enrolled students",
    },
    {
      title: "Total Earnings",
      value: `$${(stats?.total_earnings || 0).toLocaleString()}`,
      icon: DollarSign,
      description: "Lifetime earnings",
    },
    {
      title: "Avg Rating",
      value: stats?.average_rating?.toFixed(1) || "0.0",
      icon: Star,
      description: "Course rating",
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Instructor Dashboard</h2>
        <p className="text-muted-foreground">
          Overview of your courses and earnings
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
              <p className="text-xs text-muted-foreground">
                {card.description}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* This Month */}
      <Card>
        <CardHeader>
          <CardTitle>This Month&apos;s Earnings</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="text-4xl font-bold">
              ${(stats?.this_month_earnings || 0).toLocaleString()}
            </div>
            {stats?.earnings_change !== undefined && (
              <div className="flex items-center text-green-600">
                <ArrowUpRight className="h-5 w-5" />
                <span className="font-medium">
                  {stats.earnings_change}% vs last month
                </span>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions */}
      <div className="grid lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8 text-muted-foreground">
              <TrendingUp className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No recent activity</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Recent Reviews</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8 text-muted-foreground">
              <Star className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No recent reviews</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
