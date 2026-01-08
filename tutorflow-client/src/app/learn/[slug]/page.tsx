import Link from "next/link";
import { cookies } from "next/headers";
import { redirect, notFound } from "next/navigation";
import { GraduationCap } from "lucide-react";
import { LearningContent } from "@/components/learn/learning-content";

interface Module {
  id: string;
  title: string;
  order: number;
  lessons: Lesson[];
}

interface Lesson {
  id: string;
  title: string;
  type: "video" | "text" | "quiz";
  duration_minutes: number;
  order: number;
  is_preview: boolean;
}

interface Enrollment {
  id: string;
  progress: number;
  course: {
    id: string;
    title: string;
    slug: string;
    modules: Module[];
  };
}

async function getEnrollment(slug: string): Promise<Enrollment | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("accessToken");

  if (!token) {
    redirect("/login");
  }

  try {
    const response = await fetch(
      `${
        process.env.API_URL || "http://localhost:8080/api/v1"
      }/enrollments/course/${slug}`,
      {
        headers: {
          Authorization: `Bearer ${token.value}`,
        },
        cache: "no-store",
      }
    );

    if (!response.ok) return null;
    const data = await response.json();
    return data.data;
  } catch {
    return null;
  }
}

// Server Component with SSR
export default async function LearnPage({
  params,
  searchParams,
}: {
  params: Promise<{ slug: string }>;
  searchParams: Promise<{ lesson?: string }>;
}) {
  const { slug } = await params;
  const { lesson: lessonId } = await searchParams;

  const enrollment = await getEnrollment(slug);

  // If no enrollment, redirect to course page
  if (!enrollment) {
    redirect(`/courses/${slug}`);
  }

  // Get first lesson if none selected
  const firstLesson = enrollment.course.modules[0]?.lessons[0];
  const currentLessonId = lessonId || firstLesson?.id;

  // Find current lesson
  let currentLesson: Lesson | undefined;
  for (const module of enrollment.course.modules) {
    for (const lesson of module.lessons) {
      if (lesson.id === currentLessonId) {
        currentLesson = lesson;
        break;
      }
    }
  }

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur">
        <div className="flex h-14 items-center justify-between px-4">
          <div className="flex items-center gap-4">
            <Link href="/dashboard" className="flex items-center gap-2">
              <GraduationCap className="h-6 w-6 text-primary" />
            </Link>
            <div className="h-4 w-px bg-border" />
            <h1 className="text-sm font-medium truncate max-w-[300px]">
              {enrollment.course.title}
            </h1>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-sm text-muted-foreground">
              {enrollment.progress}% complete
            </div>
            <Link
              href={`/courses/${slug}`}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Course Info
            </Link>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <LearningContent
        courseSlug={slug}
        courseId={enrollment.course.id}
        modules={enrollment.course.modules}
        currentLessonId={currentLessonId}
        currentLesson={currentLesson}
      />
    </div>
  );
}
