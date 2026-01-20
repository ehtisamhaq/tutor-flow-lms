import { Suspense } from "react";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { CourseForm } from "@/components/tutor/course-form";
import { authServerFetch, type Course } from "@/lib/server-api";

interface AdminCourse extends Course {
  status: "draft" | "published" | "archived";
}

export default async function AdminEditCoursePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let course: AdminCourse | null = null;
  try {
    // Admin can fetch any course
    course = await authServerFetch<AdminCourse>(`/admin/courses/${id}`);
  } catch (error) {
    notFound();
  }

  if (!course) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link href="/admin/courses">
            <ArrowLeft className="h-5 w-5" />
          </Link>
        </Button>
        <div>
          <h2 className="text-2xl font-bold">Edit Course (Admin)</h2>
          <p className="text-muted-foreground">
            Administrative update for "{course.title}"
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Course Details</CardTitle>
          <CardDescription>
            Update course information as an administrator
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Suspense fallback={<div>Loading...</div>}>
            <CourseForm initialData={course} />
          </Suspense>
        </CardContent>
      </Card>
    </div>
  );
}
