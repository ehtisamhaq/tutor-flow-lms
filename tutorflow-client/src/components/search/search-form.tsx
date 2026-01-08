"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface SearchFormProps {
  defaultValue?: string;
}

export function SearchForm({ defaultValue = "" }: SearchFormProps) {
  const router = useRouter();
  const [query, setQuery] = useState(defaultValue);
  const [isPending, startTransition] = useTransition();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      startTransition(() => {
        router.push(`/search?q=${encodeURIComponent(query.trim())}`);
      });
    }
  };

  return (
    <form onSubmit={handleSubmit} className="relative">
      <Input
        type="text"
        placeholder="Search courses..."
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="pr-12"
      />
      <Button
        type="submit"
        size="icon"
        variant="ghost"
        className="absolute right-1 top-1/2 -translate-y-1/2 h-8 w-8"
        disabled={isPending}
      >
        <Search className="h-4 w-4" />
      </Button>
    </form>
  );
}
