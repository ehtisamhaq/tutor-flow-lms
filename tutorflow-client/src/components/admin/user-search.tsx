"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Search, X } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { useDebounce } from "@/lib/hooks/use-debounce";

export function UserSearch() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [query, setQuery] = useState(searchParams.get("q") || "");
  const debouncedQuery = useDebounce(query, 500);

  useEffect(() => {
    const currentQuery = searchParams.get("q") || "";
    if (currentQuery === debouncedQuery) return;

    const params = new URLSearchParams(searchParams.toString());
    if (debouncedQuery) {
      params.set("q", debouncedQuery);
    } else {
      params.delete("q");
    }
    params.set("page", "1"); // Reset to first page on search
    router.push(`/admin/users?${params.toString()}`);
  }, [debouncedQuery, router, searchParams]);

  return (
    <div className="relative w-full max-w-sm">
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
      <Input
        placeholder="Search users by name or email..."
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="pl-9 pr-9"
      />
      {query && (
        <button
          onClick={() => setQuery("")}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}
