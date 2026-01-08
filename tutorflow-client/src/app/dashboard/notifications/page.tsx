import Link from "next/link";
import { redirect } from "next/navigation";
import { Bell, BookOpen, MessageSquare, Award, Check } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch, type PaginatedResponse } from "@/lib/server-api";

interface Notification {
  id: string;
  type: "course" | "message" | "certificate" | "announcement" | "system";
  title: string;
  message: string;
  read: boolean;
  created_at: string;
  data?: {
    course_id?: string;
    course_slug?: string;
  };
}

function getNotificationIcon(type: string) {
  switch (type) {
    case "course":
      return BookOpen;
    case "message":
      return MessageSquare;
    case "certificate":
      return Award;
    default:
      return Bell;
  }
}

function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return "Just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
  return date.toLocaleDateString();
}

// Server Component
export default async function NotificationsPage() {
  const result = await authServerFetch<PaginatedResponse<Notification>>(
    "/notifications"
  );
  const notifications = result?.items || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Notifications</h2>
          <p className="text-muted-foreground">
            Stay updated on your courses and activities
          </p>
        </div>
        <Button variant="outline">
          <Check className="mr-2 h-4 w-4" />
          Mark All Read
        </Button>
      </div>

      {notifications.length === 0 ? (
        <Card className="text-center py-16">
          <CardContent>
            <Bell className="h-16 w-16 mx-auto mb-4 text-muted-foreground" />
            <h3 className="text-xl font-semibold mb-2">No notifications</h3>
            <p className="text-muted-foreground">You&apos;re all caught up!</p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Recent</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y">
              {notifications.map((notification) => {
                const Icon = getNotificationIcon(notification.type);

                return (
                  <div
                    key={notification.id}
                    className={`flex items-start gap-4 p-4 ${
                      !notification.read ? "bg-primary/5" : ""
                    }`}
                  >
                    <div
                      className={`h-10 w-10 rounded-full flex items-center justify-center shrink-0 ${
                        !notification.read
                          ? "bg-primary/10 text-primary"
                          : "bg-muted text-muted-foreground"
                      }`}
                    >
                      <Icon className="h-5 w-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-2">
                        <p className="font-medium">{notification.title}</p>
                        <span className="text-xs text-muted-foreground shrink-0">
                          {formatTimeAgo(notification.created_at)}
                        </span>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        {notification.message}
                      </p>
                      {notification.data?.course_slug && (
                        <Link
                          href={`/courses/${notification.data.course_slug}`}
                          className="text-sm text-primary hover:underline mt-2 inline-block"
                        >
                          View Course →
                        </Link>
                      )}
                    </div>
                    {!notification.read && (
                      <div className="h-2 w-2 rounded-full bg-primary shrink-0 mt-2" />
                    )}
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
