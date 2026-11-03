import { useEffect, useRef } from "react"
import { API_BASE_URL } from "@/lib/api"
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

    const es = new EventSource(`${API_BASE_URL}/v1/forms/${formId}/stream`, {
      withCredentials: true,
    })

    es.addEventListener("ticket_updated", (e: MessageEvent) => {
      const { ticket } = JSON.parse(e.data as string) as TicketUpdatedEvent
      callbackRef.current(ticket)
    })

    return () => es.close()
  }, [formId])
}
