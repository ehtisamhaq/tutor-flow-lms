import Link from "next/link";
import Image from "next/image";
import { redirect } from "next/navigation";
import { GraduationCap, Trash2, ShoppingBag } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { RemoveFromCartButton } from "@/components/cart/remove-button";
import { CheckoutButton } from "@/components/cart/checkout-button";
import { authServerFetch } from "@/lib/server-api";

interface CartItem {
  id: string;
  course_id: string;
  course: {
    id: string;
    title: string;
    slug: string;
    thumbnail_url?: string;
    price: number;
    discount_price?: number;
    instructor: {
      first_name: string;
      last_name: string;
    };
  };
}

interface Cart {
  items: CartItem[];
  total: number;
}

// Server Component with SSR data fetching
export default async function CartPage() {
  const cart = await authServerFetch<Cart>("/cart", {}, false);
  const items = cart?.items || [];
  const total = items.reduce((sum, item) => {
    const price = item.course.discount_price ?? item.course.price;
    return sum + price;
  }, 0);

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <Link href="/" className="flex items-center gap-2">
            <GraduationCap className="h-8 w-8 text-primary" />
            <span className="text-xl font-bold">TutorFlow</span>
          </Link>
          <div className="flex items-center gap-3">
            <Button variant="ghost" asChild>
              <Link href="/dashboard">Dashboard</Link>
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Shopping Cart</h1>

        {items.length === 0 ? (
          <Card className="text-center py-16">
            <CardContent>
              <ShoppingBag className="h-16 w-16 mx-auto mb-4 text-muted-foreground" />
              <h2 className="text-xl font-semibold mb-2">Your cart is empty</h2>
              <p className="text-muted-foreground mb-6">
                Browse our courses and add some to your cart
              </p>
              <Button asChild>
                <Link href="/courses">Browse Courses</Link>
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2 space-y-4">
              {items.map((item) => (
                <Card key={item.id}>
                  <CardContent className="p-4">
                    <div className="flex gap-4">
                      <div className="relative w-32 h-20 bg-muted rounded-lg overflow-hidden shrink-0">
                        {item.course.thumbnail_url ? (
                          <Image
                            src={item.course.thumbnail_url}
                            alt={item.course.title}
                            fill
                            className="object-cover"
                          />
                        ) : (
                          <div className="absolute inset-0 flex items-center justify-center">
                            <GraduationCap className="h-8 w-8 text-muted-foreground" />
                          </div>
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <Link
                          href={`/courses/${item.course.slug}`}
                          className="font-semibold hover:text-primary transition-colors line-clamp-2"
                        >
                          {item.course.title}
                        </Link>
                        <p className="text-sm text-muted-foreground">
                          {item.course.instructor.first_name}{" "}
                          {item.course.instructor.last_name}
                        </p>
                      </div>
                      <div className="flex flex-col items-end gap-2">
                        <div className="text-lg font-bold">
                          $
                          {(
                            item.course.discount_price ?? item.course.price
                          ).toFixed(2)}
                        </div>
                        {item.course.discount_price && (
                          <div className="text-sm text-muted-foreground line-through">
                            ${item.course.price.toFixed(2)}
                          </div>
                        )}
                        <RemoveFromCartButton
                          itemId={item.id}
                          courseId={item.course_id}
                        />
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>

            {/* Order Summary */}
            <div className="lg:col-span-1">
              <Card className="sticky top-24">
                <CardHeader>
                  <CardTitle>Order Summary</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Subtotal</span>
                    <span>${total.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Discount</span>
                    <span className="text-green-600">-$0.00</span>
                  </div>
                  <div className="border-t pt-4 flex justify-between font-semibold text-lg">
                    <span>Total</span>
                    <span>${total.toFixed(2)}</span>
                  </div>
                </CardContent>
                <CardFooter>
                  <CheckoutButton total={total} />
                </CardFooter>
              </Card>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
