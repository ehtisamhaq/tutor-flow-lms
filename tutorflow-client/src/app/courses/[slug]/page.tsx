import Link from "next/link";
import Image from "next/image";
import { notFound } from "next/navigation";
import {
  Star,
  Clock,
  Users,
  GraduationCap,
  Play,
  CheckCircle,
  BookOpen,
  Award,
  FileText,
} from "lucide-react";
import { serverApi } from "@/lib/server-api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AddToCartButton } from "@/components/course/add-to-cart-button";
import { BuyNowButton } from "@/components/course/buy-now-button";

// Server Component with SSR data fetching
export default async function CourseDetailPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const course = await serverApi.courses.getBySlug(slug);

  if (!course) {
    notFound();
  }

  const price = course.discount_price ?? course.price;
  const hasDiscount =
    course.discount_price && course.discount_price < course.price;
  const discountPercent = hasDiscount
    ? Math.round(((course.price - course.discount_price!) / course.price) * 100)
    : 0;

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

      {/* Course Hero */}
      <section className="bg-gradient-to-r from-slate-900 to-slate-800 text-white py-12">
        <div className="container mx-auto px-4">
          <div className="grid lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2">
              <nav className="text-sm text-slate-400 mb-4">
                <Link href="/courses" className="hover:text-white">
                  Courses
                </Link>
                <span className="mx-2">/</span>
                <span className="capitalize">{course.level}</span>
              </nav>

              <h1 className="text-3xl md:text-4xl font-bold mb-4">
                {course.title}
              </h1>

              <p className="text-lg text-slate-300 mb-6">
                {course.description}
              </p>

              <div className="flex flex-wrap items-center gap-4 text-sm">
                <div className="flex items-center gap-1">
                  <Star className="h-5 w-5 fill-yellow-400 text-yellow-400" />
                  <span className="font-semibold">
                    {course.rating.toFixed(1)}
                  </span>
                </div>
                <div className="flex items-center gap-1">
                  <Users className="h-5 w-5" />
                  <span>{course.total_students.toLocaleString()} students</span>
                </div>
                <div className="flex items-center gap-1">
                  <Clock className="h-5 w-5" />
                  <span>{course.duration_hours} hours</span>
                </div>
              </div>

              <div className="flex items-center gap-3 mt-6">
                {course.instructor.avatar_url ? (
                  <Image
                    src={course.instructor.avatar_url}
                    alt={`${course.instructor.first_name} ${course.instructor.last_name}`}
                    width={48}
                    height={48}
                    className="rounded-full"
                  />
                ) : (
                  <div className="h-12 w-12 rounded-full bg-slate-700 flex items-center justify-center">
                    <span className="text-lg font-semibold">
                      {course.instructor.first_name[0]}
                      {course.instructor.last_name[0]}
                    </span>
                  </div>
                )}
                <div>
                  <p className="text-sm text-slate-400">Instructor</p>
                  <p className="font-medium">
                    {course.instructor.first_name} {course.instructor.last_name}
                  </p>
                </div>
              </div>
            </div>

            {/* Pricing Card */}
            <div className="lg:col-span-1">
              <Card className="sticky top-24">
                <div className="relative aspect-video bg-muted overflow-hidden rounded-t-lg">
                  {course.thumbnail_url ? (
                    <Image
                      src={course.thumbnail_url}
                      alt={course.title}
                      fill
                      className="object-cover"
                    />
                  ) : (
                    <div className="absolute inset-0 flex items-center justify-center">
                      <Play className="h-16 w-16 text-muted-foreground" />
                    </div>
                  )}
                </div>
                <CardContent className="p-6">
                  <div className="flex items-baseline gap-2 mb-4">
                    <span className="text-3xl font-bold">
                      ${price.toFixed(2)}
                    </span>
                    {hasDiscount && (
                      <>
                        <span className="text-lg text-muted-foreground line-through">
                          ${course.price.toFixed(2)}
                        </span>
                        <span className="text-sm font-medium text-green-600">
                          {discountPercent}% off
                        </span>
                      </>
                    )}
                  </div>

                  <div className="space-y-3">
                    <AddToCartButton courseId={course.id} />
                    <BuyNowButton courseId={course.id} />
                  </div>

                  <div className="mt-6 space-y-3 text-sm">
                    <div className="flex items-center gap-2">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                      <span>Lifetime access</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                      <span>Certificate of completion</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                      <span>30-day money-back guarantee</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </section>

      {/* Course Content */}
      <section className="py-12">
        <div className="container mx-auto px-4">
          <div className="grid lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2 space-y-8">
              {/* Curriculum Area */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <BookOpen className="h-5 w-5" />
                    Course Content
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-0">
                  {course.modules?.length ? (
                    <div className="divide-y">
                      {course.modules
                        .sort((a, b) => a.sort_order - b.sort_order)
                        .map((module) => (
                          <div key={module.id} className="p-4">
                            <h4 className="font-semibold mb-3 flex items-center justify-between">
                              {module.title}
                              <span className="text-xs font-normal text-muted-foreground whitespace-nowrap">
                                {module.lessons?.length || 0} lessons
                              </span>
                            </h4>
                            <div className="space-y-2">
                              {module.lessons
                                ?.sort((a, b) => a.sort_order - b.sort_order)
                                .map((lesson) => (
                                  <div
                                    key={lesson.id}
                                    className="flex items-center justify-between text-sm py-1 hover:text-primary transition-colors group cursor-default"
                                  >
                                    <div className="flex items-center gap-2">
                                      {lesson.type === "video" ? (
                                        <Play className="h-4 w-4 text-muted-foreground group-hover:text-primary" />
                                      ) : (
                                        <FileText className="h-4 w-4 text-muted-foreground group-hover:text-primary" />
                                      )}
                                      <span>{lesson.title}</span>
                                    </div>
                                    {lesson.duration_minutes > 0 && (
                                      <span className="text-muted-foreground text-xs">
                                        {lesson.duration_minutes}m
                                      </span>
                                    )}
                                  </div>
                                ))}
                            </div>
                          </div>
                        ))}
                    </div>
                  ) : (
                    <div className="p-8 text-center text-muted-foreground italic">
                      No curriculum available yet.
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* What you'll learn */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <CheckCircle className="h-5 w-5" />
                    What You&apos;ll Learn
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid sm:grid-cols-2 gap-3">
                    {[
                      "Build real-world projects from scratch",
                      "Master core concepts and best practices",
                      "Learn industry-standard tools and workflows",
                      "Get hands-on experience with practical exercises",
                      "Understand advanced patterns and techniques",
                      "Prepare for professional certification",
                    ].map((item) => (
                      <div key={item} className="flex items-start gap-2">
                        <CheckCircle className="h-5 w-5 text-green-600 mt-0.5 shrink-0" />
                        <span className="text-sm">{item}</span>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              {/* Requirements */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Award className="h-5 w-5" />
                    Requirements
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="space-y-2 text-sm">
                    <li className="flex items-start gap-2">
                      <span className="text-muted-foreground">•</span>
                      Basic computer skills and internet access
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-muted-foreground">•</span>
                      Eagerness to learn and practice
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-muted-foreground">•</span>
                      No prior experience required
                    </li>
                  </ul>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
