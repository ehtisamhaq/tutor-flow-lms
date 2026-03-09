"use client";

import { Controller, useFormContext } from "react-hook-form";
import { Input } from "@/components/ui/input";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

export interface TimeFieldProps {
  name: string;
  label: string;
  disabled?: boolean;
  className?: string;
  description?: string;
}

export function TimeField({
  name,
  label,
  disabled,
  className,
  description,
}: TimeFieldProps) {
  const { control } = useFormContext();

  return (
    <Controller
      name={name}
      control={control}
      render={({ field, fieldState }) => {
        const hasError = !!fieldState.error;

        return (
          <Field
            data-invalid={hasError}
            data-disabled={disabled}
            className={className}
          >
            <FieldLabel htmlFor={name}>{label}</FieldLabel>
            <Input
              id={name}
              type="time"
              disabled={disabled}
              value={field.value ?? ""}
              onBlur={field.onBlur}
              onChange={(e) => field.onChange(e.target.value)}
              aria-invalid={hasError}
            />
            {description && <FieldDescription>{description}</FieldDescription>}
            {hasError && <FieldError>{fieldState.error?.message}</FieldError>}
          </Field>
        );
      }}
    />
  );
}
