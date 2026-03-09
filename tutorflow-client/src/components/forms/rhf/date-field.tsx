"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import { Input } from "@/components/ui/input";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

export interface DateFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  disabled?: boolean;
  className?: string;
  description?: string;
  min?: string;
  max?: string;
}

export function DateField<T extends FieldValues>({
  name,
  label,
  disabled,
  className,
  description,
  min,
  max,
}: DateFieldProps<T>) {
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
            type="date"
            disabled={disabled}
            aria-invalid={fieldState.invalid}
            value={field.value ?? ""}
            min={min}
            max={max}
          />
          {description && <FieldDescription>{description}</FieldDescription>}
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  );
}
