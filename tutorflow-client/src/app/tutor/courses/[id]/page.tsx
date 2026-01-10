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

interface TutorCourse extends Course {
  status: "draft" | "published" | "archived";
}

export default async function EditCoursePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let course: TutorCourse | null = null;
  try {
    course = await authServerFetch<TutorCourse>(`/courses/${id}`, {
      next: { tags: [`course-${id}`] },
    });
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
          <Link href="/tutor/courses">
            <ArrowLeft className="h-5 w-5" />
          </Link>
        </Button>
        <div>
          <h2 className="text-2xl font-bold">Edit Course</h2>
          <p className="text-muted-foreground">
            Update the details of "{course.title}"
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Course Details</CardTitle>
          <CardDescription>Basic information about your course</CardDescription>
        </CardHeader>
        <CardContent>
          <Suspense fallback={<div>Loading...</div>}>
            <CourseForm
              initialData={course}
              initialModules={(course.modules || []).map((m: any) => ({
                ...m,
                order: m.sort_order, // Map sort_order from API to order in UI
                isExpanded: false, // Default state
                lessons: (m.lessons || []).map((l: any) => ({
                  ...l,
                  order: l.sort_order,
                  duration_minutes: l.video_duration
                    ? Math.round(l.video_duration / 60)
                    : 0,
                  type: l.lesson_type || "text", // Map lesson_type
                })),
              }))}
            />
          </Suspense>
        </CardContent>
      </Card>
    </div>
  );
}
