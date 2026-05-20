"use client"

import { useEffect, useRef } from "react"
import type { TicketDetail } from "@/types"

export function useTicketStream(
  formId: string | null,
  onTicketUpdated: (ticket: TicketDetail) => void
) {
  const callbackRef = useRef(onTicketUpdated)
  useEffect(() => {
    callbackRef.current = onTicketUpdated
  })

  useEffect(() => {
    if (!formId) return

    const baseUrl = import.meta.env.VITE_API_URL || "http://localhost:8080"
    const es = new EventSource(`${baseUrl}/v1/forms/${formId}/stream`, { withCredentials: true })

    es.addEventListener("ticket_updated", (e: MessageEvent) => {
      const { ticket } = JSON.parse(e.data) as { ticket: TicketDetail }
      callbackRef.current(ticket)
    })

    return () => es.close()
  }, [formId])
}
