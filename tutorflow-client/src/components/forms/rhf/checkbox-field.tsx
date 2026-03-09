"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

export interface CheckboxFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  disabled?: boolean;
  className?: string;
  description?: string;
}

export function CheckboxField<T extends FieldValues>({
  name,
  label,
  disabled,
  className,
  description,
}: CheckboxFieldProps<T>) {
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
          <Checkbox
            id={field.name}
            checked={field.value ?? false}
            onCheckedChange={field.onChange}
            disabled={disabled}
            aria-invalid={fieldState.invalid}
          />
          <div className="flex flex-col gap-1">
            <FieldLabel htmlFor={field.name} className="font-normal">
              {label}
            </FieldLabel>
            {description && <FieldDescription>{description}</FieldDescription>}
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </div>
        </Field>
      )}
    />
  );
}
