export interface FormResponse {
  id: string
  formId: string
  formTitle: string
  respondentEmail: string
  respondentName: string
  submittedAt: Date
  status: "new" | "in-review" | "needs-info" | "completed"
  assignedTo: string | null
  responses: Record<string, string>
  priority: "low" | "medium" | "high"
}

export interface FormQuestion {
  questionId: string
  question: string
  answer: string
}

export interface ChatMessage {
  id: string
  responseId: string
  senderId: string
  senderName: string
  senderType: "staff" | "respondent"
  message: string
  timestamp: Date
  isRead: boolean
}

export interface User {
  id: string
  name: string
  email: string
  avatar?: string
}
