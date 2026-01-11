"use client"

import { useEffect, useState } from "react"
import type { FormResponse } from "@/types/form-response"
import type { TicketAnswer, TicketDetail, TicketSummary } from "@/types"
import { apiClient } from "@/lib/api"

const buildResponsesMap = (answers: TicketAnswer[]): Record<string, string> =>
  answers.reduce<Record<string, string>>((acc, answer) => {
    acc[answer.question_id] = answer.display_value
    return acc
  }, {})

const buildQuestions = (answers: TicketAnswer[]): FormResponse["questions"] =>
  answers.map((answer) => ({
    questionId: answer.question_id,
    question: answer.question_title,
    answer: answer.display_value,
  }))

const normalizeRespondentEmail = (email: string | null): string =>
  email ?? "メールアドレス未登録"

const mapTicketDetailToFormResponse = (ticket: TicketDetail): FormResponse => {
  return {
    id: ticket.id,
    formId: ticket.form_id,
    formTitle: ticket.form_title,
    respondentEmail: normalizeRespondentEmail(ticket.respondent_email),
    submittedAt: new Date(ticket.submitted_at),
    status: ticket.status.id,
    statusName: ticket.status.name,
    statusColor: ticket.status.color,
    assignedTo: ticket.assignee?.id ?? null,
    responses: buildResponsesMap(ticket.answers),
    questions: buildQuestions(ticket.answers),
    priority: ticket.priority,
  }
}

const mapSummaryToFormResponse = (ticket: TicketSummary): FormResponse => {
  return {
    id: ticket.id,
    formId: ticket.form_id,
    formTitle: ticket.form_title,
    respondentEmail: normalizeRespondentEmail(ticket.respondent_email),
    submittedAt: new Date(ticket.submitted_at),
    status: ticket.status.id,
    statusName: ticket.status.name,
    statusColor: ticket.status.color,
    assignedTo: ticket.assignee?.id ?? null,
    responses: {},
    questions: [],
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

  const updateResponseStatus = async (id: string, statusId: string) => {
    try {
      const updated = await apiClient.updateTicket(id, { status_id: statusId })
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
