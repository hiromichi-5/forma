import { useEffect, useRef } from "react"
import type { TicketDetail, TicketUpdatedEvent } from "@/types"

export function useTicketStream(
  formId: string | null,
  onTicketUpdated: (ticket: TicketDetail) => void,
) {
  const callbackRef = useRef(onTicketUpdated)
  useEffect(() => {
    callbackRef.current = onTicketUpdated
  })

  useEffect(() => {
    if (!formId) return

    const baseUrl = import.meta.env.VITE_API_URL || "http://localhost:8080"
    const es = new EventSource(`${baseUrl}/v1/forms/${formId}/stream`, { withCredentials: true })

    es.addEventListener("ticket_updated", (e: MessageEvent<string>) => {
      let event: TicketUpdatedEvent
      try {
        event = JSON.parse(e.data) as TicketUpdatedEvent
      } catch (error) {
        console.error("Failed to parse ticket_updated event:", error)
        return
      }
      callbackRef.current(event.ticket)
    })

    return () => es.close()
  }, [formId])
}
