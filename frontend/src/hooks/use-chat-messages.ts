"use client"

import { useState } from "react"
import type { ChatMessage } from "@/types/form-response"

export function useChatMessages(responseId: string) {
  const [messages, setMessages] = useState<ChatMessage[]>([])

  const sendMessage = async (message: string, senderId: string, senderName: string, respondentEmail?: string) => {
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

    if (respondentEmail) {
      try {
        await fetch("/api/send-notification", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            to: respondentEmail,
            subject: "フォーム回答に関する追加のご質問",
            message: `${senderName}から新しいメッセージがあります:\n\n${message}`,
            fromName: senderName,
          }),
        })
        console.log("[v0] Email notification sent to:", respondentEmail)
      } catch (error) {
        console.error("[v0] Failed to send email notification:", error)
      }
    }
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
