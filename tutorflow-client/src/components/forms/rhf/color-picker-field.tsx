"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

const DEFAULT_COLORS = [
  "#ef4444", // red
  "#f97316", // orange
  "#f59e0b", // amber
  "#84cc16", // lime
  "#22c55e", // green
  "#14b8a6", // teal
  "#06b6d4", // cyan
  "#3b82f6", // blue
  "#6366f1", // indigo
  "#8b5cf6", // violet
  "#a855f7", // purple
  "#ec4899", // pink
];

export interface ColorPickerFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  colors?: string[];
  disabled?: boolean;
  className?: string;
  description?: string;
}

export function ColorPickerField<T extends FieldValues>({
  name,
  label,
  colors = DEFAULT_COLORS,
  disabled,
  className,
  description,
}: ColorPickerFieldProps<T>) {
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
          <FieldLabel>{label}</FieldLabel>
          <div className="flex flex-wrap gap-2">
            {colors.map((color) => (
              <button
                key={color}
                type="button"
                disabled={disabled}
                className={`h-8 w-8 rounded-full border-2 transition-transform hover:scale-110 ${
                  field.value === color
                    ? "border-foreground scale-110"
                    : "border-transparent"
                } ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
                style={{ backgroundColor: color }}
                onClick={() => field.onChange(color)}
              />
            ))}
          </div>
          {description && <FieldDescription>{description}</FieldDescription>}
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  );
}
