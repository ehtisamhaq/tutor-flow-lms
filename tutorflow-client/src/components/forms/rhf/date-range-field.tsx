"use client";

import * as React from "react";
import { format, parseISO, isValid } from "date-fns";
import { IconCalendar } from "@tabler/icons-react";
import { DateRange } from "react-day-picker";
import { useFormContext, useWatch } from "react-hook-form";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Field, FieldLabel, FieldError } from "@/components/ui/field";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

export interface DateRangeFieldProps {
  startName: string;
  endName: string;
  label: string;
  disabled?: boolean;
  className?: string;
}

export function DateRangeField({
  startName,
  endName,
  label,
  disabled,
  className,
}: DateRangeFieldProps) {
  const {
    setValue,
    formState: { errors },
    trigger,
  } = useFormContext();

  const startDateValue = useWatch({ name: startName }) as string | undefined;
  const endDateValue = useWatch({ name: endName }) as string | undefined;

  const startError = errors[startName]?.message as string | undefined;
  const endError = errors[endName]?.message as string | undefined;
  const hasError = !!(startError || endError);

  const dateRange: DateRange | undefined = React.useMemo(() => {
    const from = startDateValue
      ? startDateValue.includes("T")
        ? parseISO(startDateValue)
        : new Date(startDateValue + "T00:00:00")
      : undefined;
    const to = endDateValue
      ? endDateValue.includes("T")
        ? parseISO(endDateValue)
        : new Date(endDateValue + "T00:00:00")
      : undefined;

    return {
      from: from && isValid(from) ? from : undefined,
      to: to && isValid(to) ? to : undefined,
    };
  }, [startDateValue, endDateValue]);

  const handleSelect = (range: DateRange | undefined) => {
    if (range?.from) {
      setValue(startName, format(range.from, "yyyy-MM-dd"));
    } else {
      setValue(startName, "");
    }

    if (range?.to) {
      setValue(endName, format(range.to, "yyyy-MM-dd"));
    } else {
      setValue(endName, "");
    }
  };

  const handleBlur = () => {
    trigger(startName);
    trigger(endName);
  };

  return (
    <Field
      data-invalid={hasError}
      data-disabled={disabled}
      className={cn("flex flex-col gap-1", className)}
    >
      <FieldLabel>{label}</FieldLabel>
      <Popover>
        <PopoverTrigger
          render={
            <Button
              variant={"outline"}
              className={cn(
                "w-full justify-start text-left font-normal",
                !dateRange?.from && "text-muted-foreground",
              )}
              disabled={disabled}
              onBlur={handleBlur}
            />
          }
        >
          <IconCalendar className="mr-2 h-4 w-4" />
          {dateRange?.from ? (
            dateRange.to ? (
              <>
                {format(dateRange.from, "LLL dd, y")} -{" "}
                {format(dateRange.to, "LLL dd, y")}
              </>
            ) : (
              format(dateRange.from, "LLL dd, y")
            )
          ) : (
            <span>Pick a date range</span>
          )}
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            initialFocus
            mode="range"
            defaultMonth={dateRange?.from}
            selected={dateRange}
            onSelect={handleSelect}
            numberOfMonths={2}
          />
        </PopoverContent>
      </Popover>
      {startError && <FieldError>{startError}</FieldError>}
      {endError && <FieldError>{endError}</FieldError>}
    </Field>
  );
}
