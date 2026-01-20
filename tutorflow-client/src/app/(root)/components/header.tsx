import { Button } from "@/components/ui/button";
import { GraduationCap, Search, ShoppingCart } from "lucide-react";
import Link from "next/link";
import { cookies } from "next/headers";

export default async function Header() {
  const cookieStore = await cookies();
  const isAuthenticated = cookieStore.has("accessToken");

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        <Link href="/" className="flex items-center gap-2">
          <GraduationCap className="h-8 w-8 text-primary" />
          <span className="text-xl font-bold">TutorFlow</span>
        </Link>

        <nav className="hidden md:flex items-center gap-6">
          <Link
            href="/courses"
            className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            Courses
          </Link>
          <Link
            href="/bundles"
            className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            Bundles
          </Link>
          <Link
            href="/learning-paths"
            className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            Learning Paths
          </Link>
          <Link
            href="/pricing"
            className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            Pricing
          </Link>
        </nav>

        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" asChild>
            <Link href="/search">
              <Search className="h-5 w-5" />
              <span className="sr-only">Search</span>
            </Link>
          </Button>
          <Button variant="ghost" size="icon" asChild>
            <Link href="/cart">
              <ShoppingCart className="h-5 w-5" />
              <span className="sr-only">Cart</span>
            </Link>
          </Button>

          {isAuthenticated ? (
            <Button asChild>
              <Link href="/dashboard">Go to Dashboard</Link>
            </Button>
          ) : (
            <>
              <Button variant="ghost" asChild>
                <Link href="/login">Sign In</Link>
              </Button>
              <Button asChild>
                <Link href="/register">Get Started</Link>
              </Button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
