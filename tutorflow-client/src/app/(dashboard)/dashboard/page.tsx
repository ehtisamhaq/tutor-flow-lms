import Link from "next/link";
import { BookOpen, Clock, Award, TrendingUp, Play } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch, type PaginatedResponse } from "@/lib/api";
import { redirect } from "next/navigation";

interface DashboardStats {
  enrolled_courses: number;
  hours_completed: number;
  certificates: number;
  average_progress: number;
}

interface Enrollment {
  id: string;
  progress_percent: number;
  course: {
    id: string;
    title: string;
    slug: string;
  };
}

// Server Component - dashboard home
export default async function DashboardPage() {
  // Fetch stats and recent enrollments in parallel
  const [statsData, enrollmentsData] = await Promise.all([
    authServerFetch<DashboardStats>("/enrollments/dashboard-stats").catch(
      () => null,
    ),
    authServerFetch<PaginatedResponse<Enrollment>>(
      "/enrollments/my?limit=1",
    ).catch(() => null),
  ]);

  const stats = [
    {
      title: "Enrolled Courses",
      value: (statsData?.enrolled_courses ?? 0).toString(),
      icon: BookOpen,
      description: "Active learning",
    },
    {
      title: "Hours Completed",
      value: (statsData?.hours_completed ?? 0).toString(),
      icon: Clock,
      description: "Learning time",
    },
    {
      title: "Certificates",
      value: (statsData?.certificates ?? 0).toString(),
      icon: Award,
      description: "Earned",
    },
    {
      title: "Avg. Progress",
      value: `${Math.round(statsData?.average_progress ?? 0)}%`,
      icon: TrendingUp,
      description: "Overall completion",
    },
  ];

  const recentEnrollment = enrollmentsData?.items?.[0];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Welcome back!</h2>
          <p className="text-muted-foreground">
            Here&apos;s what&apos;s happening with your learning
          </p>
        </div>
        <Button variant="outline">
          <Link href="/courses">Explore courses</Link>
        </Button>
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

      {/* Recent Activity / Continue Learning */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Continue Learning</CardTitle>
          {recentEnrollment && (
            <Button variant="ghost" size="sm">
              <Link href="/dashboard/my-courses">View all</Link>
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {recentEnrollment ? (
            <div className="flex flex-col md:flex-row gap-6 items-center bg-muted/30 p-4 rounded-lg border">
              <div className="flex-1 w-full space-y-4">
                <div>
                  <h3 className="font-semibold text-lg mb-1 leading-tight">
                    {recentEnrollment.course.title}
                  </h3>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <TrendingUp className="h-4 w-4" />
                    <span>
                      {Math.round(recentEnrollment.progress_percent)}% completed
                    </span>
                  </div>
                </div>

                <div className="h-3 bg-muted rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary transition-all duration-500"
                    style={{ width: `${recentEnrollment.progress_percent}%` }}
                  />
                </div>

                <Button className="w-full md:w-auto">
                  <Link href={`/learn/${recentEnrollment.course.slug}`}>
                    <Play className="mr-2 h-4 w-4" />
                    {recentEnrollment.progress_percent > 0
                      ? "Continue Learning"
                      : "Start Learning"}
                  </Link>
                </Button>
              </div>
            </div>
          ) : (
            <div className="text-center py-12 text-muted-foreground">
              <div className="inline-flex h-12 w-12 items-center justify-center rounded-full bg-muted mb-4 text-muted-foreground/50">
                <BookOpen className="h-6 w-6" />
              </div>
              <p className="font-medium text-foreground">
                No courses in progress
              </p>
              <p className="text-sm mb-6 max-w-xs mx-auto">
                Explore our catalog to find your next learning adventure.
              </p>
              <Button>
                <Link href="/courses">Browse Catalog</Link>
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
