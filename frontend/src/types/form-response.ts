export type FormResponse = {
  id: string
  formId: string
  formTitle: string
  respondentEmail: string
  hasRespondentEmail: boolean
  submittedAt: Date
  status: string
  statusName: string
  statusColor?: string | null
  assignedTo: string | null
  responses: Record<string, string>
  questions: FormResponseQuestion[]
  priority: "high" | "medium" | "low"
  notifications: NotificationStatus[]
}

export type NotificationStatus = {
  notificationType: "status_change" | "assignee_assigned"
  lastSentAt: Date | null
}

export type FormResponseQuestion = {
  questionId: string
  question: string
  answer: string
}

export type ChatMessage = {
  id: string
  responseId: string
  senderId: string
  senderName: string
  senderType: "staff" | "respondent"
  message: string
  timestamp: Date
  isRead: boolean
}

export type User = {
  id: string
  name: string
  email: string
}
