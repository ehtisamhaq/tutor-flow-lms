import Link from "next/link";
import Image from "next/image";
import { Star, Clock, Users, GraduationCap } from "lucide-react";
import { serverApi, type Course } from "@/lib/server-api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter } from "@/components/ui/card";

// Server Component with SSR data fetching
export default async function CoursesPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; search?: string }>;
}) {
  const params = await searchParams;
  const page = Number(params.page) || 1;
  const search = params.search || "";

  const data = await serverApi.courses.list({ page, limit: 12, search });
  const courses = data?.items || [];

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <Link href="/" className="flex items-center gap-2">
            <GraduationCap className="h-8 w-8 text-primary" />
            <span className="text-xl font-bold">TutorFlow</span>
          </Link>

          <nav className="hidden md:flex items-center gap-6">
            <Link
              href="/courses"
              className="text-sm font-medium text-foreground"
            >
              Courses
            </Link>
            <Link
              href="/learning-paths"
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              Learning Paths
            </Link>
          </nav>

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

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Explore Courses</h1>
          <p className="text-muted-foreground">
            Browse our collection of courses taught by expert instructors
          </p>
        </div>

        {/* Course Grid */}
        {courses.length > 0 ? (
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {courses.map((course) => (
              <CourseCard key={course.id} course={course} />
            ))}
          </div>
        ) : (
          <div className="text-center py-16">
            <p className="text-muted-foreground mb-4">No courses found</p>
            <Button asChild>
              <Link href="/courses">View All Courses</Link>
            </Button>
          </div>
        )}

        {/* Pagination */}
        {data && data.total > 12 && (
          <div className="flex justify-center gap-2 mt-8">
            {page > 1 && (
              <Button variant="outline" asChild>
                <Link
                  href={`/courses?page=${page - 1}${
                    search ? `&search=${search}` : ""
                  }`}
                >
                  Previous
                </Link>
              </Button>
            )}
            {page * 12 < data.total && (
              <Button variant="outline" asChild>
                <Link
                  href={`/courses?page=${page + 1}${
                    search ? `&search=${search}` : ""
                  }`}
                >
                  Next
                </Link>
              </Button>
            )}
          </div>
        )}
      </main>
    </div>
  );
}

function CourseCard({ course }: { course: Course }) {
  const price = course.discount_price ?? course.price;
  const hasDiscount =
    course.discount_price && course.discount_price < course.price;

  return (
    <Card className="overflow-hidden hover:shadow-lg transition-shadow">
      <Link href={`/courses/${course.slug}`}>
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
              <GraduationCap className="h-12 w-12 text-muted-foreground" />
            </div>
          )}
        </div>
      </Link>
      <CardContent className="p-4">
        <Link href={`/courses/${course.slug}`}>
          <h3 className="font-semibold line-clamp-2 hover:text-primary transition-colors">
            {course.title}
          </h3>
        </Link>
        <p className="text-sm text-muted-foreground mt-1">
          {course.instructor.first_name} {course.instructor.last_name}
        </p>
        <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
          <div className="flex items-center gap-1">
            <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
            {course.rating.toFixed(1)}
          </div>
          <div className="flex items-center gap-1">
            <Users className="h-4 w-4" />
            {course.total_students}
          </div>
          <div className="flex items-center gap-1">
            <Clock className="h-4 w-4" />
            {course.duration_hours}h
          </div>
        </div>
      </CardContent>
      <CardFooter className="p-4 pt-0 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-lg font-bold">${price.toFixed(2)}</span>
          {hasDiscount && (
            <span className="text-sm text-muted-foreground line-through">
              ${course.price.toFixed(2)}
            </span>
          )}
        </div>
        <span className="text-xs px-2 py-1 bg-muted rounded capitalize">
          {course.level}
        </span>
      </CardFooter>
    </Card>
  );
}
