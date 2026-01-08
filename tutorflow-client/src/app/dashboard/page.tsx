import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BookOpen, Clock, Award, TrendingUp } from "lucide-react";

// Server Component - dashboard home
export default function DashboardPage() {
  const stats = [
    {
      title: "Enrolled Courses",
      value: "12",
      icon: BookOpen,
      description: "Active learning",
    },
    {
      title: "Hours Completed",
      value: "48",
      icon: Clock,
      description: "This month",
    },
    {
      title: "Certificates",
      value: "3",
      icon: Award,
      description: "Earned",
    },
    {
      title: "Progress",
      value: "67%",
      icon: TrendingUp,
      description: "Average",
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Welcome back!</h2>
        <p className="text-muted-foreground">
          Here&apos;s what&apos;s happening with your learning
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <Card key={stat.title}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {stat.title}
              </CardTitle>
              <stat.icon className="h-5 w-5 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground">
                {stat.description}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Continue Learning</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            <BookOpen className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>No courses in progress</p>
            <p className="text-sm">Start learning by enrolling in a course</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
