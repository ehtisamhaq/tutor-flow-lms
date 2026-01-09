import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch } from "@/lib/server-api";
import { Plus, Edit, Trash2, CreditCard, Check } from "lucide-react";
import Link from "next/link";

interface SubscriptionPlan {
  id: string;
  name: string;
  slug: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  features: string[];
  is_active: boolean;
  subscriber_count: number;
  created_at: string;
}

interface PlansResponse {
  items: SubscriptionPlan[];
  total: number;
}

export default async function SubscriptionPlansPage() {
  const plans = await authServerFetch<PlansResponse>(
    "/admin/subscription-plans"
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Subscription Plans</h2>
          <p className="text-muted-foreground">
            Manage your subscription pricing tiers
          </p>
        </div>
        <Button asChild>
          <Link href="/admin/subscription-plans/new">
            <Plus className="h-4 w-4 mr-2" />
            Create Plan
          </Link>
        </Button>
      </div>

      {/* Stats */}
      <div className="grid sm:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Plans
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{plans?.total || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Plans
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {plans?.items?.filter((p) => p.is_active).length || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Subscribers
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {plans?.items?.reduce(
                (sum, p) => sum + (p.subscriber_count || 0),
                0
              ) || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Plans List */}
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
        {plans?.items?.map((plan) => (
          <Card key={plan.id} className={!plan.is_active ? "opacity-60" : ""}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <CreditCard className="h-5 w-5" />
                  {plan.name}
                </CardTitle>
                <span
                  className={`px-2 py-1 rounded-full text-xs font-medium ${
                    plan.is_active
                      ? "bg-green-100 text-green-700"
                      : "bg-gray-100 text-gray-600"
                  }`}
                >
                  {plan.is_active ? "Active" : "Inactive"}
                </span>
              </div>
              <CardDescription>{plan.description}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-baseline gap-2">
                <span className="text-3xl font-bold">
                  ${plan.price_monthly}
                </span>
                <span className="text-muted-foreground">/mo</span>
                <span className="text-sm text-muted-foreground ml-2">
                  (${plan.price_yearly}/yr)
                </span>
              </div>

              <ul className="space-y-2">
                {plan.features?.slice(0, 4).map((feature, i) => (
                  <li key={i} className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-green-600" />
                    {feature}
                  </li>
                ))}
                {(plan.features?.length || 0) > 4 && (
                  <li className="text-sm text-muted-foreground">
                    +{plan.features.length - 4} more features
                  </li>
                )}
              </ul>

              <div className="text-sm text-muted-foreground">
                {plan.subscriber_count} subscriber
                {plan.subscriber_count !== 1 ? "s" : ""}
              </div>

              <div className="flex gap-2 pt-2 border-t">
                <Button variant="outline" size="sm" asChild className="flex-1">
                  <Link href={`/admin/subscription-plans/${plan.id}/edit`}>
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
              <CreditCard className="h-12 w-12 mx-auto mb-4 text-muted-foreground opacity-50" />
              <p className="text-muted-foreground">No subscription plans yet</p>
              <Button className="mt-4" asChild>
                <Link href="/admin/subscription-plans/new">
                  Create your first plan
                </Link>
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
