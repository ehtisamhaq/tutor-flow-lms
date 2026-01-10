"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Check, Loader2 } from "lucide-react";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

interface Plan {
  id: string;
  name: string;
  slug: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  features: string[];
}

export default function SubscribePage() {
  const { slug } = useParams();
  const router = useRouter();
  const { user } = useAuthStore();
  const [plan, setPlan] = useState<Plan | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [interval, setInterval] = useState<"monthly" | "yearly">("monthly");

  useEffect(() => {
    async function fetchPlan() {
      try {
        const response = await api.get<Plan>(`/subscription-plans/${slug}`);
        setPlan(response.data);
      } catch (error) {
        toast.error("Failed to load plan details");
        router.push("/pricing");
      } finally {
        setLoading(false);
      }
    }

    if (slug) {
      fetchPlan();
    }
  }, [slug, router]);

  const handleSubscribe = async () => {
    if (!user) {
      toast.error("Please login to subscribe");
      router.push(`/login?redirect=/subscribe/${slug}`);
      return;
    }

    try {
      setSubmitting(true);
      // Using /subscriptions/subscribe but it should return a checkout_url
      // mirroring the order flow for paid plans.
      const response = await api.post<any>("/subscriptions/subscribe", {
        plan_slug: slug,
        interval: interval,
      });

      if (response.data.checkout_url) {
        window.location.href = response.data.checkout_url;
      } else {
        toast.success("Subscribed successfully!");
        router.push("/dashboard");
      }
    } catch (error: any) {
      toast.error(
        error.response?.data?.error?.message ||
          "Failed to initiate subscription"
      );
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!plan) return null;

  const price = interval === "monthly" ? plan.price_monthly : plan.price_yearly;

  return (
    <div className="container max-w-4xl py-12">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4">Complete your subscription</h1>
        <p className="text-muted-foreground">
          You're one step away from unlocking premium features
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <Card className="flex flex-col border-primary/50 bg-primary/5 shadow-lg">
          <CardHeader>
            <CardTitle>{plan.name}</CardTitle>
            <CardDescription>{plan.description}</CardDescription>
          </CardHeader>
          <CardContent className="flex-grow">
            <div className="mb-6">
              <span className="text-4xl font-bold">${price}</span>
              <span className="text-muted-foreground ml-2">/{interval}</span>
            </div>
            <ul className="space-y-3">
              {plan.features?.map((feature, index) => (
                <li key={index} className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-primary" />
                  <span>{feature}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card className="flex flex-col">
          <CardHeader>
            <CardTitle>Select Interval</CardTitle>
            <CardDescription>Choose how you want to be billed</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6 flex-grow">
            <div className="grid grid-cols-2 gap-4">
              <button
                onClick={() => setInterval("monthly")}
                className={`p-4 rounded-lg border-2 text-center transition-all ${
                  interval === "monthly"
                    ? "border-primary bg-primary/5 shadow-md"
                    : "border-muted hover:border-primary/50"
                }`}
              >
                <div className="font-semibold">Monthly</div>
                <div className="text-sm text-muted-foreground">
                  ${plan.price_monthly}/mo
                </div>
              </button>
              <button
                onClick={() => setInterval("yearly")}
                className={`p-4 rounded-lg border-2 text-center transition-all ${
                  interval === "yearly"
                    ? "border-primary bg-primary/5 shadow-md"
                    : "border-muted hover:border-primary/50"
                }`}
              >
                <div className="font-semibold">Yearly</div>
                <div className="text-sm text-muted-foreground">
                  ${plan.price_yearly}/yr
                </div>
              </button>
            </div>

            <div className="pt-6 border-t font-medium">
              <div className="flex justify-between mb-2">
                <span>Subtotal</span>
                <span>${price}</span>
              </div>
              <div className="flex justify-between text-lg font-bold">
                <span>Total Due Now</span>
                <span>${price}</span>
              </div>
            </div>
          </CardContent>
          <CardFooter>
            <Button
              className="w-full h-12 text-lg"
              onClick={handleSubscribe}
              disabled={submitting}
            >
              {submitting ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin mr-2" />
                  Initiating...
                </>
              ) : (
                "Proceed to Payment"
              )}
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
