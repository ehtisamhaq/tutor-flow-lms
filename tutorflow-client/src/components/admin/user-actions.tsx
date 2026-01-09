"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  MoreHorizontal,
  Shield,
  User,
  GraduationCap,
  XCircle,
  CheckCircle,
  Trash2,
  ShieldAlert,
} from "lucide-react";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

interface UserActionsProps {
  user: {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    role: "admin" | "manager" | "tutor" | "student";
    status: "active" | "inactive" | "suspended" | "pending";
  };
}

export function UserActions({ user }: UserActionsProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleUpdateStatus = async (status: string) => {
    setLoading(true);
    try {
      await api.patch(`/users/${user.id}/status`, { status });
      toast.success(`User status updated to ${status}`);
      router.refresh();
      setIsOpen(false);
    } catch (error: any) {
      toast.error(
        error.response?.data?.error?.message || "Failed to update status"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateRole = async (role: string) => {
    setLoading(true);
    try {
      await api.patch(`/users/${user.id}/role`, { role });
      toast.success(`User role updated to ${role}`);
      router.refresh();
      setIsOpen(false);
    } catch (error: any) {
      toast.error(
        error.response?.data?.error?.message || "Failed to update role"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (
      !confirm(
        `Are you sure you want to delete ${user.first_name}? This action cannot be undone.`
      )
    ) {
      return;
    }

    setLoading(true);
    try {
      await api.delete(`/users/${user.id}`);
      toast.success("User deleted successfully");
      router.refresh();
      setIsOpen(false);
    } catch (error: any) {
      toast.error(
        error.response?.data?.error?.message || "Failed to delete user"
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative">
      <Button
        variant="ghost"
        size="icon"
        onClick={() => setIsOpen(!isOpen)}
        disabled={loading}
      >
        <MoreHorizontal className="h-4 w-4" />
      </Button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute right-0 mt-2 w-56 rounded-md shadow-lg bg-white dark:bg-zinc-900 ring-1 ring-black ring-opacity-5 z-20 overflow-hidden border border-zinc-200 dark:border-zinc-800">
            <div className="py-1 divide-y divide-zinc-100 dark:divide-zinc-800">
              {/* Status Section */}
              <div className="px-3 py-2">
                <p className="text-[10px] font-semibold text-zinc-500 uppercase tracking-wider mb-1">
                  Status
                </p>
                <div className="grid grid-cols-2 gap-1">
                  <button
                    onClick={() => handleUpdateStatus("active")}
                    disabled={user.status === "active"}
                    className="flex items-center gap-2 px-2 py-1 text-xs rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-50"
                  >
                    <CheckCircle className="h-3 w-3 text-green-500" /> Active
                  </button>
                  <button
                    onClick={() => handleUpdateStatus("suspended")}
                    disabled={user.status === "suspended"}
                    className="flex items-center gap-2 px-2 py-1 text-xs rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-50"
                  >
                    <XCircle className="h-3 w-3 text-red-500" /> Suspend
                  </button>
                </div>
              </div>

              {/* Role Section */}
              <div className="px-3 py-2">
                <p className="text-[10px] font-semibold text-zinc-500 uppercase tracking-wider mb-1">
                  Update Role
                </p>
                <div className="grid grid-cols-2 gap-1">
                  <button
                    onClick={() => handleUpdateRole("student")}
                    disabled={user.role === "student"}
                    className="flex items-center gap-2 px-2 py-1 text-xs rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-50"
                  >
                    <User className="h-3 w-3" /> Student
                  </button>
                  <button
                    onClick={() => handleUpdateRole("tutor")}
                    disabled={user.role === "tutor"}
                    className="flex items-center gap-2 px-2 py-1 text-xs rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-50"
                  >
                    <GraduationCap className="h-3 w-3" /> Tutor
                  </button>
                  <button
                    onClick={() => handleUpdateRole("manager")}
                    disabled={user.role === "manager"}
                    className="flex items-center gap-2 px-2 py-1 text-xs rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-50"
                  >
                    <Shield className="h-3 w-3" /> Manager
                  </button>
                  <button
                    onClick={() => handleUpdateRole("admin")}
                    disabled={user.role === "admin"}
                    className="flex items-center gap-2 px-2 py-1 text-xs rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-50"
                  >
                    <ShieldAlert className="h-3 w-3 text-red-500" /> Admin
                  </button>
                </div>
              </div>

              {/* Delete Section */}
              <div className="px-3 py-1">
                <button
                  onClick={handleDelete}
                  className="flex w-full items-center gap-2 px-2 py-2 text-xs text-red-600 rounded hover:bg-red-50 dark:hover:bg-red-950/20"
                >
                  <Trash2 className="h-3 w-3" /> Delete User
                </button>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
