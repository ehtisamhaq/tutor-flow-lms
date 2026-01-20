import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { serverFetch } from "@/lib/server-api";
import { Package, Clock, Tag } from "lucide-react";
import Link from "next/link";
import Image from "next/image";

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
    course: {
      id: string;
      title: string;
      thumbnail_url?: string;
    };
  }[];
  end_date?: string;
  purchase_count: number;
}

interface BundlesResponse {
  items: Bundle[];
  total: number;
}

// Server Component
export default async function BundlesPage() {
  const data = await serverFetch<BundlesResponse>("/bundles?limit=20");
  const bundles = data?.items || [];

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted/30">
      {/* Header */}
      <div className="bg-primary/5 py-12">
        <div className="container mx-auto px-4 text-center">
          <Package className="h-16 w-16 mx-auto mb-4 text-primary" />
          <h1 className="text-3xl font-bold mb-2">Course Bundles</h1>
          <p className="text-muted-foreground max-w-xl mx-auto">
            Save big with our curated course bundles. Get multiple courses at a
            discounted price.
          </p>
        </div>
      </div>

      {/* Bundles Grid */}
      <div className="container mx-auto px-4 py-12">
        {bundles.length === 0 ? (
          <div className="text-center py-16">
            <Package className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
            <h2 className="text-xl font-semibold mb-2">No Bundles Available</h2>
            <p className="text-muted-foreground mb-6">
              Check back soon for exciting course bundles!
            </p>
            <Button asChild>
              <Link href="/courses">Browse Individual Courses</Link>
            </Button>
          </div>
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {bundles.map((bundle) => (
              <Card
                key={bundle.id}
                className="overflow-hidden group hover:shadow-lg transition-shadow"
              >
                <div className="relative aspect-video bg-muted">
                  {bundle.thumbnail_url ? (
                    <Image
                      src={bundle.thumbnail_url}
                      alt={bundle.title}
                      fill
                      className="object-cover"
                    />
                  ) : (
                    <div className="absolute inset-0 flex items-center justify-center">
                      <Package className="h-12 w-12 text-muted-foreground" />
                    </div>
                  )}
                  {bundle.discount_percent > 0 && (
                    <div className="absolute top-3 right-3 bg-red-500 text-white px-2 py-1 rounded-full text-sm font-bold">
                      {bundle.discount_percent}% OFF
                    </div>
                  )}
                </div>
                <CardHeader>
                  <CardTitle className="group-hover:text-primary transition-colors">
                    {bundle.title}
                  </CardTitle>
                  <p className="text-sm text-muted-foreground line-clamp-2">
                    {bundle.description}
                  </p>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Course Thumbnails */}
                  <div className="flex -space-x-3">
                    {bundle.courses.slice(0, 5).map((item, i) => (
                      <div
                        key={i}
                        className="h-10 w-10 rounded-full border-2 border-background bg-muted flex items-center justify-center overflow-hidden"
                      >
                        {item.course.thumbnail_url ? (
                          <Image
                            src={item.course.thumbnail_url}
                            alt={item.course.title}
                            width={40}
                            height={40}
                            className="object-cover"
                          />
                        ) : (
                          <span className="text-xs">📚</span>
                        )}
                      </div>
                    ))}
                    {bundle.courses.length > 5 && (
                      <div className="h-10 w-10 rounded-full border-2 border-background bg-primary text-primary-foreground flex items-center justify-center text-xs font-medium">
                        +{bundle.courses.length - 5}
                      </div>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {bundle.courses.length} courses included
                  </p>

                  {/* Pricing */}
                  <div className="flex items-baseline gap-2">
                    <span className="text-2xl font-bold">
                      ${bundle.bundle_price.toFixed(2)}
                    </span>
                    <span className="text-muted-foreground line-through">
                      ${bundle.original_price.toFixed(2)}
                    </span>
                    <span className="text-green-600 text-sm font-medium">
                      Save $
                      {(bundle.original_price - bundle.bundle_price).toFixed(2)}
                    </span>
                  </div>

                  {/* Deadline */}
                  {bundle.end_date && (
                    <div className="flex items-center gap-2 text-sm text-orange-600">
                      <Clock className="h-4 w-4" />
                      <span>
                        Ends {new Date(bundle.end_date).toLocaleDateString()}
                      </span>
                    </div>
                  )}
                </CardContent>
                <CardFooter>
                  <Button className="w-full" asChild>
                    <Link href={`/bundles/${bundle.slug}`}>View Bundle</Link>
                  </Button>
                </CardFooter>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
