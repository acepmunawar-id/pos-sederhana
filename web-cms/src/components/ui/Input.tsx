import { forwardRef, useId, type InputHTMLAttributes, type ReactNode, type TextareaHTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

const fieldBase =
  'w-full rounded-md bg-fill px-3 text-subheadline text-text placeholder:text-text-secondary transition ' +
  'focus:outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60';

interface FieldProps {
  label?: string;
  error?: string;
  hint?: string;
  required?: boolean;
}

/** Label + control + error/hint wrapper used by every form control. */
export function Field({
  label,
  error,
  hint,
  required,
  htmlFor,
  children,
  className,
}: FieldProps & { htmlFor?: string; children: ReactNode; className?: string }) {
  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      {label && (
        <label htmlFor={htmlFor} className="text-footnote font-medium text-text-secondary">
          {label}
          {required && <span className="ml-0.5 text-danger">*</span>}
        </label>
      )}
      {children}
      {error ? (
        <p className="text-caption text-danger" role="alert">
          {error}
        </p>
      ) : hint ? (
        <p className="text-caption text-text-secondary">{hint}</p>
      ) : null}
    </div>
  );
}

export interface InputProps extends InputHTMLAttributes<HTMLInputElement>, FieldProps {
  leftIcon?: ReactNode;
  prefix?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { label, error, hint, required, leftIcon, prefix, className, id, ...rest },
  ref,
) {
  const autoId = useId();
  const inputId = id ?? autoId;
  return (
    <Field label={label} error={error} hint={hint} required={required} htmlFor={inputId} className={className}>
      <div className="relative">
        {leftIcon && (
          <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray">{leftIcon}</span>
        )}
        {prefix && (
          <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-subheadline text-text-secondary">
            {prefix}
          </span>
        )}
        <input
          ref={ref}
          id={inputId}
          aria-invalid={!!error}
          className={cn(fieldBase, 'h-touch', (leftIcon || prefix) && 'pl-10', error && 'ring-2 ring-danger/50')}
          {...rest}
        />
      </div>
    </Field>
  );
});

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement>, FieldProps {}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(function Textarea(
  { label, error, hint, required, className, id, rows = 3, ...rest },
  ref,
) {
  const autoId = useId();
  const inputId = id ?? autoId;
  return (
    <Field label={label} error={error} hint={hint} required={required} htmlFor={inputId} className={className}>
      <textarea
        ref={ref}
        id={inputId}
        rows={rows}
        aria-invalid={!!error}
        className={cn(fieldBase, 'resize-none py-2.5', error && 'ring-2 ring-danger/50')}
        {...rest}
      />
    </Field>
  );
});
