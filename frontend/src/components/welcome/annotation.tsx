import type { CSSProperties, ReactNode } from "react"
import { cn } from "@/lib/utils"

type AnnotationProps = {
  children: ReactNode
  /** 矢印の向き。注釈の上下どちらに指す対象があるか。 */
  arrow: "up" | "down"
  /** 文字に対する矢印の横位置。 */
  align?: "start" | "center" | "end"
  className?: string
  style?: CSSProperties
}

/** 画面の要素を指す注釈。 */
export function Annotation({
  children,
  arrow,
  align = "center",
  className,
  style,
}: AnnotationProps) {
  return (
    <div
      className={cn(
        "pointer-events-none flex w-max flex-col gap-1 font-annotation text-[15px] leading-snug font-medium text-sky-600 select-none",
        arrow === "up" && "flex-col-reverse",
        align === "start" && "items-start",
        align === "center" && "items-center",
        align === "end" && "items-end",
        className,
      )}
      style={style}
    >
      <p className="whitespace-nowrap">{children}</p>
      <StraightArrow className={cn("mx-4 h-11 w-3 text-sky-500", arrow === "up" && "rotate-180")} />
    </div>
  )
}

function StraightArrow({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 12 44"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden="true"
    >
      <path d="M6 2v39" />
      <path d="M1.5 35.5 6 41l4.5-5.5" />
    </svg>
  )
}
