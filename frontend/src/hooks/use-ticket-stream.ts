import { useEffect } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { API_BASE_URL } from "@/lib/api"
import { applyTicketUpdate } from "@/lib/ticket-cache"
import type { TicketUpdatedEvent } from "@/types"

/** 他のメンバーによるチケットの更新を SSE で受け取り、表示中の一覧と詳細に反映する。 */
export function useTicketStream(formId: string | null) {
  const queryClient = useQueryClient()

  useEffect(() => {
    if (!formId) return

    const es = new EventSource(`${API_BASE_URL}/v1/forms/${formId}/stream`, {
      withCredentials: true,
    })

    es.addEventListener("ticket_updated", (e: MessageEvent<string>) => {
      let event: TicketUpdatedEvent
      try {
        event = JSON.parse(e.data) as TicketUpdatedEvent
      } catch (error) {
        console.error("Failed to parse ticket_updated event:", error)
        return
      }
      applyTicketUpdate(queryClient, event.ticket)
    })

    return () => es.close()
  }, [formId, queryClient])
}
