import type { LucideIcon } from "lucide-react";

interface IconProps {
  icon: LucideIcon;
  size?: "sm" | "md" | "lg";
  className?: string;
}

export function Icon({ icon: IconComponent, size = "md", className = "" }: IconProps) {
  const sizeClasses: Record<NonNullable<IconProps["size"]>, string> = {
    sm: "h-4 w-4",
    md: "h-5 w-5",
    lg: "h-6 w-6",
  };

  return <IconComponent className={`${sizeClasses[size]} ${className}`} />;
}
