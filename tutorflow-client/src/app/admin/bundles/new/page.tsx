"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { api } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Plus, Trash2, ArrowLeft, Search } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";

interface Course {
  id: string;
  title: string;
  price: number;
  thumbnail_url: string;
}

export default function NewBundlePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Course[]>([]);
  const [selectedCourses, setSelectedCourses] = useState<Course[]>([]);
  const [formData, setFormData] = useState({
    title: "",
    slug: "",
    description: "",
    discount_percent: "20",
    is_active: true,
  });

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;

    try {
      const response = await api.get(
        `/search?q=${encodeURIComponent(searchQuery)}`
      );
      // Go backend SearchResult has 'courses' field
      setSearchResults(response.data.data.courses || []);
    } catch (error) {
      console.error("Search failed:", error);
    }
  };

  const handleAddCourse = (course: Course) => {
    if (!selectedCourses.find((c) => c.id === course.id)) {
      setSelectedCourses([...selectedCourses, course]);
    }
    setSearchResults([]);
    setSearchQuery("");
  };

  const handleRemoveCourse = (courseId: string) => {
    setSelectedCourses(selectedCourses.filter((c) => c.id !== courseId));
  };

  const calculatePricing = () => {
    const original = selectedCourses.reduce((sum, c) => sum + c.price, 0);
    const discount = parseFloat(formData.discount_percent) || 0;
    const bundle = original * (1 - discount / 100);
    return { original, bundle, savings: original - bundle };
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (selectedCourses.length === 0) {
      toast.error("Please add at least one course to the bundle");
      return;
    }

    setLoading(true);

    try {
      await api.post("/admin/bundles", {
        ...formData,
        discount_percent: parseFloat(formData.discount_percent),
        course_ids: selectedCourses.map((c) => c.id),
      });

      toast.success("Course bundle created successfully");
      router.push("/admin/bundles");
      router.refresh();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error?.message || "Failed to create bundle"
      );
    } finally {
      setLoading(false);
    }
  };

  const pricing = calculatePricing();

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/admin/bundles">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Link>
        </Button>
        <div>
          <h2 className="text-2xl font-bold">Create Course Bundle</h2>
          <p className="text-muted-foreground">
            Combine courses at a discounted price
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="grid lg:grid-cols-3 gap-6">
          {/* Bundle Info */}
          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle>Bundle Details</CardTitle>
              <CardDescription>
                Basic information about the bundle
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="title">Bundle Title</Label>
                <Input
                  id="title"
                  placeholder="e.g., Complete Web Development Bundle"
                  value={formData.title}
                  onChange={(e) =>
                    setFormData({ ...formData, title: e.target.value })
                  }
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="slug">Slug</Label>
                <Input
                  id="slug"
                  placeholder="e.g., web-dev-bundle"
                  value={formData.slug}
                  onChange={(e) =>
                    setFormData({ ...formData, slug: e.target.value })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  placeholder="Describe what this bundle offers..."
                  value={formData.description}
                  onChange={(e) =>
                    setFormData({ ...formData, description: e.target.value })
                  }
                  rows={4}
                />
              </div>

              <div className="grid sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="discount_percent">Discount (%)</Label>
                  <Input
                    id="discount_percent"
                    type="number"
                    min="0"
                    max="100"
                    value={formData.discount_percent}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        discount_percent: e.target.value,
                      })
                    }
                  />
                </div>
                <div className="flex items-center justify-between pt-8">
                  <Label htmlFor="is_active">Active</Label>
                  <Switch
                    id="is_active"
                    checked={formData.is_active}
                    onCheckedChange={(checked) =>
                      setFormData({ ...formData, is_active: checked })
                    }
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Pricing Preview */}
          <Card>
            <CardHeader>
              <CardTitle>Pricing Preview</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">Original Price</p>
                <p className="text-xl line-through">
                  ${pricing.original.toFixed(2)}
                </p>
              </div>
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">Bundle Price</p>
                <p className="text-3xl font-bold text-primary">
                  ${pricing.bundle.toFixed(2)}
                </p>
              </div>
              <div className="space-y-2 pt-2 border-t">
                <p className="text-sm text-muted-foreground">Customer Saves</p>
                <p className="text-lg font-medium text-green-600">
                  ${pricing.savings.toFixed(2)} ({formData.discount_percent}%)
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Course Selection */}
          <Card className="lg:col-span-3">
            <CardHeader>
              <CardTitle>Select Courses</CardTitle>
              <CardDescription>
                Search and add courses to this bundle
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Search */}
              <div className="flex gap-2">
                <Input
                  placeholder="Search courses..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyDown={(e) =>
                    e.key === "Enter" && (e.preventDefault(), handleSearch())
                  }
                />
                <Button type="button" variant="outline" onClick={handleSearch}>
                  <Search className="h-4 w-4" />
                </Button>
              </div>

              {/* Search Results */}
              {searchResults.length > 0 && (
                <div className="border rounded-lg p-4 space-y-2">
                  <p className="text-sm font-medium">Search Results</p>
                  {searchResults.map((course) => (
                    <div
                      key={course.id}
                      className="flex items-center justify-between p-2 hover:bg-muted rounded cursor-pointer"
                      onClick={() => handleAddCourse(course)}
                    >
                      <span>{course.title}</span>
                      <span className="text-muted-foreground">
                        ${course.price}
                      </span>
                    </div>
                  ))}
                </div>
              )}

              {/* Selected Courses */}
              <div>
                <p className="text-sm font-medium mb-3">
                  Selected Courses ({selectedCourses.length})
                </p>
                {selectedCourses.length === 0 ? (
                  <p className="text-muted-foreground text-center py-8">
                    No courses selected. Search and add courses above.
                  </p>
                ) : (
                  <div className="space-y-2">
                    {selectedCourses.map((course) => (
                      <div
                        key={course.id}
                        className="flex items-center justify-between p-3 border rounded-lg"
                      >
                        <span className="font-medium">{course.title}</span>
                        <div className="flex items-center gap-3">
                          <span className="text-muted-foreground">
                            ${course.price}
                          </span>
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRemoveCourse(course.id)}
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="flex justify-end gap-3 mt-6">
          <Button type="button" variant="outline" asChild>
            <Link href="/admin/bundles">Cancel</Link>
          </Button>
          <Button
            type="submit"
            disabled={loading || selectedCourses.length === 0}
          >
            {loading ? "Creating..." : "Create Bundle"}
          </Button>
        </div>
      </form>
    </div>
  );
}
