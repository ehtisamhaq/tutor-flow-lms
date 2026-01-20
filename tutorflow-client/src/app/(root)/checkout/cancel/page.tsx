"use client";

import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { XCircle, ArrowLeft } from "lucide-react";
import Link from "next/link";

export default function CheckoutCancelPage() {
  const router = useRouter();

  return (
    <div className="container max-w-2xl py-20">
      <Card className="text-center">
        <CardHeader className="pt-10">
          <div className="flex justify-center mb-6">
            <XCircle className="w-20 h-20 text-red-500" />
          </div>
          <CardTitle className="text-3xl font-bold">
            Payment Cancelled
          </CardTitle>
          <p className="text-muted-foreground mt-2">
            Your payment process was cancelled and no charges were made.
          </p>
        </CardHeader>
        <CardContent className="space-y-4 py-6">
          <p>
            If you encountered any issues during checkout, please try again or
            contact our support team for assistance. Items in your cart have
            been saved so you can complete your purchase later.
          </p>
        </CardContent>
        <CardFooter className="flex flex-col sm:flex-row gap-4 justify-center pb-10">
          <Button asChild size="lg" className="px-8">
            <Link href="/cart">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Return to Cart
            </Link>
          </Button>
          <Button asChild variant="outline" size="lg" className="px-8">
            <Link href="/pricing">View Plans</Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
