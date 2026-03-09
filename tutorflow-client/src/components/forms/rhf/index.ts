// React Hook Form reusable components
// Central export for all form components

// Form wrapper with Zod validation
export {
  FormWrapper,
  useForm,
  useFormContext,
  type FormWrapperProps,
  type UseFormReturn,
  type Control,
  type FieldValues,
} from "./form-wrapper";

// Field components
export { InputField, type InputFieldProps } from "./input-field";
export { TextareaField, type TextareaFieldProps } from "./textarea-field";
export {
  SelectField,
  type SelectFieldProps,
  type SelectOption,
} from "./select-field";
export { CheckboxField, type CheckboxFieldProps } from "./checkbox-field";
export { SwitchField, type SwitchFieldProps } from "./switch-field";
export { DateField, type DateFieldProps } from "./date-field";
export { DateTimeField, type DateTimeFieldProps } from "./datetime-field";
export {
  ColorPickerField,
  type ColorPickerFieldProps,
} from "./color-picker-field";
export {
  ComboboxField,
  type ComboboxFieldProps,
  type ComboboxOption,
} from "./combobox-field";
export { TimeField, type TimeFieldProps } from "./time-field";
export {
  WorkingDaysField,
  type WorkingDaysFieldProps,
} from "./working-days-field";
export {
  FileUploadField,
  type FileUploadFieldProps,
} from "./file-upload-field";
export { DateRangeField, type DateRangeFieldProps } from "./date-range-field";

// Re-export Controller for custom fields
export { Controller, useWatch } from "react-hook-form";
export type {
  ControllerProps,
  ControllerRenderProps,
  ControllerFieldState,
} from "react-hook-form";
