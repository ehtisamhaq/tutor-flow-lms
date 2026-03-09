"use client";

import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export interface SelectOption {
  value: string;
  label: string;
}

export interface SelectFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  options: SelectOption[];
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  description?: string;
  /** Optional callback when value changes */
  onValueChange?: (value: string) => void;
}

export function SelectField<T extends FieldValues>({
  name,
  label,
  options,
  placeholder = "Select an option",
  disabled,
  className,
  description,
  onValueChange,
}: SelectFieldProps<T>) {
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
          <Select
            items={options}
            onValueChange={(val) => {
              field.onChange(val);
              onValueChange?.(val as string);
            }}
            value={field.value || ""}
            disabled={disabled}
          >
            <SelectTrigger
              id={field.name}
              className="w-full"
              aria-invalid={fieldState.invalid}
            >
              <SelectValue placeholder={placeholder} />
            </SelectTrigger>
            <SelectContent>
              {options.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {description && <FieldDescription>{description}</FieldDescription>}
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  );
}
