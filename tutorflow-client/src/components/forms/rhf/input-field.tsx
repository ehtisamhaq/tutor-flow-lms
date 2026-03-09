"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import { Input } from "@/components/ui/input";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

export interface InputFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  placeholder?: string;
  type?: "text" | "email" | "password" | "number" | "url" | "tel" | "time";
  disabled?: boolean;
  className?: string;
  description?: string;
  autoComplete?: string;
}

export function InputField<T extends FieldValues>({
  name,
  label,
  placeholder,
  type = "text",
  disabled,
  className,
  description,
  autoComplete,
}: InputFieldProps<T>) {
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
            type={type}
            placeholder={placeholder}
            disabled={disabled}
            aria-invalid={fieldState.invalid}
            autoComplete={autoComplete}
            value={field.value ?? ""}
          />
          {description && <FieldDescription>{description}</FieldDescription>}
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  );
}
