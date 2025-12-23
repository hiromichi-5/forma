type StatusKey = "new" | "in_progress" | "done"
type StatusConfig = {
  label: string
  color: string
  description: string
}

type PriorityKey = "Low" | "Medium" | "High"
type PriorityConfig = {
  label: string
  color: string
}

export const statusConfig: Record<StatusKey, StatusConfig> = {
  new: {
    label: "新規",
    color: "bg-blue-100 text-blue-800",
    description: "未対応の新しい回答",
  },
  in_progress: {
    label: "対応中",
    color: "bg-yellow-100 text-yellow-800",
    description: "現在対応中",
  },
  done: {
    label: "完了",
    color: "bg-green-100 text-green-800",
    description: "対応完了",
  },
}

export const priorityConfig: Record<PriorityKey, PriorityConfig> = {
  Low: {
    label: "低",
    color: "text-gray-600",
  },
  Medium: {
    label: "中",
    color: "text-blue-600",
  },
  High: {
    label: "高",
    color: "text-red-600",
  },
}
