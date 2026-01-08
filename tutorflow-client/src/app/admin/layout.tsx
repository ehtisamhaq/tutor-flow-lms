import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import {
  GraduationCap,
  LayoutDashboard,
  Users,
  BookOpen,
  DollarSign,
  BarChart3,
  Settings,
  ShieldCheck,
} from "lucide-react";
import { AdminSidebar } from "@/components/admin/sidebar";
import { AdminHeader } from "@/components/admin/header";

// Server-side auth check for admin
async function checkAdminAuth() {
  const cookieStore = await cookies();
  const token = cookieStore.get("accessToken");
  if (!token) {
    redirect("/login");
  }
  // In production, decode JWT and check role
  return token.value;
}

export default async function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  await checkAdminAuth();

  const navItems = [
    { href: "/admin", label: "Dashboard", icon: LayoutDashboard },
    { href: "/admin/users", label: "Users", icon: Users },
    { href: "/admin/courses", label: "Courses", icon: BookOpen },
    { href: "/admin/revenue", label: "Revenue", icon: DollarSign },
    { href: "/admin/analytics", label: "Analytics", icon: BarChart3 },
    { href: "/admin/settings", label: "Settings", icon: Settings },
  ];

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="hidden lg:flex w-64 flex-col border-r bg-card">
        <div className="p-6 border-b">
          <Link href="/admin" className="flex items-center gap-2">
            <ShieldCheck className="h-8 w-8 text-primary" />
            <span className="text-xl font-bold">Admin</span>
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
          <AdminSidebar />
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col">
        <AdminHeader />
        <main className="flex-1 p-6 bg-muted/30">{children}</main>
      </div>
    </div>
  );
}
