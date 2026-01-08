import Link from "next/link";
import Image from "next/image";
import { serverApi, type Course } from "@/lib/server-api";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  GraduationCap,
  BookOpen,
  Clock,
  Award,
  ChevronRight,
} from "lucide-react";

interface LearningPath {
  id: string;
  title: string;
  slug: string;
  description: string;
  thumbnail_url?: string;
  level: "beginner" | "intermediate" | "advanced";
  duration_hours: number;
  course_count: number;
  courses: Course[];
}

async function getLearningPaths(): Promise<LearningPath[]> {
  const API_URL = process.env.API_URL || "http://localhost:8080/api/v1";

  try {
    const response = await fetch(`${API_URL}/learning-paths`, {
      next: { revalidate: 60 },
    });

    if (!response.ok) return [];
    const data = await response.json();
    return data.data?.items || [];
  } catch {
    return [];
  }
}

// Server Component
export default async function LearningPathsPage() {
  const paths = await getLearningPaths();

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <Link href="/" className="flex items-center gap-2">
            <GraduationCap className="h-8 w-8 text-primary" />
            <span className="text-xl font-bold">TutorFlow</span>
          </Link>
          <div className="flex items-center gap-3">
            <Button variant="ghost" asChild>
              <Link href="/login">Sign In</Link>
            </Button>
            <Button asChild>
              <Link href="/register">Get Started</Link>
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Learning Paths</h1>
          <p className="text-muted-foreground">
            Structured course sequences to master new skills
          </p>
        </div>

        {paths.length === 0 ? (
          <Card className="text-center py-16">
            <CardContent>
              <BookOpen className="h-16 w-16 mx-auto mb-4 text-muted-foreground" />
              <h2 className="text-xl font-semibold mb-2">
                No learning paths yet
              </h2>
              <p className="text-muted-foreground mb-6">
                Check back soon for curated learning paths
              </p>
              <Button asChild>
                <Link href="/courses">Browse Courses</Link>
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-6">
            {paths.map((path) => (
              <Card key={path.id} className="overflow-hidden">
                <div className="md:flex">
                  <div className="relative md:w-64 h-48 md:h-auto bg-muted shrink-0">
                    {path.thumbnail_url ? (
                      <Image
                        src={path.thumbnail_url}
                        alt={path.title}
                        fill
                        className="object-cover"
                      />
                    ) : (
                      <div className="absolute inset-0 flex items-center justify-center">
                        <Award className="h-12 w-12 text-muted-foreground" />
                      </div>
                    )}
                  </div>
                  <CardContent className="p-6 flex-1">
                    <div className="flex items-start justify-between">
                      <div>
                        <span className="text-xs px-2 py-1 bg-primary/10 text-primary rounded capitalize mb-2 inline-block">
                          {path.level}
                        </span>
                        <h2 className="text-xl font-semibold mb-2">
                          {path.title}
                        </h2>
                        <p className="text-muted-foreground mb-4 line-clamp-2">
                          {path.description}
                        </p>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <BookOpen className="h-4 w-4" />
                            {path.course_count} courses
                          </span>
                          <span className="flex items-center gap-1">
                            <Clock className="h-4 w-4" />
                            {path.duration_hours} hours
                          </span>
                          <span className="flex items-center gap-1">
                            <Award className="h-4 w-4" />
                            Certificate
                          </span>
                        </div>
                      </div>
                      <Button asChild>
                        <Link href={`/learning-paths/${path.slug}`}>
                          View Path
                          <ChevronRight className="ml-2 h-4 w-4" />
                        </Link>
                      </Button>
                    </div>
                  </CardContent>
                </div>
              </Card>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
