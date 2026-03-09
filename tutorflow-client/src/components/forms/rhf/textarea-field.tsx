"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import { Textarea } from "@/components/ui/textarea";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

export interface TextareaFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  description?: string;
  rows?: number;
}

export function TextareaField<T extends FieldValues>({
  name,
  label,
  placeholder,
  disabled,
  className,
  description,
  rows = 3,
}: TextareaFieldProps<T>) {
  const { control } = useFormContext();

  return (
    <Controller
      name={name}
      control={control}
      render={({ field, fieldState }) => (
        <Field
          data-invalid={fieldState.invalid}
          data-disabled={disabled}
          className={className}
        >
          <FieldLabel htmlFor={field.name}>{label}</FieldLabel>
          <Textarea
            {...field}
            id={field.name}
            placeholder={placeholder}
            disabled={disabled}
            aria-invalid={fieldState.invalid}
            rows={rows}
            value={field.value ?? ""}
          />
          {description && <FieldDescription>{description}</FieldDescription>}
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  );
}
