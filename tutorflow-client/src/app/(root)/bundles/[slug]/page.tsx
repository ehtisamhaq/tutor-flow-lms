import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { serverFetch } from "@/lib/server-api";
import { Package, Check, Clock, Users, Star } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";

interface BundleCourse {
  id: string;
  title: string;
  slug: string;
  thumbnail_url?: string;
  description: string;
  price: number;
  rating: number;
  students_count: number;
  instructor: {
    name: string;
  };
}

interface Bundle {
  id: string;
  title: string;
  slug: string;
  description: string;
  thumbnail_url?: string;
  original_price: number;
  bundle_price: number;
  discount_percent: number;
  courses: {
    course: BundleCourse;
  }[];
  end_date?: string;
  purchase_count: number;
}

// Server Component
export default async function BundleDetailPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const bundle = await serverFetch<Bundle>(`/bundles/${slug}`);

  if (!bundle) {
    notFound();
  }

  const savings = bundle.original_price - bundle.bundle_price;

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted/30">
      {/* Hero */}
      <div className="bg-primary/5 py-12">
        <div className="container mx-auto px-4">
          <div className="grid lg:grid-cols-2 gap-8 items-center">
            <div>
              <div className="inline-flex items-center gap-2 bg-primary/10 text-primary px-3 py-1 rounded-full text-sm mb-4">
                <Package className="h-4 w-4" />
                Course Bundle
              </div>
              <h1 className="text-3xl lg:text-4xl font-bold mb-4">
                {bundle.title}
              </h1>
              <p className="text-lg text-muted-foreground mb-6">
                {bundle.description}
              </p>

              <div className="flex flex-wrap items-center gap-4 mb-6">
                <span className="text-sm text-muted-foreground">
                  {bundle.courses.length} courses included
                </span>
                <span className="text-sm text-muted-foreground">
                  {bundle.purchase_count} students enrolled
                </span>
              </div>

              {/* Pricing Card */}
              <Card className="max-w-sm">
                <CardContent className="pt-6">
                  <div className="flex items-baseline gap-3 mb-4">
                    <span className="text-4xl font-bold">
                      ${bundle.bundle_price.toFixed(2)}
                    </span>
                    <span className="text-xl text-muted-foreground line-through">
                      ${bundle.original_price.toFixed(2)}
                    </span>
                  </div>
                  <div className="bg-green-100 dark:bg-green-900/20 text-green-700 dark:text-green-300 px-3 py-2 rounded-lg text-sm font-medium mb-4">
                    🎉 You save ${savings.toFixed(2)} ({bundle.discount_percent}
                    % off)
                  </div>

                  {bundle.end_date && (
                    <p className="text-sm text-orange-600 flex items-center gap-2 mb-4">
                      <Clock className="h-4 w-4" />
                      Offer ends{" "}
                      {new Date(bundle.end_date).toLocaleDateString()}
                    </p>
                  )}

                  <Button className="w-full" size="lg">
                    Get This Bundle
                  </Button>
                  <p className="text-xs text-center text-muted-foreground mt-3">
                    30-day money-back guarantee
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* Bundle Preview */}
            <div className="hidden lg:block">
              <div className="relative">
                <div className="grid grid-cols-2 gap-4">
                  {bundle.courses.slice(0, 4).map((item, i) => (
                    <div
                      key={item.course.id}
                      className={`rounded-lg overflow-hidden shadow-lg ${
                        i === 0 ? "col-span-2" : ""
                      }`}
                    >
                      <div
                        className={`relative ${
                          i === 0 ? "aspect-video" : "aspect-[4/3]"
                        }`}
                      >
                        {item.course.thumbnail_url ? (
                          <Image
                            src={item.course.thumbnail_url}
                            alt={item.course.title}
                            fill
                            className="object-cover"
                          />
                        ) : (
                          <div className="absolute inset-0 bg-muted flex items-center justify-center">
                            <span className="text-4xl">📚</span>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* What's Included */}
      <div className="container mx-auto px-4 py-12">
        <h2 className="text-2xl font-bold mb-6">What's Included</h2>
        <div className="grid gap-4">
          {bundle.courses.map((item, index) => (
            <Card key={item.course.id}>
              <CardContent className="p-4">
                <div className="flex gap-4">
                  <div className="relative w-40 aspect-video rounded overflow-hidden flex-shrink-0">
                    {item.course.thumbnail_url ? (
                      <Image
                        src={item.course.thumbnail_url}
                        alt={item.course.title}
                        fill
                        className="object-cover"
                      />
                    ) : (
                      <div className="absolute inset-0 bg-muted flex items-center justify-center">
                        <span className="text-2xl">📚</span>
                      </div>
                    )}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-start justify-between">
                      <div>
                        <h3 className="font-semibold mb-1">
                          {item.course.title}
                        </h3>
                        <p className="text-sm text-muted-foreground mb-2">
                          By {item.course.instructor.name}
                        </p>
                      </div>
                      <span className="text-muted-foreground line-through">
                        ${item.course.price.toFixed(2)}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground line-clamp-2 mb-3">
                      {item.course.description}
                    </p>
                    <div className="flex items-center gap-4 text-sm">
                      <span className="flex items-center gap-1">
                        <Star className="h-4 w-4 text-yellow-400 fill-yellow-400" />
                        {item.course.rating.toFixed(1)}
                      </span>
                      <span className="flex items-center gap-1 text-muted-foreground">
                        <Users className="h-4 w-4" />
                        {item.course.students_count.toLocaleString()} students
                      </span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Benefits */}
        <div className="mt-12 bg-primary/5 rounded-2xl p-8">
          <h2 className="text-2xl font-bold mb-6 text-center">
            Why Buy This Bundle?
          </h2>
          <div className="grid md:grid-cols-3 gap-6">
            <div className="text-center">
              <div className="h-12 w-12 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                <span className="text-2xl">💰</span>
              </div>
              <h3 className="font-semibold mb-2">Save ${savings.toFixed(2)}</h3>
              <p className="text-sm text-muted-foreground">
                Get {bundle.courses.length} courses for the price of{" "}
                {Math.floor(
                  bundle.bundle_price /
                    (bundle.original_price / bundle.courses.length)
                )}
              </p>
            </div>
            <div className="text-center">
              <div className="h-12 w-12 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                <span className="text-2xl">♾️</span>
              </div>
              <h3 className="font-semibold mb-2">Lifetime Access</h3>
              <p className="text-sm text-muted-foreground">
                Once purchased, you have unlimited access to all courses forever
              </p>
            </div>
            <div className="text-center">
              <div className="h-12 w-12 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                <span className="text-2xl">📜</span>
              </div>
              <h3 className="font-semibold mb-2">Certificate of Completion</h3>
              <p className="text-sm text-muted-foreground">
                Earn a certificate for each course you complete
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
