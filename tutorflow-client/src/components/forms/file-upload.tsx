"use client";

import { useCallback, useRef } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export interface FileUploadProps {
  files?: File;
  onChange: (file: File | undefined) => void;
  accept?: string[];
  maxSizeMB?: number;
  className?: string;
}

export function FileUpload({
  files,
  onChange,
  accept,
  maxSizeMB = 50,
  className,
}: FileUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) {
        // Check file size
        const sizeMB = file.size / (1024 * 1024);
        if (sizeMB > maxSizeMB) {
          toast.error(`File size must be less than ${maxSizeMB}MB`);
          return;
        }
        onChange(file);
      }
    },
    [onChange, maxSizeMB],
  );

  const handleRemove = useCallback(() => {
    onChange(undefined);
    if (inputRef.current) {
      inputRef.current.value = "";
    }
  }, [onChange]);

  const acceptString = accept?.join(",");

  return (
    <div className={cn("space-y-2", className)}>
      <input
        ref={inputRef}
        type="file"
        accept={acceptString}
        onChange={handleFileChange}
        className="hidden"
      />
      {files ? (
        <div className="flex items-center justify-between border rounded-lg p-3 bg-muted/50">
          <span className="text-sm truncate max-w-75">{files.name}</span>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleRemove}
            className="text-destructive hover:text-destructive"
          >
            Remove
          </Button>
        </div>
      ) : (
        <Button
          type="button"
          variant="outline"
          onClick={() => inputRef.current?.click()}
          className="w-full"
        >
          Choose File
        </Button>
      )}
    </div>
  );
}
