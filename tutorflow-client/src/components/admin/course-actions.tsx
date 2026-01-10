"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  MoreHorizontal,
  Eye,
  EyeOff,
  Trash2,
  Archive,
  Star,
  CheckCircle,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import api from "@/lib/api";

interface CourseActionsProps {
  courseId: string;
  currentStatus: string;
  isFeatured?: boolean;
}

export function CourseActions({
  courseId,
  currentStatus,
  isFeatured = false,
}: CourseActionsProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleUpdateStatus = async (status: string) => {
    setLoading(true);
    try {
      await api.patch(`/admin/courses/${courseId}/status`, { status });
      toast.success(`Course status updated to ${status}`);
      router.refresh();
      setIsOpen(false);
    } catch (error) {
      toast.error("Failed to update status");
    } finally {
      setLoading(false);
    }
  };

  const handleToggleFeatured = async () => {
    setLoading(true);
    try {
      await api.patch(`/admin/courses/${courseId}/featured`, {
        is_featured: !isFeatured,
      });
      toast.success(
        isFeatured
          ? "Course removed from featured"
          : "Course marked as featured"
      );
      router.refresh();
      setIsOpen(false);
    } catch (error) {
      toast.error("Failed to update featured status");
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Are you sure? This cannot be undone.")) return;

    setLoading(true);
    try {
      await api.delete(`/admin/courses/${courseId}`);
      toast.success("Course deleted");
      router.refresh();
      setIsOpen(false);
    } catch (error) {
      toast.error("Failed to delete");
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
          <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white dark:bg-zinc-900 ring-1 ring-black ring-opacity-5 z-20 overflow-hidden border border-zinc-200 dark:border-zinc-800">
            <div className="py-1">
              <button
                onClick={() => handleUpdateStatus("published")}
                className="flex w-full items-center px-4 py-2 text-sm hover:bg-zinc-100 dark:hover:bg-zinc-800"
              >
                <Eye className="mr-2 h-4 w-4" /> Publish
              </button>
              <button
                onClick={() => handleUpdateStatus("draft")}
                className="flex w-full items-center px-4 py-2 text-sm hover:bg-zinc-100 dark:hover:bg-zinc-800"
              >
                <EyeOff className="mr-2 h-4 w-4" /> Set to Draft
              </button>
              <button
                onClick={() => handleUpdateStatus("archived")}
                className="flex w-full items-center px-4 py-2 text-sm hover:bg-zinc-100 dark:hover:bg-zinc-800"
              >
                <Archive className="mr-2 h-4 w-4" /> Archive
              </button>
              <div className="border-t border-zinc-100 dark:border-zinc-800 my-1" />
              <button
                onClick={handleToggleFeatured}
                className="flex w-full items-center px-4 py-2 text-sm hover:bg-zinc-100 dark:hover:bg-zinc-800"
              >
                <Star
                  className={`mr-2 h-4 w-4 ${
                    isFeatured ? "fill-yellow-400 text-yellow-400" : ""
                  }`}
                />
                {isFeatured ? "Unfeature" : "Feature"}
              </button>
              <div className="border-t border-zinc-100 dark:border-zinc-800 my-1" />
              <button
                onClick={handleDelete}
                className="flex w-full items-center px-4 py-2 text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-950/20"
              >
                <Trash2 className="mr-2 h-4 w-4" /> Delete Course
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
