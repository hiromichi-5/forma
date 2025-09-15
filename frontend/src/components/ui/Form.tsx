import type { ReactNode } from 'react';

interface InputProps {
  id: string;
  type?: 'text' | 'email' | 'password' | 'number';
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
  required?: boolean;
  disabled?: boolean;
  error?: string;
  className?: string;
  'aria-describedby'?: string;
}

interface LabelProps {
  htmlFor: string;
  children: ReactNode;
  required?: boolean;
  className?: string;
}

export function Label({
  htmlFor,
  children,
  required = false,
  className = '',
}: LabelProps) {
  return (
    <label
      htmlFor={htmlFor}
      className={`block text-sm font-medium text-gray-700 ${className}`}
    >
      {children}
      {required && (
        <span className="text-red-500 ml-1" aria-label="必須">
          *
        </span>
      )}
    </label>
  );
}

export function Input({
  id,
  type = 'text',
  placeholder,
  value,
  onChange,
  onBlur,
  required = false,
  disabled = false,
  error,
  className = '',
  'aria-describedby': ariaDescribedBy,
}: InputProps) {
  const baseClasses = [
    'block w-full px-3 py-2 border rounded-md shadow-sm',
    'focus-ring placeholder-gray-400',
    'disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed',
  ];

  const errorClasses = error
    ? 'border-red-300 text-red-900 focus:border-red-500'
    : 'border-gray-300 focus:border-primary-500';

  const allClasses = [...baseClasses, errorClasses, className].join(' ');

  return (
    <div>
      <input
        id={id}
        type={type}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={onBlur}
        required={required}
        disabled={disabled}
        className={allClasses}
        aria-describedby={ariaDescribedBy}
        aria-invalid={error ? 'true' : 'false'}
      />
      {error && (
        <p
          id={`${id}-error`}
          className="mt-1 text-sm text-red-600"
          role="alert"
        >
          {error}
        </p>
      )}
    </div>
  );
}

interface FieldProps {
  label: string;
  id: string;
  required?: boolean;
  error?: string;
  children: ReactNode;
  className?: string;
}

export function Field({
  label,
  id,
  required = false,
  error,
  children,
  className = '',
}: FieldProps) {
  return (
    <div className={`space-y-1 ${className}`}>
      <Label htmlFor={id} required={required}>
        {label}
      </Label>
      {children}
      {error && (
        <p
          id={`${id}-error`}
          className="text-sm text-red-600"
          role="alert"
        >
          {error}
        </p>
      )}
    </div>
  );
}
