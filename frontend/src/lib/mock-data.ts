import type { ChatMessage } from "@/types/form-response"

export const mockTicketRespondents: Record<string, { name: string; email: string }> = {
  "mock-ticket-1": { name: "山田 次郎", email: "customer1@example.com" },
  "mock-ticket-2": { name: "伊藤 美咲", email: "customer2@example.com" },
  "mock-ticket-3": { name: "高橋 健太", email: "customer3@example.com" },
  "mock-ticket-4": { name: "中村 直美", email: "customer4@example.com" },
  "mock-ticket-5": { name: "小林 誠", email: "customer5@example.com" },
}

export const mockChatMessages: ChatMessage[] = [
  {
    id: "msg-1",
    responseId: "resp-2",
    senderId: "1",
    senderName: "田中 太郎",
    senderType: "staff",
    message: "お問い合わせありがとうございます。製品Bの導入事例について、業種を教えていただけますか？",
    timestamp: new Date("2024-01-14T15:00:00"),
    isRead: true,
  },
  {
    id: "msg-2",
    responseId: "resp-2",
    senderId: "customer2",
    senderName: "伊藤 美咲",
    senderType: "respondent",
    message: "IT業界です。特にSaaS企業の事例があれば教えてください。",
    timestamp: new Date("2024-01-14T15:30:00"),
    isRead: true,
  },
  {
    id: "msg-3",
    responseId: "resp-3",
    senderId: "2",
    senderName: "佐藤 花子",
    senderType: "staff",
    message: "ご不便をおかけしております。登録されているメールアドレスを確認させていただけますか？",
    timestamp: new Date("2024-01-13T10:00:00"),
    isRead: true,
  },
  {
    id: "msg-4",
    responseId: "resp-3",
    senderId: "customer3",
    senderName: "高橋 健太",
    senderType: "respondent",
    message: "customer3@example.comです。",
    timestamp: new Date("2024-01-13T10:15:00"),
    isRead: false,
  },
]
