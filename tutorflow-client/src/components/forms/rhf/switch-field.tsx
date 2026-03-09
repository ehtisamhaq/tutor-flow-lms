"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import { Switch } from "@/components/ui/switch";
import {
  Field,
  FieldContent,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

export interface SwitchFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  disabled?: boolean;
  className?: string;
  description?: string;
}

export function SwitchField<T extends FieldValues>({
  name,
  label,
  disabled,
  className,
  description,
}: SwitchFieldProps<T>) {
  const { control } = useFormContext();

  return (
    <Controller
      name={name}
      control={control}
      render={({ field, fieldState }) => (
        <Field
          orientation="horizontal"
          data-invalid={fieldState.invalid}
          data-disabled={disabled}
          className={className}
        >
          <FieldContent>
            <FieldLabel htmlFor={field.name}>{label}</FieldLabel>
            {description && <FieldDescription>{description}</FieldDescription>}
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </FieldContent>
          <Switch
            id={field.name}
            checked={field.value ?? false}
            onCheckedChange={field.onChange}
            disabled={disabled}
            aria-invalid={fieldState.invalid}
          />
        </Field>
      )}
    />
  );
}
