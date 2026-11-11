import { useQuery } from "@tanstack/react-query"
import type { TicketHistory } from "@/types"
import { ticketQueries } from "@/queries/tickets"

const EMPTY_HISTORIES: TicketHistory[] = []

export function useTicketHistories(ticketId: string) {
  const {
    data: histories = EMPTY_HISTORIES,
    isPending,
    error,
  } = useQuery(ticketQueries.histories(ticketId))
  return { histories, isLoading: isPending, error }
}
