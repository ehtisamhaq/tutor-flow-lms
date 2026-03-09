"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import { Input } from "@/components/ui/input";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

export interface DateTimeFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  disabled?: boolean;
  className?: string;
  description?: string;
}

export function DateTimeField<T extends FieldValues>({
  name,
  label,
  disabled,
  className,
  description,
}: DateTimeFieldProps<T>) {
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
          <Input
            {...field}
            id={field.name}
            type="datetime-local"
            disabled={disabled}
            aria-invalid={fieldState.invalid}
            value={field.value ?? ""}
          />
          {description && <FieldDescription>{description}</FieldDescription>}
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  );
}
