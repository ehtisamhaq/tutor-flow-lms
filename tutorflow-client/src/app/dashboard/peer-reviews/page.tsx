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
  ClipboardCheck,
  Clock,
  CheckCircle,
  AlertTriangle,
  FileText,
} from "lucide-react";
import Link from "next/link";

interface PeerReviewAssignment {
  id: string;
  submission_id: string;
  reviewer_id: string;
  status: "pending" | "assigned" | "completed" | "disputed";
  assigned_at: string;
  due_at: string;
  completed_at: string | null;
  submission?: {
    id: string;
    assignment?: {
      title: string;
    };
    user?: {
      name: string;
    };
  };
}

interface AssignmentsResponse {
  items: PeerReviewAssignment[];
  total: number;
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function isOverdue(dueAt: string): boolean {
  return new Date(dueAt) < new Date();
}

function getStatusColor(status: string, dueAt: string) {
  if (status === "completed") return "bg-green-100 text-green-700";
  if (status === "disputed") return "bg-red-100 text-red-700";
  if (isOverdue(dueAt)) return "bg-red-100 text-red-700";
  if (status === "assigned") return "bg-blue-100 text-blue-700";
  return "bg-yellow-100 text-yellow-700";
}

export default async function PeerReviewsPage() {
  const assignments = await authServerFetch<AssignmentsResponse>(
    "/peer-reviews/my-assignments"
  );

  const pendingReviews =
    assignments?.items?.filter(
      (a) => a.status === "assigned" || a.status === "pending"
    ) || [];

  const overdueReviews = pendingReviews.filter((a) => isOverdue(a.due_at));
  const completedReviews =
    assignments?.items?.filter((a) => a.status === "completed") || [];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">My Peer Reviews</h2>
        <p className="text-muted-foreground">
          Review and provide feedback on peer submissions
        </p>
      </div>

      {/* Stats */}
      <div className="grid sm:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <ClipboardCheck className="h-4 w-4" />
              Total Assigned
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{assignments?.total || 0}</div>
          </CardContent>
        </Card>
        <Card className={pendingReviews.length > 0 ? "border-blue-300" : ""}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Clock className="h-4 w-4" />
              To Review
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {pendingReviews.length}
            </div>
          </CardContent>
        </Card>
        <Card className={overdueReviews.length > 0 ? "border-red-300" : ""}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <AlertTriangle className="h-4 w-4" />
              Overdue
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {overdueReviews.length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <CheckCircle className="h-4 w-4" />
              Completed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {completedReviews.length}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Overdue Alert */}
      {overdueReviews.length > 0 && (
        <Card className="border-red-300 bg-red-50 dark:bg-red-900/20">
          <CardContent className="py-4">
            <div className="flex items-center gap-3">
              <AlertTriangle className="h-5 w-5 text-red-600" />
              <p className="font-medium text-red-800 dark:text-red-200">
                {overdueReviews.length} review
                {overdueReviews.length > 1 ? "s are" : " is"} overdue! Please
                complete them as soon as possible.
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Pending Reviews */}
      {pendingReviews.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Reviews to Complete</CardTitle>
            <CardDescription>
              Click on a review to start providing feedback
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {pendingReviews.map((assignment) => (
                <div
                  key={assignment.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                >
                  <div className="space-y-1">
                    <h4 className="font-medium">
                      {assignment.submission?.assignment?.title ||
                        "Assignment Submission"}
                    </h4>
                    <p className="text-sm text-muted-foreground">
                      Submitted by:{" "}
                      {assignment.submission?.user?.name || "Anonymous"}
                    </p>
                    <p className="text-sm">
                      <span className="font-medium">Due:</span>{" "}
                      <span
                        className={
                          isOverdue(assignment.due_at)
                            ? "text-red-600 font-medium"
                            : ""
                        }
                      >
                        {formatDate(assignment.due_at)}
                      </span>
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(
                        assignment.status,
                        assignment.due_at
                      )}`}
                    >
                      {isOverdue(assignment.due_at) &&
                      assignment.status !== "completed"
                        ? "Overdue"
                        : assignment.status}
                    </span>
                    <Button asChild>
                      <Link href={`/dashboard/peer-reviews/${assignment.id}`}>
                        <FileText className="h-4 w-4 mr-2" />
                        Review
                      </Link>
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Completed Reviews */}
      <Card>
        <CardHeader>
          <CardTitle>Completed Reviews</CardTitle>
        </CardHeader>
        <CardContent>
          {completedReviews.length > 0 ? (
            <div className="space-y-3">
              {completedReviews.map((assignment) => (
                <div
                  key={assignment.id}
                  className="flex items-center justify-between p-3 border rounded-lg bg-green-50/50 dark:bg-green-900/10"
                >
                  <div>
                    <h4 className="font-medium">
                      {assignment.submission?.assignment?.title ||
                        "Assignment Submission"}
                    </h4>
                    <p className="text-sm text-muted-foreground">
                      Completed:{" "}
                      {assignment.completed_at
                        ? formatDate(assignment.completed_at)
                        : "N/A"}
                    </p>
                  </div>
                  <CheckCircle className="h-5 w-5 text-green-600" />
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <ClipboardCheck className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No completed reviews yet</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
