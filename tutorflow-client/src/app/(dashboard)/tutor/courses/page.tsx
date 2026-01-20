import Link from "next/link";
import Image from "next/image";
import {
  authServerFetch,
  type Course,
  type PaginatedResponse,
} from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  BookOpen,
  Plus,
  Edit,
  Eye,
  EyeOff,
  Star,
  Users,
  MoreHorizontal,
} from "lucide-react";

interface TutorCourse extends Course {
  status: "draft" | "published" | "archived";
  total_revenue: number;
}

// Server Component
export default async function TutorCoursesPage() {
  const data = await authServerFetch<PaginatedResponse<TutorCourse>>(
    "/courses/my"
  );
  const courses = data?.items || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">My Courses</h2>
          <p className="text-muted-foreground">
            Create and manage your courses
          </p>
        </div>
        <Button asChild>
          <Link href="/tutor/courses/new">
            <Plus className="mr-2 h-4 w-4" />
            Create Course
          </Link>
        </Button>
      </div>

      {!courses.length ? (
        <Card className="text-center py-16">
          <CardContent>
            <BookOpen className="h-16 w-16 mx-auto mb-4 text-muted-foreground" />
            <h3 className="text-xl font-semibold mb-2">No courses yet</h3>
            <p className="text-muted-foreground mb-6">
              Create your first course and start teaching
            </p>
            <Button asChild>
              <Link href="/tutor/courses/new">
                <Plus className="mr-2 h-4 w-4" />
                Create Course
              </Link>
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {courses.map((course) => (
            <Card key={course.id} className="overflow-hidden">
              <div className="relative aspect-video bg-muted">
                {course.thumbnail_url ? (
                  <Image
                    src={course.thumbnail_url}
                    alt={course.title}
                    fill
                    className="object-cover"
                  />
                ) : (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <BookOpen className="h-12 w-12 text-muted-foreground" />
                  </div>
                )}
                <div className="absolute top-2 right-2">
                  <span
                    className={`flex items-center gap-1 text-xs px-2 py-1 rounded-full ${
                      course.status === "published"
                        ? "bg-green-600 text-white"
                        : course.status === "draft"
                        ? "bg-yellow-600 text-white"
                        : "bg-gray-600 text-white"
                    }`}
                  >
                    {course.status === "published" ? (
                      <Eye className="h-3 w-3" />
                    ) : (
                      <EyeOff className="h-3 w-3" />
                    )}
                    {course.status}
                  </span>
                </div>
              </div>
              <CardContent className="p-4">
                <h3 className="font-semibold line-clamp-2 mb-2">
                  {course.title}
                </h3>
                <div className="flex items-center gap-4 text-sm text-muted-foreground mb-3">
                  <span className="flex items-center gap-1">
                    <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                    {course.rating.toFixed(1)}
                  </span>
                  <span className="flex items-center gap-1">
                    <Users className="h-4 w-4" />
                    {course.total_students}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-medium">
                    ${course.price.toFixed(2)}
                  </span>
                  <div className="flex gap-2">
                    <Button size="sm" variant="outline" asChild>
                      <Link href={`/tutor/courses/${course.id}`}>
                        <Edit className="h-4 w-4" />
                      </Link>
                    </Button>
                    <Button size="sm" variant="ghost">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
