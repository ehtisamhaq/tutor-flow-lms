"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  CheckCircle2,
  Loader2,
  ArrowRight,
  PlayCircle,
  ReceiptText,
  Calendar,
} from "lucide-react";
import Link from "next/link";
import Image from "next/image";
import { toast } from "sonner";
import confetti from "canvas-confetti";
import api from "@/lib/api";

interface OrderItem {
  id: string;
  course: {
    title: string;
    thumbnail_url?: string;
    instructor?: {
      first_name: string;
      last_name: string;
    };
  };
  price: number;
}

interface Order {
  id: string;
  order_number: string;
  total: number;
  created_at: string;
  items: OrderItem[];
}

function SuccessContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const sessionId = searchParams.get("session_id");
  const [loading, setLoading] = useState(true);
  const [order, setOrder] = useState<Order | null>(null);

  useEffect(() => {
    if (!sessionId) {
      setLoading(false);
      return;
    }

    const fetchOrder = async () => {
      try {
        const response = await api.get<Order>(
          `/orders/checkout-session/${sessionId}`,
        );
        setOrder(response.data);

        // Celebration!
        confetti({
          particleCount: 150,
          spread: 70,
          origin: { y: 0.6 },
          colors: ["#22c55e", "#3b82f6", "#f59e0b"],
        });

        toast.success("Payment confirmed!");
      } catch (error) {
        console.error("Failed to fetch order:", error);
        toast.error("Could not retrieve order details");
      } finally {
        setLoading(false);
      }
    };

    fetchOrder();
  }, [sessionId]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
        <div className="relative">
          <Loader2 className="w-16 h-16 animate-spin text-primary opacity-20" />
          <Loader2 className="w-16 h-16 animate-spin text-primary absolute inset-0 [animation-delay:-0.3s]" />
        </div>
        <div className="text-center space-y-2">
          <p className="text-2xl font-bold bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent">
            Confirming your payment
          </p>
          <p className="text-muted-foreground animate-pulse">
            Setting up your access to the courses...
          </p>
        </div>
      </div>
    );
  }

  if (!order) {
    return (
      <Card className="border-dashed">
        <CardContent className="py-12 text-center space-y-4">
          <div className="p-4 bg-red-50 text-red-600 rounded-full w-fit mx-auto">
            <ReceiptText className="w-8 h-8" />
          </div>
          <h2 className="text-xl font-bold">Order Not Found</h2>
          <p className="text-muted-foreground max-w-sm mx-auto">
            We couldn't find your order details, but don't worry—your enrollment
            is being processed. Check your dashboard in a few minutes.
          </p>
          <Button variant="outline">
            <Link href="/dashboard">Go to Dashboard</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-1000">
      {/* Header Section */}
      <div className="text-center space-y-4">
        <div className="inline-flex items-center justify-center p-3 bg-green-100 rounded-full mb-2">
          <CheckCircle2 className="w-12 h-12 text-green-600" />
        </div>
        <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl">
          Purchase Complete!
        </h1>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Welcome to the community, {order.order_number.split("-")[2]}. Your
          learning journey starts now.
        </p>
      </div>

      <div className="grid gap-8 lg:grid-cols-5 items-start">
        {/* Order Details */}
        <div className="lg:col-span-3 space-y-6">
          <Card className="overflow-hidden border-none shadow-xl bg-gradient-to-br from-white to-slate-50">
            <CardHeader className="border-b bg-white/50 backdrop-blur-sm">
              <CardTitle className="text-lg flex items-center gap-2">
                <PlayCircle className="w-5 h-5 text-primary" />
                Purchased Courses
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <div className="divide-y divide-slate-100">
                {order.items.map((item) => (
                  <div
                    key={item.id}
                    className="p-4 flex gap-4 hover:bg-white/80 transition-colors"
                  >
                    <div className="relative w-24 h-14 rounded-md overflow-hidden bg-slate-100 shrink-0 shadow-sm border">
                      {item.course.thumbnail_url ? (
                        <Image
                          src={item.course.thumbnail_url}
                          alt={item.course.title}
                          fill
                          className="object-cover"
                        />
                      ) : (
                        <div className="absolute inset-0 flex items-center justify-center">
                          <PlayCircle className="w-6 h-6 text-slate-300" />
                        </div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0 py-1">
                      <h3 className="font-semibold text-sm line-clamp-1">
                        {item.course.title}
                      </h3>
                      <p className="text-xs text-muted-foreground">
                        by {item.course.instructor?.first_name}{" "}
                        {item.course.instructor?.last_name}
                      </p>
                    </div>
                    <div className="text-sm font-bold self-center">
                      ${item.price.toFixed(2)}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
            <CardFooter className="bg-slate-50/50 p-4 border-t flex justify-between items-center">
              <span className="text-sm font-medium text-muted-foreground">
                Amount Paid
              </span>
              <span className="text-xl font-bold text-primary">
                ${order.total.toFixed(2)}
              </span>
            </CardFooter>
          </Card>
        </div>

        {/* Info & Actions */}
        <div className="lg:col-span-2 space-y-6">
          <Card className="border-none shadow-lg">
            <CardHeader>
              <CardTitle className="text-lg font-bold">Order Summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <div className="flex justify-between items-center text-muted-foreground">
                <div className="flex items-center gap-2">
                  <ReceiptText className="w-4 h-4" />
                  Order Number
                </div>
                <span className="font-mono text-foreground font-medium">
                  {order.order_number}
                </span>
              </div>
              <div className="flex justify-between items-center text-muted-foreground">
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4" />
                  Date
                </div>
                <span className="text-foreground font-medium">
                  {new Date(order.created_at).toLocaleDateString(undefined, {
                    dateStyle: "medium",
                  })}
                </span>
              </div>
              <div className="flex justify-between items-center text-muted-foreground">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4" />
                  Status
                </div>
                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold bg-green-100 text-green-700 uppercase">
                  Confirmed
                </span>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4">
            <Button
              size="lg"
              className="h-14 text-lg font-bold shadow-lg shadow-primary/20 hover:shadow-xl transition-all"
            >
              <Link href="/dashboard/my-courses">
                Start Learning Now
                <ArrowRight className="w-5 h-5 ml-2" />
              </Link>
            </Button>
            <Button variant="outline" size="lg" className="h-14">
              <Link href="/courses">Browse More Courses</Link>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function CheckoutSuccessPage() {
  return (
    <div className="min-h-screen bg-slate-50/50">
      <main className="container max-w-6xl py-12 lg:py-20 px-4">
        <Suspense
          fallback={
            <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
              <Loader2 className="w-12 h-12 animate-spin text-primary" />
              <p className="text-xl font-medium">Loading success page...</p>
            </div>
          }
        >
          <SuccessContent />
        </Suspense>
      </main>
    </div>
  );
}
