"use client";

import { Controller, useFormContext } from "react-hook-form";
import { Badge } from "@/components/ui/badge";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";

const DAYS = [
  { value: "monday", name: "Monday" },
  { value: "tuesday", name: "Tuesday" },
  { value: "wednesday", name: "Wednesday" },
  { value: "thursday", name: "Thursday" },
  { value: "friday", name: "Friday" },
  { value: "saturday", name: "Saturday" },
  { value: "sunday", name: "Sunday" },
];

const DEFAULT_DAY_NAMES = DAYS.map((day) => day.name.slice(0, 3));

export interface WorkingDaysFieldProps {
  name: string;
  label: string;
  dayNames?: string[];
  disabled?: boolean;
  className?: string;
  description?: string;
}

export function WorkingDaysField({
  name,
  label,
  dayNames = DEFAULT_DAY_NAMES,
  disabled,
  className,
  description,
}: WorkingDaysFieldProps) {
  const { control } = useFormContext();

  return (
    <Controller
      name={name}
      control={control}
      render={({ field, fieldState }) => {
        const selectedDays = (field.value as string[]) || [];
        const hasError = !!fieldState.error;

        const toggleDay = (day: string) => {
          if (disabled) return;

          if (selectedDays.includes(day)) {
            field.onChange(selectedDays.filter((d) => d !== day));
          } else {
            field.onChange([...selectedDays, day]);
          }
        };
        console.log("Selected days:", selectedDays);

        return (
          <Field
            data-invalid={hasError}
            data-disabled={disabled}
            className={className}
          >
            <FieldLabel>{label}</FieldLabel>
            <div className="flex flex-wrap gap-2">
              {DAYS.map((day) => (
                <Badge
                  key={day.value}
                  variant={
                    selectedDays.includes(day.value) ? "default" : "outline"
                  }
                  className={`cursor-pointer ${
                    disabled ? "opacity-50 cursor-not-allowed" : ""
                  }`}
                  onClick={() => toggleDay(day.value)}
                >
                  {day.name.slice(0, 3)}
                </Badge>
              ))}
            </div>
            {description && <FieldDescription>{description}</FieldDescription>}
            {hasError && <FieldError>{fieldState.error?.message}</FieldError>}
          </Field>
        );
      }}
    />
  );
}
