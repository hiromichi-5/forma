import type { ReactNode } from "react";

interface ButtonProps {
  children: ReactNode;
  type?: "button" | "submit" | "reset";
  variant?: "primary" | "secondary" | "danger" | "ghost";
  size?: "sm" | "md" | "lg";
  disabled?: boolean;
  loading?: boolean;
  onClick?: () => void;
  className?: string;
  "aria-label"?: string;
}

export function Button({
  children,
  type = "button",
  variant = "primary",
  size = "md",
  disabled = false,
  loading = false,
  onClick,
  className = "",
  "aria-label": ariaLabel,
}: ButtonProps) {
  const baseClasses = [
    "inline-flex items-center justify-center",
    "rounded-md border transition-colors",
    "focus-ring disabled:opacity-50 disabled:cursor-not-allowed",
    "font-medium",
  ];

  const variantClasses = {
    primary: [
      "bg-primary-600 text-black border-primary-600",
      "hover:bg-primary-700 hover:border-primary-700",
      "active:bg-primary-800",
    ],
    secondary: [
      "bg-white text-gray-700 border-gray-300",
      "hover:bg-gray-50 hover:border-gray-400",
      "active:bg-gray-100",
    ],
    danger: [
      "bg-red-600 text-white border-red-600",
      "hover:bg-red-700 hover:border-red-700",
      "active:bg-red-800",
    ],
    ghost: [
      "bg-transparent text-gray-700 border-transparent",
      "hover:bg-gray-100",
      "active:bg-gray-200",
    ],
  };

  const sizeClasses = {
    sm: "px-3 py-1.5 text-sm",
    md: "px-4 py-2 text-sm",
    lg: "px-6 py-3 text-base",
  };

  const allClasses = [
    ...baseClasses,
    ...variantClasses[variant],
    sizeClasses[size],
    className,
  ].join(" ");

  return (
    <button
      type={type}
      disabled={disabled || loading}
      onClick={onClick}
      className={allClasses}
      aria-label={ariaLabel}
      aria-busy={loading}
    >
      {loading && (
        <svg
          className="animate-spin -ml-1 mr-2 h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  );
}
