import Link from "next/link";
import { redirect } from "next/navigation";
import { cookies } from "next/headers";
import {
  GraduationCap,
  LayoutDashboard,
  BookOpen,
  MessageSquare,
  Bell,
  Settings,
  User,
  LogOut,
  ShoppingCart,
  ClipboardCheck,
  CreditCard,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { DashboardSidebar } from "@/components/dashboard/sidebar";
import { DashboardHeader } from "@/components/dashboard/header";

// Server-side auth check
async function checkAuth() {
  const cookieStore = await cookies();
  const token = cookieStore.get("accessToken");
  if (!token) {
    redirect("/login");
  }
  return token.value;
}

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  await checkAuth();

  const navItems = [
    { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { href: "/dashboard/my-courses", label: "My Courses", icon: BookOpen },
    {
      href: "/dashboard/peer-reviews",
      label: "Peer Reviews",
      icon: ClipboardCheck,
    },
    {
      href: "/dashboard/subscription",
      label: "Subscription",
      icon: CreditCard,
    },
    { href: "/dashboard/messages", label: "Messages", icon: MessageSquare },
    { href: "/dashboard/notifications", label: "Notifications", icon: Bell },
    { href: "/dashboard/settings", label: "Settings", icon: Settings },
  ];

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="hidden lg:flex w-64 flex-col border-r bg-card">
        <div className="p-6 border-b">
          <Link href="/" className="flex items-center gap-2">
            <GraduationCap className="h-8 w-8 text-primary" />
            <span className="text-xl font-bold">TutorFlow</span>
          </Link>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="flex items-center gap-3 px-3 py-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
            >
              <item.icon className="h-5 w-5" />
              <span>{item.label}</span>
            </Link>
          ))}
        </nav>

        <div className="p-4 border-t">
          <DashboardSidebar />
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col">
        <DashboardHeader />
        <main className="flex-1 p-6 bg-muted/30">{children}</main>
      </div>
    </div>
  );
}
