import Link from "next/link";
import { authServerFetch, type PaginatedResponse } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { UserSearch } from "@/components/admin/user-search";
import { UserActions } from "@/components/admin/user-actions";
import { Users, Plus, Shield, User, GraduationCap } from "lucide-react";

interface AdminUser {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: "admin" | "manager" | "tutor" | "student";
  status: "active" | "inactive" | "suspended" | "pending";
  created_at: string;
}

function getRoleIcon(role: string) {
  switch (role) {
    case "admin":
      return Shield;
    case "tutor":
      return GraduationCap;
    default:
      return User;
  }
}

function getRoleBadgeColor(role: string) {
  switch (role) {
    case "admin":
      return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200";
    case "manager":
      return "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200";
    case "tutor":
      return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200";
  }
}

// Server Component
export default async function AdminUsersPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; role?: string; q?: string }>;
}) {
  const params = await searchParams;
  const page = Number(params.page) || 1;
  const role = params.role || "";
  const query = params.q || "";

  // Build query string
  const queryParts = [`page=${page}`, "limit=20"];
  if (role) queryParts.push(`role=${role}`);
  if (query) queryParts.push(`search=${encodeURIComponent(query)}`);

  const users = await authServerFetch<PaginatedResponse<AdminUser>>(
    `/users?${queryParts.join("&")}`
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold">Users</h2>
          <p className="text-muted-foreground">
            Manage all users on the platform
          </p>
        </div>
        <Button asChild>
          <Link href="/admin/users/new">
            <Plus className="mr-2 h-4 w-4" />
            Add User
          </Link>
        </Button>
      </div>

      <div className="flex flex-col lg:flex-row gap-4 items-start lg:items-center justify-between">
        {/* Filters */}
        <div className="flex flex-wrap gap-2">
          {["", "admin", "manager", "tutor", "student"].map((r) => (
            <Button
              key={r}
              variant={role === r ? "default" : "outline"}
              size="sm"
              asChild
            >
              <Link
                href={`/admin/users?role=${r}${query ? `&q=${query}` : ""}`}
              >
                {r ? r.charAt(0).toUpperCase() + r.slice(1) : "All"}
              </Link>
            </Button>
          ))}
        </div>

        {/* Search */}
        <UserSearch />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{users?.total?.toLocaleString() || 0} Users</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {users?.items?.length ? (
            <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
              {users.items.map((user) => {
                const RoleIcon = getRoleIcon(user.role);
                return (
                  <div
                    key={user.id}
                    className="flex items-center justify-between p-4 hover:bg-muted/50"
                  >
                    <div className="flex items-center gap-4">
                      <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                        <RoleIcon className="h-5 w-5 text-primary" />
                      </div>
                      <div>
                        <p className="font-medium">
                          {user.first_name} {user.last_name}
                        </p>
                        <p className="text-sm text-muted-foreground">
                          {user.email}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <span
                        className={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase tracking-tight ${getRoleBadgeColor(
                          user.role
                        )}`}
                      >
                        {user.role}
                      </span>
                      <span
                        className={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase tracking-tight ${
                          user.status === "active"
                            ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                            : user.status === "suspended"
                            ? "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
                            : "bg-zinc-100 text-zinc-800 dark:bg-zinc-800 dark:text-zinc-200"
                        }`}
                      >
                        {user.status}
                      </span>
                      <UserActions user={user} />
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="text-center py-16">
              <Users className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <p className="text-muted-foreground">No users found</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {users && users.total > 20 && (
        <div className="flex justify-center gap-2">
          {page > 1 && (
            <Button variant="outline" asChild>
              <Link
                href={`/admin/users?page=${page - 1}${
                  role ? `&role=${role}` : ""
                }`}
              >
                Previous
              </Link>
            </Button>
          )}
          {page * 20 < users.total && (
            <Button variant="outline" asChild>
              <Link
                href={`/admin/users?page=${page + 1}${
                  role ? `&role=${role}` : ""
                }`}
              >
                Next
              </Link>
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
