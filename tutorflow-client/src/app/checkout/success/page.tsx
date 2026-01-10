"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CheckCircle2, Loader2, ArrowRight } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

export default function CheckoutSuccessPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const sessionId = searchParams.get("session_id");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // In a real app, we might verify the session with the server here
    // But for this project, Stripe webhooks handle the heavy lifting.
    const timer = setTimeout(() => {
      setLoading(false);
      toast.success("Payment successful!");
    }, 2000);

    return () => clearTimeout(timer);
  }, [sessionId]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="w-12 h-12 animate-spin text-primary" />
        <p className="text-xl font-medium">Confirming your payment...</p>
      </div>
    );
  }

  return (
    <div className="container max-w-2xl py-20">
      <Card className="text-center">
        <CardHeader className="pt-10">
          <div className="flex justify-center mb-6">
            <CheckCircle2 className="w-20 h-20 text-green-500" />
          </div>
          <CardTitle className="text-3xl font-bold">
            Payment Successful!
          </CardTitle>
          <p className="text-muted-foreground mt-2">
            Thank you for your purchase. Your enrollment has been processed.
          </p>
        </CardHeader>
        <CardContent className="space-y-4 py-6">
          <p>
            You can now access your courses and start learning right away. A
            confirmation email has been sent to your registered email address.
          </p>
          {sessionId && (
            <p className="text-xs text-muted-foreground">
              Reference ID: {sessionId}
            </p>
          )}
        </CardContent>
        <CardFooter className="flex flex-col sm:flex-row gap-4 justify-center pb-10">
          <Button asChild size="lg" className="px-8">
            <Link href="/dashboard">Go to Dashboard</Link>
          </Button>
          <Button asChild variant="outline" size="lg" className="px-8">
            <Link href="/learn">
              Browse My Courses
              <ArrowRight className="w-4 h-4 ml-2" />
            </Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
