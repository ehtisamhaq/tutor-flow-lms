import Link from "next/link";
import Image from "next/image";
import {
  authServerFetch,
  type PaginatedResponse,
  type Course,
} from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { BookOpen, Eye, EyeOff, Plus, Star, Users } from "lucide-react";
import { CourseSearch } from "@/components/admin/course-search";
import { CourseActions } from "@/components/admin/course-actions";

interface AdminCourse extends Course {
  status: "draft" | "published" | "archived";
  created_at: string;
}

// Server Component
export default async function AdminCoursesPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; status?: string; search?: string }>;
}) {
  const params = await searchParams;
  const page = Number(params.page) || 1;
  const status = params.status || "";
  const search = params.search || "";

  const courses = await authServerFetch<PaginatedResponse<AdminCourse>>(
    `/admin/courses?page=${page}&limit=20${status ? `&status=${status}` : ""}${
      search ? `&search=${search}` : ""
    }`
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Courses</h2>
          <p className="text-muted-foreground">
            Manage all courses on the platform
          </p>
        </div>
        <Button asChild>
          <Link href="/admin/courses/new">
            <Plus className="mr-2 h-4 w-4" />
            Add Course
          </Link>
        </Button>
      </div>

      {/* Filters & Search */}
      <div className="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between">
        <div className="flex gap-2">
          {["", "draft", "published", "archived"].map((s) => (
            <Button
              key={s}
              variant={status === s ? "default" : "outline"}
              size="sm"
              asChild
            >
              <Link href={`/admin/courses${s ? `?status=${s}` : ""}`}>
                {s || "All"}
              </Link>
            </Button>
          ))}
        </div>
        <CourseSearch />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{courses?.total?.toLocaleString() || 0} Courses</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {courses?.items?.length ? (
            <div className="divide-y">
              {courses.items.map((course) => (
                <div
                  key={course.id}
                  className="flex items-center gap-4 p-4 hover:bg-muted/50"
                >
                  <div className="relative w-24 h-16 rounded-lg bg-muted overflow-hidden shrink-0">
                    {course.thumbnail_url ? (
                      <Image
                        src={course.thumbnail_url}
                        alt={course.title}
                        fill
                        className="object-cover"
                      />
                    ) : (
                      <div className="absolute inset-0 flex items-center justify-center">
                        <BookOpen className="h-6 w-6 text-muted-foreground" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <Link
                      href={`/admin/courses/${course.id}`}
                      className="font-medium hover:text-primary line-clamp-1"
                    >
                      {course.title}
                    </Link>
                    <div className="flex items-center gap-4 text-sm text-muted-foreground mt-1">
                      <span>
                        {course.instructor.first_name}{" "}
                        {course.instructor.last_name}
                      </span>
                      <span className="flex items-center gap-1">
                        <Star className="h-3 w-3 fill-yellow-400 text-yellow-400" />
                        {course.rating.toFixed(1)}
                      </span>
                      <span className="flex items-center gap-1">
                        <Users className="h-3 w-3" />
                        {course.total_students}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="font-medium">
                      ${course.price.toFixed(2)}
                    </span>
                    <span
                      className={`flex items-center gap-1 text-xs px-2 py-1 rounded-full ${
                        course.status === "published"
                          ? "bg-green-100 text-green-800"
                          : course.status === "draft"
                          ? "bg-yellow-100 text-yellow-800"
                          : "bg-gray-100 text-gray-800"
                      }`}
                    >
                      {course.status === "published" ? (
                        <Eye className="h-3 w-3" />
                      ) : (
                        <EyeOff className="h-3 w-3" />
                      )}
                      {course.status}
                    </span>
                    <CourseActions
                      courseId={course.id}
                      currentStatus={course.status}
                    />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-16">
              <BookOpen className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <p className="text-muted-foreground">No courses found</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
