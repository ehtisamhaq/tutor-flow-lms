"use client";

import React from "react";
import { Controller, FieldValues, Path, useFormContext } from "react-hook-form";
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field";
import {
  Combobox,
  ComboboxChip,
  ComboboxChips,
  ComboboxChipsInput,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
  ComboboxValue,
  useComboboxAnchor,
} from "@/components/ui/combobox";

export interface ComboboxOption {
  value: string;
  label: string;
}

export interface ComboboxFieldProps<T extends FieldValues> {
  name: Path<T>;
  label: string;
  options: ComboboxOption[];
  placeholder?: string;
  disabled?: boolean;
  multiple?: boolean;
  className?: string;
  description?: string;
}

export function ComboboxField<T extends FieldValues>({
  name,
  label,
  options,
  placeholder = "Select an option",
  disabled,
  multiple = false,
  className,
  description,
}: ComboboxFieldProps<T>) {
  const { control } = useFormContext();
  const anchor = useComboboxAnchor();

  return (
    <Controller
      name={name}
      control={control}
      render={({ field, fieldState }) => {
        // Normalize value to items from options for reference stability
        const comboboxValue = multiple
          ? (Array.isArray(field.value) ? field.value : [])
              .map((v: string) => options.find((opt) => opt.value === v))
              .filter((opt): opt is ComboboxOption => !!opt)
          : options.find((opt) => opt.value === field.value) || null;

        return (
          <Field
            data-invalid={fieldState.invalid}
            data-disabled={disabled}
            className={className}
          >
            <FieldLabel htmlFor={field.name}>{label}</FieldLabel>
            <Combobox
              disabled={disabled}
              multiple={multiple}
              autoHighlight
              items={options}
              value={comboboxValue}
              onValueChange={(newValue) => {
                // Combobox may return single value or array depending on internal state
                const items = (
                  Array.isArray(newValue) ? newValue : [newValue]
                ).filter(Boolean) as ComboboxOption[];

                if (multiple) {
                  field.onChange(items.map((i) => i.value));
                } else {
                  field.onChange(items[0]?.value || null);
                }
              }}
              itemToStringLabel={(item: ComboboxOption) => item.label}
              itemToStringValue={(item: ComboboxOption) => item.value}
            >
              {multiple ? (
                <ComboboxChips ref={anchor} className="w-full">
                  <ComboboxValue placeholder={placeholder}>
                    {(value: ComboboxOption | ComboboxOption[] | null) => {
                      const items = Array.isArray(value)
                        ? value
                        : value
                          ? [value]
                          : [];
                      return (
                        <React.Fragment>
                          {items.map((option) => (
                            <ComboboxChip key={option.value}>
                              {option.label}
                            </ComboboxChip>
                          ))}
                          <ComboboxChipsInput
                            placeholder={items.length === 0 ? placeholder : ""}
                          />
                        </React.Fragment>
                      );
                    }}
                  </ComboboxValue>
                </ComboboxChips>
              ) : (
                <ComboboxInput placeholder={placeholder} />
              )}

              <ComboboxContent anchor={anchor}>
                <ComboboxEmpty>No items found.</ComboboxEmpty>
                <ComboboxList>
                  {(item: ComboboxOption) => (
                    <ComboboxItem key={item.value} value={item}>
                      {item.label}
                    </ComboboxItem>
                  )}
                </ComboboxList>
              </ComboboxContent>
            </Combobox>
            {description && <FieldDescription>{description}</FieldDescription>}
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </Field>
        );
      }}
    />
  );
}
