export type FormResponse = {
  id: string
  formId: string
  formTitle: string
  respondentEmail: string
  respondentName: string
  submittedAt: Date
  status: "new" | "in_progress" | "done"
  assignedTo: string | null
  responses: Record<string, string>
  priority: "High" | "Medium" | "Low"
}

export type FormQuestion = {
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
  avatar?: string
}
