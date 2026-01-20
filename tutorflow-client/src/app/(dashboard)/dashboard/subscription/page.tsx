import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch } from "@/lib/server-api";
import { Check, X, Calendar, CreditCard, AlertTriangle } from "lucide-react";
import Link from "next/link";

interface Subscription {
  id: string;
  status: "active" | "canceled" | "past_due" | "trialing" | "expired";
  interval: "monthly" | "yearly";
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
  plan: {
    id: string;
    name: string;
    slug: string;
    price_monthly: number;
    price_yearly: number;
    features: string[];
  };
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

// Server Component
export default async function SubscriptionPage() {
  const subscription = await authServerFetch<Subscription>("/subscriptions/my");

  const getStatusColor = (status: string) => {
    switch (status) {
      case "active":
      case "trialing":
        return "text-green-600 bg-green-100";
      case "past_due":
        return "text-yellow-600 bg-yellow-100";
      case "canceled":
      case "expired":
        return "text-red-600 bg-red-100";
      default:
        return "text-gray-600 bg-gray-100";
    }
  };

  if (!subscription) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold">Subscription</h2>
          <p className="text-muted-foreground">Manage your subscription plan</p>
        </div>

        <Card className="text-center py-12">
          <CardContent>
            <CreditCard className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
            <h3 className="text-xl font-semibold mb-2">
              No Active Subscription
            </h3>
            <p className="text-muted-foreground mb-6">
              Unlock premium features with a subscription plan.
            </p>
            <Button asChild>
              <Link href="/pricing">View Plans</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Subscription</h2>
        <p className="text-muted-foreground">Manage your subscription plan</p>
      </div>

      {/* Status Alert */}
      {subscription.cancel_at_period_end && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4 flex items-start gap-3">
          <AlertTriangle className="h-5 w-5 text-yellow-600 mt-0.5" />
          <div>
            <p className="font-medium text-yellow-800 dark:text-yellow-200">
              Your subscription is set to cancel
            </p>
            <p className="text-sm text-yellow-700 dark:text-yellow-300">
              You will lose access on{" "}
              {formatDate(subscription.current_period_end)}
            </p>
          </div>
        </div>
      )}

      {/* Current Plan */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{subscription.plan.name} Plan</CardTitle>
              <CardDescription>Your current subscription</CardDescription>
            </div>
            <span
              className={`px-3 py-1 rounded-full text-sm font-medium capitalize ${getStatusColor(
                subscription.status
              )}`}
            >
              {subscription.status}
            </span>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Pricing */}
          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-bold">
              $
              {subscription.interval === "yearly"
                ? subscription.plan.price_yearly
                : subscription.plan.price_monthly}
            </span>
            <span className="text-muted-foreground">
              /{subscription.interval === "yearly" ? "year" : "month"}
            </span>
          </div>

          {/* Billing Period */}
          <div className="grid sm:grid-cols-2 gap-4">
            <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
              <Calendar className="h-5 w-5 text-muted-foreground" />
              <div>
                <p className="text-sm text-muted-foreground">Current Period</p>
                <p className="font-medium">
                  {formatDate(subscription.current_period_start)} -{" "}
                  {formatDate(subscription.current_period_end)}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
              <CreditCard className="h-5 w-5 text-muted-foreground" />
              <div>
                <p className="text-sm text-muted-foreground">Next Payment</p>
                <p className="font-medium">
                  {subscription.cancel_at_period_end
                    ? "No upcoming payments"
                    : formatDate(subscription.current_period_end)}
                </p>
              </div>
            </div>
          </div>

          {/* Features */}
          <div>
            <h4 className="font-medium mb-3">Plan Features</h4>
            <ul className="grid sm:grid-cols-2 gap-2">
              {subscription.plan.features.map((feature, i) => (
                <li key={i} className="flex items-center gap-2 text-sm">
                  <Check className="h-4 w-4 text-green-600" />
                  {feature}
                </li>
              ))}
            </ul>
          </div>

          {/* Actions */}
          <div className="flex flex-wrap gap-3 pt-4 border-t">
            <Button variant="outline" asChild>
              <Link href="/pricing">Change Plan</Link>
            </Button>
            {subscription.cancel_at_period_end ? (
              <form action="/api/subscriptions/resume" method="POST">
                <Button type="submit">Resume Subscription</Button>
              </form>
            ) : (
              <Button variant="destructive" asChild>
                <Link href="/dashboard/subscription/cancel">
                  Cancel Subscription
                </Link>
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Payment History */}
      <Card>
        <CardHeader>
          <CardTitle>Payment History</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-8">
            Payment history will appear here after your first payment.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
