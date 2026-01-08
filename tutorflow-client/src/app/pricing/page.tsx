import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { serverFetch } from "@/lib/server-api";
import { Check } from "lucide-react";
import Link from "next/link";

interface SubscriptionPlan {
  id: string;
  name: string;
  slug: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  features: string[];
  max_courses: number | null;
  offline_access: boolean;
  certificate_access: boolean;
  priority: number;
}

// Server Component
export default async function PricingPage() {
  const plans = await serverFetch<SubscriptionPlan[]>("/subscription-plans");

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted/30">
      {/* Header */}
      <div className="text-center py-16">
        <h1 className="text-4xl font-bold mb-4">Simple, Transparent Pricing</h1>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Choose the perfect plan for your learning journey. Upgrade or
          downgrade anytime.
        </p>
      </div>

      {/* Plans */}
      <div className="container mx-auto px-4 pb-16">
        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          {(
            plans || [
              {
                id: "1",
                name: "Free",
                slug: "free",
                description: "Perfect for getting started",
                price_monthly: 0,
                price_yearly: 0,
                features: [
                  "Access to free courses",
                  "Community support",
                  "Basic certificate",
                ],
                max_courses: 5,
                offline_access: false,
                certificate_access: true,
                priority: 1,
              },
              {
                id: "2",
                name: "Pro",
                slug: "pro",
                description: "For serious learners",
                price_monthly: 19.99,
                price_yearly: 199,
                features: [
                  "Unlimited courses",
                  "Priority support",
                  "Certificates",
                  "Progress tracking",
                  "Offline access",
                ],
                max_courses: null,
                offline_access: true,
                certificate_access: true,
                priority: 2,
              },
              {
                id: "3",
                name: "Team",
                slug: "team",
                description: "For organizations",
                price_monthly: 49.99,
                price_yearly: 499,
                features: [
                  "Everything in Pro",
                  "Team management",
                  "Analytics dashboard",
                  "Custom branding",
                  "API access",
                ],
                max_courses: null,
                offline_access: true,
                certificate_access: true,
                priority: 3,
              },
            ]
          ).map((plan, index) => (
            <Card
              key={plan.id}
              className={`relative ${
                index === 1 ? "border-primary shadow-lg scale-105 z-10" : ""
              }`}
            >
              {index === 1 && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-primary text-primary-foreground px-3 py-1 rounded-full text-xs font-semibold">
                  Most Popular
                </div>
              )}
              <CardHeader className="text-center">
                <CardTitle>{plan.name}</CardTitle>
                <CardDescription>{plan.description}</CardDescription>
              </CardHeader>
              <CardContent className="text-center">
                <div className="mb-6">
                  <span className="text-4xl font-bold">
                    ${plan.price_monthly}
                  </span>
                  <span className="text-muted-foreground">/month</span>
                </div>
                {plan.price_yearly > 0 && (
                  <p className="text-sm text-muted-foreground mb-6">
                    or ${plan.price_yearly}/year (save{" "}
                    {Math.round(
                      (1 - plan.price_yearly / (plan.price_monthly * 12)) * 100
                    )}
                    %)
                  </p>
                )}
                <ul className="space-y-3 text-left">
                  {plan.features.map((feature, i) => (
                    <li key={i} className="flex items-center gap-2">
                      <Check className="h-4 w-4 text-green-600" />
                      <span className="text-sm">{feature}</span>
                    </li>
                  ))}
                  {plan.max_courses !== null && (
                    <li className="flex items-center gap-2">
                      <Check className="h-4 w-4 text-green-600" />
                      <span className="text-sm">
                        Up to {plan.max_courses} courses
                      </span>
                    </li>
                  )}
                </ul>
              </CardContent>
              <CardFooter>
                <Button
                  className="w-full"
                  variant={index === 1 ? "default" : "outline"}
                  asChild
                >
                  <Link href={`/subscribe/${plan.slug}`}>
                    {plan.price_monthly === 0 ? "Get Started" : "Subscribe"}
                  </Link>
                </Button>
              </CardFooter>
            </Card>
          ))}
        </div>

        {/* FAQ */}
        <div className="mt-20 max-w-2xl mx-auto">
          <h2 className="text-2xl font-bold text-center mb-8">
            Frequently Asked Questions
          </h2>
          <div className="space-y-6">
            <div>
              <h3 className="font-semibold mb-2">Can I cancel anytime?</h3>
              <p className="text-muted-foreground text-sm">
                Yes! You can cancel your subscription at any time. You'll
                continue to have access until the end of your billing period.
              </p>
            </div>
            <div>
              <h3 className="font-semibold mb-2">
                What payment methods do you accept?
              </h3>
              <p className="text-muted-foreground text-sm">
                We accept all major credit cards, debit cards, and PayPal
                through our secure payment processor Stripe.
              </p>
            </div>
            <div>
              <h3 className="font-semibold mb-2">Is there a free trial?</h3>
              <p className="text-muted-foreground text-sm">
                Yes! New Pro subscribers get a 7-day free trial to explore all
                features before being charged.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
