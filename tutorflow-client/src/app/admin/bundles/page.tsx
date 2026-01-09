import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch } from "@/lib/server-api";
import { Plus, Edit, Trash2, Package, BookOpen, Percent } from "lucide-react";
import Link from "next/link";

interface Bundle {
  id: string;
  title: string;
  slug: string;
  description: string;
  thumbnail_url: string;
  original_price: number;
  bundle_price: number;
  discount_percent: number;
  is_active: boolean;
  purchase_count: number;
  courses: {
    id: string;
    course: {
      id: string;
      title: string;
      thumbnail_url: string;
    };
  }[];
  created_at: string;
}

interface BundlesResponse {
  items: Bundle[];
  total: number;
}

export default async function AdminBundlesPage() {
  const bundles = await authServerFetch<BundlesResponse>("/admin/bundles");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Bundle Management</h2>
          <p className="text-muted-foreground">
            Create and manage course bundles
          </p>
        </div>
        <Button asChild>
          <Link href="/admin/bundles/new">
            <Plus className="h-4 w-4 mr-2" />
            Create Bundle
          </Link>
        </Button>
      </div>

      {/* Stats */}
      <div className="grid sm:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Bundles
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{bundles?.total || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Bundles
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {bundles?.items?.filter((b) => b.is_active).length || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Sales
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {bundles?.items?.reduce(
                (sum, b) => sum + (b.purchase_count || 0),
                0
              ) || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Revenue
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              $
              {bundles?.items
                ?.reduce((sum, b) => sum + b.bundle_price * b.purchase_count, 0)
                .toFixed(2) || "0.00"}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Bundles List */}
      <div className="grid md:grid-cols-2 gap-6">
        {bundles?.items?.map((bundle) => (
          <Card
            key={bundle.id}
            className={!bundle.is_active ? "opacity-60" : ""}
          >
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Package className="h-5 w-5" />
                  {bundle.title}
                </CardTitle>
                <span
                  className={`px-2 py-1 rounded-full text-xs font-medium ${
                    bundle.is_active
                      ? "bg-green-100 text-green-700"
                      : "bg-gray-100 text-gray-600"
                  }`}
                >
                  {bundle.is_active ? "Active" : "Inactive"}
                </span>
              </div>
              <CardDescription>{bundle.description}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Pricing */}
              <div className="flex items-center gap-3">
                <span className="text-2xl font-bold">
                  ${bundle.bundle_price}
                </span>
                <span className="text-muted-foreground line-through">
                  ${bundle.original_price}
                </span>
                <span className="px-2 py-1 bg-primary/10 text-primary rounded-full text-sm font-medium flex items-center">
                  <Percent className="h-3 w-3 mr-1" />
                  {bundle.discount_percent}% off
                </span>
              </div>

              {/* Courses */}
              <div>
                <p className="text-sm font-medium mb-2 flex items-center gap-2">
                  <BookOpen className="h-4 w-4" />
                  {bundle.courses?.length || 0} Courses
                </p>
                <div className="flex flex-wrap gap-2">
                  {bundle.courses?.slice(0, 3).map((c) => (
                    <span
                      key={c.id}
                      className="px-2 py-1 bg-muted rounded text-xs"
                    >
                      {c.course?.title}
                    </span>
                  ))}
                  {(bundle.courses?.length || 0) > 3 && (
                    <span className="px-2 py-1 bg-muted rounded text-xs text-muted-foreground">
                      +{bundle.courses.length - 3} more
                    </span>
                  )}
                </div>
              </div>

              <div className="text-sm text-muted-foreground">
                {bundle.purchase_count} purchase
                {bundle.purchase_count !== 1 ? "s" : ""}
              </div>

              <div className="flex gap-2 pt-2 border-t">
                <Button variant="outline" size="sm" asChild className="flex-1">
                  <Link href={`/admin/bundles/${bundle.id}/edit`}>
                    <Edit className="h-4 w-4 mr-1" />
                    Edit
                  </Link>
                </Button>
                <Button variant="ghost" size="sm" className="text-destructive">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </CardContent>
          </Card>
        )) || (
          <Card className="col-span-full">
            <CardContent className="text-center py-12">
              <Package className="h-12 w-12 mx-auto mb-4 text-muted-foreground opacity-50" />
              <p className="text-muted-foreground">No bundles created yet</p>
              <Button className="mt-4" asChild>
                <Link href="/admin/bundles/new">Create your first bundle</Link>
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
