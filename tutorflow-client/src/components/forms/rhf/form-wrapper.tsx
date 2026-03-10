"use client";

import { ReactNode } from "react";
import {
  useForm,
  FormProvider,
  UseFormReturn,
  DefaultValues,
  UseFormProps,
  FieldValues,
} from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { FieldGroup } from "@/components/ui/field";

export interface FormWrapperProps<TFormValues extends FieldValues> {
  /** Zod schema for form validation */
  schema: z.ZodType<TFormValues>;
  /** Default values matching the schema shape */
  defaultValues: DefaultValues<TFormValues>;
  /** Called with validated values and the form instance on successful submission */
  onSubmit: (
    values: TFormValues,
    form: UseFormReturn<TFormValues>,
  ) => Promise<void> | void;
  /** Optional cancel handler */
  onCancel?: () => void;
  /** Form content - can be JSX or render function receiving form instance */
  children: ReactNode | ((form: UseFormReturn<TFormValues>) => ReactNode);
  /** External loading state (e.g., API call in progress) */
  isLoading?: boolean;
  /** Changes default submit label to "Save Changes" */
  isEditMode?: boolean;
  /** Custom submit button label */
  submitLabel?: string;
  /** Custom cancel button label */
  cancelLabel?: string;
  /** Additional CSS classes for the form element */
  className?: string;
  /** Whether to show action buttons */
  showActions?: boolean;
  /** React Hook Form mode for validation */
  mode?: UseFormProps<TFormValues>["mode"];
}

/**
 * React Hook Form wrapper with Zod validation.
 *
 * @example
 * ```tsx
 * const schema = z.object({
 *   name: z.string().min(1, "Name is required"),
 *   email: z.string().email("Invalid email"),
 * });
 *
 * <FormWrapper
 *   schema={schema}
 *   defaultValues={{ name: "", email: "" }}
 *   onSubmit={async (values) => {
 *     await saveData(values);
 *   }}
 * >
 *   {(form) => (
 *     <>
 *       <InputField control={form.control} name="name" label="Name" />
 *       <InputField control={form.control} name="email" label="Email" type="email" />
 *     </>
 *   )}
 * </FormWrapper>
 * ```
 */
export function FormWrapper<TFormValues extends FieldValues>({
  schema,
  defaultValues,
  onSubmit,
  onCancel,
  children,
  isLoading = false,
  isEditMode = false,
  submitLabel,
  cancelLabel = "Cancel",
  className,
  showActions = true,
  mode = "onBlur",
}: FormWrapperProps<TFormValues>) {
  const form = useForm<TFormValues>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(schema as any),
    defaultValues,
    mode,
  });

  const defaultSubmitLabel = isEditMode ? "Save Changes" : "Create";
  const { isSubmitting } = form.formState;

  return (
    <FormProvider {...form}>
      <form
        onSubmit={form.handleSubmit((values) => onSubmit(values, form))}
        className={className}
      >
        <FieldGroup>
          {typeof children === "function" ? children(form) : children}
        </FieldGroup>

        {showActions && (
          <div className="flex justify-end gap-2 pt-4">
            {onCancel && (
              <Button
                type="button"
                variant="outline"
                onClick={onCancel}
                disabled={isLoading || isSubmitting}
              >
                {cancelLabel}
              </Button>
            )}
            <Button
              className={`flex-1`}
              type="submit"
              disabled={isLoading || isSubmitting}
            >
              {isLoading || isSubmitting
                ? "Saving..."
                : submitLabel || defaultSubmitLabel}
            </Button>
          </div>
        )}
      </form>
    </FormProvider>
  );
}

/**
 * Hook to use existing form from context or create new one
 */
export { useForm, useFormContext } from "react-hook-form";
export type { UseFormReturn, Control, FieldValues } from "react-hook-form";
