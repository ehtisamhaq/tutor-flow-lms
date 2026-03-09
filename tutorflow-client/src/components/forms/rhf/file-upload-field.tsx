"use client";

import { Controller, useFormContext, useWatch } from "react-hook-form";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";
import { FileUpload } from "../file-upload";

export interface FileUploadFieldProps {
  name: string;
  removeFieldName?: string;
  label: string;
  accept?: string[];
  maxSizeMB?: number;
  existingFileName?: string | null;
  className?: string;
  description?: string;
}

export function FileUploadField({
  name,
  removeFieldName,
  label,
  accept,
  maxSizeMB = 50,
  existingFileName,
  className,
  description,
}: FileUploadFieldProps) {
  const { control, setValue } = useFormContext();
  // Always call useWatch, but use a fallback name if removeFieldName is not provided
  const removeValue = useWatch({ name: removeFieldName || "__unused__" });
  const effectiveRemoveValue = removeFieldName ? removeValue : false;

  return (
    <Controller
      name={name}
      control={control}
      render={({ field, fieldState }) => {
        const fileValue = field.value;
        const hasError = !!fieldState.error;
        const showExistingFile =
          existingFileName && !effectiveRemoveValue && !fileValue;

        return (
          <Field data-invalid={hasError} className={className}>
            <FieldLabel htmlFor={name}>{label}</FieldLabel>
            {showExistingFile ? (
              <div className="flex items-center justify-between border rounded-lg p-3 bg-muted/50">
                <span className="text-sm truncate max-w-75">
                  Current: {existingFileName}
                </span>
                {removeFieldName && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => setValue(removeFieldName, true)}
                    className="text-destructive hover:text-destructive"
                  >
                    Remove
                  </Button>
                )}
              </div>
            ) : (
              <FileUpload
                files={fileValue as File | undefined}
                onChange={(file) => field.onChange(file)}
                accept={accept}
                maxSizeMB={maxSizeMB}
              />
            )}
            {description && <FieldDescription>{description}</FieldDescription>}
            {hasError && <FieldError>{fieldState.error?.message}</FieldError>}
          </Field>
        );
      }}
    />
  );
}
