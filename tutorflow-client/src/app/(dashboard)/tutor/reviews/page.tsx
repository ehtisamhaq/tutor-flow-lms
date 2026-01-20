import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch } from "@/lib/server-api";
import {
  Star,
  MessageSquare,
  ThumbsUp,
  ThumbsDown,
  MoreVertical,
} from "lucide-react";
import Link from "next/link";

interface Review {
  id: string;
  rating: number;
  comment: string;
  created_at: string;
  response?: string;
  responded_at?: string;
  user: {
    id: string;
    name: string;
    avatar_url?: string;
  };
  course: {
    id: string;
    title: string;
    slug: string;
  };
}

interface ReviewsData {
  reviews: Review[];
  stats: {
    total: number;
    average: number;
    distribution: number[];
  };
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function StarRating({ rating }: { rating: number }) {
  return (
    <div className="flex gap-0.5">
      {[1, 2, 3, 4, 5].map((star) => (
        <Star
          key={star}
          className={`h-4 w-4 ${
            star <= rating ? "text-yellow-400 fill-yellow-400" : "text-gray-300"
          }`}
        />
      ))}
    </div>
  );
}

// Server Component
export default async function TutorReviewsPage() {
  const data = await authServerFetch<ReviewsData>("/tutors/reviews");

  const stats = data?.stats || {
    total: 0,
    average: 0,
    distribution: [0, 0, 0, 0, 0],
  };
  const reviews = data?.reviews || [];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Reviews & Ratings</h2>
        <p className="text-muted-foreground">
          See what students are saying about your courses
        </p>
      </div>

      {/* Stats Overview */}
      <div className="grid md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-6 text-center">
            <div className="text-4xl font-bold mb-1">
              {stats.average.toFixed(1)}
            </div>
            <div className="flex justify-center mb-2">
              <StarRating rating={Math.round(stats.average)} />
            </div>
            <p className="text-sm text-muted-foreground">
              {stats.total} reviews
            </p>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardContent className="pt-6">
            <div className="space-y-2">
              {[5, 4, 3, 2, 1].map((stars) => {
                const count = stats.distribution[stars - 1] || 0;
                const percent =
                  stats.total > 0 ? (count / stats.total) * 100 : 0;
                return (
                  <div key={stars} className="flex items-center gap-2">
                    <span className="text-sm w-4">{stars}</span>
                    <Star className="h-4 w-4 text-yellow-400 fill-yellow-400" />
                    <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full bg-yellow-400 rounded-full"
                        style={{ width: `${percent}%` }}
                      />
                    </div>
                    <span className="text-sm text-muted-foreground w-10">
                      {count}
                    </span>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Reviews List */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>All Reviews</CardTitle>
            <div className="flex gap-2">
              <Button variant="outline" size="sm">
                Newest First
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {reviews.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <MessageSquare className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No reviews yet</p>
              <p className="text-sm">
                Reviews will appear here when students rate your courses
              </p>
            </div>
          ) : (
            <div className="space-y-6">
              {reviews.map((review) => (
                <div
                  key={review.id}
                  className="border-b pb-6 last:border-0 last:pb-0"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                        {review.user.avatar_url ? (
                          <img
                            src={review.user.avatar_url}
                            alt={review.user.name}
                            className="h-10 w-10 rounded-full object-cover"
                          />
                        ) : (
                          <span className="text-primary font-medium">
                            {review.user.name.charAt(0).toUpperCase()}
                          </span>
                        )}
                      </div>
                      <div>
                        <p className="font-medium">{review.user.name}</p>
                        <p className="text-sm text-muted-foreground">
                          {formatDate(review.created_at)} •{" "}
                          <Link
                            href={`/courses/${review.course.slug}`}
                            className="hover:underline"
                          >
                            {review.course.title}
                          </Link>
                        </p>
                      </div>
                    </div>
                    <StarRating rating={review.rating} />
                  </div>

                  <p className="text-muted-foreground mb-4">{review.comment}</p>

                  {review.response ? (
                    <div className="bg-muted/50 rounded-lg p-4 ml-6">
                      <p className="text-sm font-medium mb-1">Your response</p>
                      <p className="text-sm text-muted-foreground">
                        {review.response}
                      </p>
                      <p className="text-xs text-muted-foreground mt-2">
                        Responded on {formatDate(review.responded_at!)}
                      </p>
                    </div>
                  ) : (
                    <Button variant="outline" size="sm">
                      <MessageSquare className="mr-2 h-4 w-4" />
                      Reply
                    </Button>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
