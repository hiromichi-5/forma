"use client"

import { useState } from "react"
import type { ChatMessage } from "@/types/form-response"

export function useChatMessages(responseId: string) {
  const [messages, setMessages] = useState<ChatMessage[]>([])

  const sendMessage = async (message: string, senderId: string, senderName: string, _respondentEmail?: string) => {
    const newMessage: ChatMessage = {
      id: `msg-${Date.now()}`,
      responseId,
      senderId,
      senderName,
      senderType: "staff",
      message,
      timestamp: new Date(),
      isRead: false,
    }
    setMessages((prev) => [...prev, newMessage])
  }

  const markAsRead = (messageId: string) => {
    setMessages((prev) => prev.map((m) => (m.id === messageId ? { ...m, isRead: true } : m)))
  }

  return {
    messages,
    sendMessage,
    markAsRead,
  }
}
