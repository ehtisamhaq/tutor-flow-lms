import Link from "next/link";
import { authServerFetch, type PaginatedResponse } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Users,
  Plus,
  MoreHorizontal,
  Shield,
  User,
  GraduationCap,
} from "lucide-react";

interface AdminUser {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: "admin" | "manager" | "tutor" | "student";
  status: "active" | "inactive" | "suspended";
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
  searchParams: Promise<{ page?: string; role?: string }>;
}) {
  const params = await searchParams;
  const page = Number(params.page) || 1;
  const role = params.role || "";

  const users = await authServerFetch<PaginatedResponse<AdminUser>>(
    `/admin/users?page=${page}&limit=20${role ? `&role=${role}` : ""}`
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Users</h2>
          <p className="text-muted-foreground">
            Manage all users on the platform
          </p>
        </div>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Add User
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-2">
        {["", "admin", "manager", "tutor", "student"].map((r) => (
          <Button
            key={r}
            variant={role === r ? "default" : "outline"}
            size="sm"
            asChild
          >
            <Link href={`/admin/users${r ? `?role=${r}` : ""}`}>
              {r || "All"}
            </Link>
          </Button>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{users?.total?.toLocaleString() || 0} Users</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {users?.items?.length ? (
            <div className="divide-y">
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
                        className={`text-xs px-2 py-1 rounded-full capitalize ${getRoleBadgeColor(
                          user.role
                        )}`}
                      >
                        {user.role}
                      </span>
                      <span
                        className={`text-xs px-2 py-1 rounded-full ${
                          user.status === "active"
                            ? "bg-green-100 text-green-800"
                            : "bg-gray-100 text-gray-800"
                        }`}
                      >
                        {user.status}
                      </span>
                      <Button variant="ghost" size="icon">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
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
