export const statusConfig = {
  new: {
    label: "新規",
    color: "bg-blue-100 text-blue-800",
    description: "未対応の新しい回答",
  },
  "in-review": {
    label: "対応中",
    color: "bg-yellow-100 text-yellow-800",
    description: "現在対応中",
  },
  "needs-info": {
    label: "情報待ち",
    color: "bg-orange-100 text-orange-800",
    description: "追加情報が必要",
  },
  completed: {
    label: "完了",
    color: "bg-green-100 text-green-800",
    description: "対応完了",
  },
} as const

export const priorityConfig = {
  low: {
    label: "低",
    color: "text-gray-600",
  },
  medium: {
    label: "中",
    color: "text-blue-600",
  },
  high: {
    label: "高",
    color: "text-red-600",
  },
} as const
