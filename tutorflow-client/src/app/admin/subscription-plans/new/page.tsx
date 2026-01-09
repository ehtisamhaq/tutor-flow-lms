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
import { Plus, Trash2, ArrowLeft } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";

export default function NewSubscriptionPlanPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [features, setFeatures] = useState<string[]>([""]);
  const [formData, setFormData] = useState({
    name: "",
    slug: "",
    description: "",
    price_monthly: "",
    price_yearly: "",
    is_active: true,
  });

  const handleAddFeature = () => {
    setFeatures([...features, ""]);
  };

  const handleRemoveFeature = (index: number) => {
    setFeatures(features.filter((_, i) => i !== index));
  };

  const handleFeatureChange = (index: number, value: string) => {
    const newFeatures = [...features];
    newFeatures[index] = value;
    setFeatures(newFeatures);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      await api.post("/admin/subscription-plans", {
        ...formData,
        price_monthly: parseFloat(formData.price_monthly),
        price_yearly: parseFloat(formData.price_yearly),
        features: features.filter((f) => f.trim() !== ""),
      });

      toast.success("Subscription plan created successfully");
      router.push("/admin/subscription-plans");
      router.refresh();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error?.message || "Failed to create plan"
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/admin/subscription-plans">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Link>
        </Button>
        <div>
          <h2 className="text-2xl font-bold">Create Subscription Plan</h2>
          <p className="text-muted-foreground">
            Define a new subscription tier
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="grid lg:grid-cols-2 gap-6">
          {/* Basic Info */}
          <Card>
            <CardHeader>
              <CardTitle>Plan Details</CardTitle>
              <CardDescription>
                Basic information about the plan
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Plan Name</Label>
                <Input
                  id="name"
                  placeholder="e.g., Pro, Enterprise"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="slug">Slug</Label>
                <Input
                  id="slug"
                  placeholder="e.g., pro, enterprise"
                  value={formData.slug}
                  onChange={(e) =>
                    setFormData({ ...formData, slug: e.target.value })
                  }
                />
                <p className="text-xs text-muted-foreground">
                  Leave empty to auto-generate from name
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  placeholder="Describe what this plan offers..."
                  value={formData.description}
                  onChange={(e) =>
                    setFormData({ ...formData, description: e.target.value })
                  }
                  rows={3}
                />
              </div>

              <div className="flex items-center justify-between">
                <Label htmlFor="is_active">Active</Label>
                <Switch
                  id="is_active"
                  checked={formData.is_active}
                  onCheckedChange={(checked) =>
                    setFormData({ ...formData, is_active: checked })
                  }
                />
              </div>
            </CardContent>
          </Card>

          {/* Pricing */}
          <Card>
            <CardHeader>
              <CardTitle>Pricing</CardTitle>
              <CardDescription>Set the subscription prices</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="price_monthly">Monthly Price ($)</Label>
                <Input
                  id="price_monthly"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="9.99"
                  value={formData.price_monthly}
                  onChange={(e) =>
                    setFormData({ ...formData, price_monthly: e.target.value })
                  }
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="price_yearly">Yearly Price ($)</Label>
                <Input
                  id="price_yearly"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="99.99"
                  value={formData.price_yearly}
                  onChange={(e) =>
                    setFormData({ ...formData, price_yearly: e.target.value })
                  }
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Typically 10-20% discount vs monthly
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Features */}
          <Card className="lg:col-span-2">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Features</CardTitle>
                  <CardDescription>
                    List what's included in this plan
                  </CardDescription>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAddFeature}
                >
                  <Plus className="h-4 w-4 mr-2" />
                  Add Feature
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {features.map((feature, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      placeholder={`Feature ${index + 1}`}
                      value={feature}
                      onChange={(e) =>
                        handleFeatureChange(index, e.target.value)
                      }
                    />
                    {features.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRemoveFeature(index)}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="flex justify-end gap-3 mt-6">
          <Button type="button" variant="outline" asChild>
            <Link href="/admin/subscription-plans">Cancel</Link>
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? "Creating..." : "Create Plan"}
          </Button>
        </div>
      </form>
    </div>
  );
}
