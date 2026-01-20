import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { authServerFetch } from "@/lib/server-api";
import {
  Users,
  BookOpen,
  TrendingUp,
  Clock,
  Eye,
  PlayCircle,
  Award,
  Target,
} from "lucide-react";

interface AnalyticsStats {
  total_watch_time_hours: number;
  avg_session_duration_minutes: number;
  completion_rate: number;
  active_users_today: number;
  active_users_week: number;
  popular_courses: {
    id: string;
    title: string;
    views: number;
    watch_time_hours: number;
  }[];
  engagement_by_day: {
    day: string;
    users: number;
    sessions: number;
  }[];
  device_breakdown: {
    device: string;
    percent: number;
  }[];
}

// Server Component
export default async function AdminAnalyticsPage() {
  const stats = await authServerFetch<AnalyticsStats>("/admin/analytics");

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Analytics</h2>
        <p className="text-muted-foreground">
          User engagement and learning metrics
        </p>
      </div>

      {/* Key Metrics */}
      <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Watch Time
            </CardTitle>
            <Clock className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats?.total_watch_time_hours || 0).toLocaleString()}h
            </div>
            <p className="text-xs text-muted-foreground">All time</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Session
            </CardTitle>
            <PlayCircle className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats?.avg_session_duration_minutes || 0).toFixed(0)} min
            </div>
            <p className="text-xs text-muted-foreground">Per user</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Completion Rate
            </CardTitle>
            <Award className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats?.completion_rate || 0).toFixed(1)}%
            </div>
            <p className="text-xs text-muted-foreground">Courses completed</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Today
            </CardTitle>
            <Users className="h-5 w-5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats?.active_users_today || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">
              {(stats?.active_users_week || 0).toLocaleString()} this week
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Charts Row */}
      <div className="grid lg:grid-cols-2 gap-6">
        {/* Engagement Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Daily Engagement</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-48 flex items-end justify-around gap-2">
              {(stats?.engagement_by_day || []).map((data, index) => (
                <div key={index} className="flex flex-col items-center gap-1">
                  <div className="flex gap-1 h-32">
                    <div
                      className="w-4 bg-primary/50 rounded-t"
                      style={{
                        height: `${Math.min((data.users / 100) * 100, 100)}%`,
                      }}
                      title={`${data.users} users`}
                    />
                    <div
                      className="w-4 bg-primary rounded-t"
                      style={{
                        height: `${Math.min(
                          (data.sessions / 200) * 100,
                          100
                        )}%`,
                      }}
                      title={`${data.sessions} sessions`}
                    />
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {data.day.slice(0, 3)}
                  </span>
                </div>
              ))}
              {!stats?.engagement_by_day?.length && (
                <div className="flex-1 flex items-center justify-center text-muted-foreground">
                  <TrendingUp className="h-12 w-12 opacity-50" />
                </div>
              )}
            </div>
            <div className="flex justify-center gap-4 mt-4 text-xs">
              <span className="flex items-center gap-1">
                <span className="w-3 h-3 bg-primary/50 rounded" /> Users
              </span>
              <span className="flex items-center gap-1">
                <span className="w-3 h-3 bg-primary rounded" /> Sessions
              </span>
            </div>
          </CardContent>
        </Card>

        {/* Device Breakdown */}
        <Card>
          <CardHeader>
            <CardTitle>Device Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {(
                stats?.device_breakdown || [
                  { device: "Desktop", percent: 55 },
                  { device: "Mobile", percent: 35 },
                  { device: "Tablet", percent: 10 },
                ]
              ).map((item) => (
                <div key={item.device} className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>{item.device}</span>
                    <span className="font-medium">{item.percent}%</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary transition-all"
                      style={{ width: `${item.percent}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Popular Courses */}
      <Card>
        <CardHeader>
          <CardTitle>Most Viewed Courses</CardTitle>
        </CardHeader>
        <CardContent>
          {stats?.popular_courses?.length ? (
            <div className="space-y-3">
              {stats.popular_courses.map((course, index) => (
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
                      <p className="text-sm text-muted-foreground flex items-center gap-2">
                        <Eye className="h-3 w-3" />
                        {course.views.toLocaleString()} views
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-medium">
                      {course.watch_time_hours.toFixed(0)}h
                    </div>
                    <p className="text-xs text-muted-foreground">watch time</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Target className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No analytics data yet</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
