"use client"

import { useEffect, useState } from "react"
import type { FormResponse } from "@/types/form-response"
import type { TicketAnswer, TicketDetail, TicketSummary } from "@/types"
import { apiClient } from "@/lib/api"
import { mockTicketRespondents } from "@/lib/mock-data"

type TicketRespondent = {
  name: string
  email: string
}

const buildResponsesMap = (answers: TicketAnswer[]): Record<string, string> =>
  answers.reduce<Record<string, string>>((acc, answer) => {
    acc[answer.question_id] = answer.display_value
    return acc
  }, {})

const fallbackRespondents: TicketRespondent[] = [
  { name: "山田 次郎", email: "yamada@example.com" },
  { name: "佐藤 花子", email: "sato@example.com" },
  { name: "鈴木 一郎", email: "suzuki@example.com" },
  { name: "高橋 健太", email: "takahashi@example.com" },
  { name: "伊藤 美咲", email: "ito@example.com" },
]

const pickFallbackRespondent = (ticketId: string): TicketRespondent => {
  const total = Array.from(ticketId).reduce((sum, ch) => sum + ch.charCodeAt(0), 0)
  const index = total % fallbackRespondents.length
  return fallbackRespondents[index]
}

const getRespondentInfo = (ticketId: string): TicketRespondent =>
  mockTicketRespondents[ticketId] ?? pickFallbackRespondent(ticketId)

const mapTicketDetailToFormResponse = (ticket: TicketDetail): FormResponse => {
  const respondent = getRespondentInfo(ticket.id)
  return {
    id: ticket.id,
    formId: ticket.form_id,
    formTitle: ticket.form_title,
    respondentEmail: respondent.email,
    respondentName: respondent.name,
    submittedAt: new Date(ticket.submitted_at),
    status: ticket.status,
    assignedTo: ticket.assignee?.id ?? null,
    responses: buildResponsesMap(ticket.answers),
    priority: ticket.priority,
  }
}

const mapSummaryToFormResponse = (ticket: TicketSummary): FormResponse => {
  const respondent = getRespondentInfo(ticket.id)
  return {
    id: ticket.id,
    formId: ticket.form_id,
    formTitle: ticket.form_title,
    respondentEmail: respondent.email,
    respondentName: respondent.name,
    submittedAt: new Date(ticket.submitted_at),
    status: ticket.status,
    assignedTo: ticket.assignee?.id ?? null,
    responses: {},
    priority: ticket.priority,
  }
}

export function useFormResponses(formId: string | null) {
  const [responses, setResponses] = useState<FormResponse[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!formId) return
    let isActive = true

    const loadResponses = async () => {
      setLoading(true)
      try {
        // 外部API(バックエンド)との同期のための処理
        const ticketList = await apiClient.getTickets(formId)
        if (!isActive) return
        const baseResponses = ticketList.tickets.map(mapSummaryToFormResponse)
        setResponses(baseResponses)

        const details = await Promise.all(
          ticketList.tickets.map((ticket) => apiClient.getTicket(ticket.id))
        )
        if (!isActive) return
        setResponses(details.map(mapTicketDetailToFormResponse))
      } catch (error) {
        if (!isActive) return
        console.error("Failed to load responses:", error)
        setResponses([])
      } finally {
        if (isActive) {
          setLoading(false)
        }
      }
    }

    // 外部API(バックエンド)との同期のための処理
    loadResponses()

    return () => {
      isActive = false
    }
  }, [formId])

  const updateResponseStatus = async (id: string, status: FormResponse["status"]) => {
    try {
      const updated = await apiClient.updateTicket(id, { status })
      setResponses((prev) => prev.map((r) => (r.id === id ? mapTicketDetailToFormResponse(updated) : r)))
    } catch (error) {
      console.error("Failed to update status:", error)
    }
  }

  const assignResponse = async (id: string, userId: string | null) => {
    try {
      const updated = await apiClient.updateTicket(id, { assignee_id: userId })
      setResponses((prev) => prev.map((r) => (r.id === id ? mapTicketDetailToFormResponse(updated) : r)))
    } catch (error) {
      console.error("Failed to update assignee:", error)
    }
  }

  const updatePriority = async (id: string, priority: FormResponse["priority"]) => {
    try {
      const updated = await apiClient.updateTicket(id, { priority })
      setResponses((prev) => prev.map((r) => (r.id === id ? mapTicketDetailToFormResponse(updated) : r)))
    } catch (error) {
      console.error("Failed to update priority:", error)
    }
  }

  return {
    responses,
    loading,
    updateResponseStatus,
    assignResponse,
    updatePriority,
  }
}
