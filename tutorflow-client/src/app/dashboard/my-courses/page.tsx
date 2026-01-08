import Link from "next/link";
import Image from "next/image";
import { redirect } from "next/navigation";
import { GraduationCap, Play, Clock, CheckCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { authServerFetch } from "@/lib/server-api";

interface Enrollment {
  id: string;
  progress: number;
  completed_at?: string;
  course: {
    id: string;
    title: string;
    slug: string;
    thumbnail_url?: string;
    duration_hours: number;
    instructor: {
      first_name: string;
      last_name: string;
    };
  };
}

// Server Component with SSR data fetching
export default async function MyCoursesPage() {
  const enrollments = await authServerFetch<Enrollment[]>("/enrollments/my");

  // Redirect to login if not authenticated
  if (enrollments === null) {
    redirect("/login");
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">My Courses</h2>
          <p className="text-muted-foreground">
            {enrollments.length} course{enrollments.length !== 1 ? "s" : ""}{" "}
            enrolled
          </p>
        </div>
        <Button asChild>
          <Link href="/courses">Browse More</Link>
        </Button>
      </div>

      {enrollments.length === 0 ? (
        <Card className="text-center py-16">
          <CardContent>
            <GraduationCap className="h-16 w-16 mx-auto mb-4 text-muted-foreground" />
            <h3 className="text-xl font-semibold mb-2">No courses yet</h3>
            <p className="text-muted-foreground mb-6">
              Start your learning journey by enrolling in a course
            </p>
            <Button asChild>
              <Link href="/courses">Browse Courses</Link>
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {enrollments.map((enrollment) => (
            <Card
              key={enrollment.id}
              className="overflow-hidden hover:shadow-lg transition-shadow"
            >
              <div className="relative aspect-video bg-muted">
                {enrollment.course.thumbnail_url ? (
                  <Image
                    src={enrollment.course.thumbnail_url}
                    alt={enrollment.course.title}
                    fill
                    className="object-cover"
                  />
                ) : (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <GraduationCap className="h-12 w-12 text-muted-foreground" />
                  </div>
                )}
                {enrollment.completed_at && (
                  <div className="absolute top-2 right-2 bg-green-600 text-white text-xs px-2 py-1 rounded-full flex items-center gap-1">
                    <CheckCircle className="h-3 w-3" />
                    Completed
                  </div>
                )}
              </div>
              <CardContent className="p-4">
                <h3 className="font-semibold line-clamp-2 mb-1">
                  {enrollment.course.title}
                </h3>
                <p className="text-sm text-muted-foreground mb-3">
                  {enrollment.course.instructor.first_name}{" "}
                  {enrollment.course.instructor.last_name}
                </p>

                {/* Progress Bar */}
                <div className="mb-3">
                  <div className="flex justify-between text-xs text-muted-foreground mb-1">
                    <span>{enrollment.progress}% complete</span>
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {enrollment.course.duration_hours}h
                    </span>
                  </div>
                  <div className="h-2 bg-muted rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary transition-all"
                      style={{ width: `${enrollment.progress}%` }}
                    />
                  </div>
                </div>

                <Button className="w-full" asChild>
                  <Link href={`/learn/${enrollment.course.slug}`}>
                    <Play className="mr-2 h-4 w-4" />
                    {enrollment.progress > 0 ? "Continue" : "Start Learning"}
                  </Link>
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
