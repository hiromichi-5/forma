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
