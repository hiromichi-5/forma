import type { FormStatus } from "@/types"
import type { FormResponse } from "@/types/form-response"

/** 優先度の表示ラベルと配色。一覧・カンバン・詳細で共通に使う。 */
export const priorityConfig = {
  low: { label: "低", color: "text-gray-600", hex: "#6B7280" },
  medium: { label: "中", color: "text-blue-600", hex: "#2563EB" },
  high: { label: "高", color: "text-red-600", hex: "#DC2626" },
}

/** ステータスに色が設定されていない場合に使う色。 */
export const fallbackStatusColor = "#9CA3AF"

const isPriorityValue = (value: string): value is FormResponse["priority"] =>
  value === "low" || value === "medium" || value === "high"

/** Select の値（string）を優先度に戻す。想定外の値は既定値の medium とする。 */
export const toPriorityValue = (value: string): FormResponse["priority"] =>
  isPriorityValue(value) ? value : "medium"

/** ステータス色から淡い背景色を作る。色が未設定ならグレーにフォールバックする。 */
export const hexToRgba = (hex: string | null | undefined, alpha: number): string => {
  if (!hex) return `rgba(156, 163, 175, ${alpha})`
  const cleanHex = hex.replace("#", "")
  const r = parseInt(cleanHex.slice(0, 2), 16)
  const g = parseInt(cleanHex.slice(2, 4), 16)
  const b = parseInt(cleanHex.slice(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

/** ステータスを表示順に並べ替える。 */
export const sortStatuses = (statuses: FormStatus[]): FormStatus[] =>
  [...statuses].sort((a, b) => a.display_order - b.display_order)
