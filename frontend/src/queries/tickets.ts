import { queryOptions } from "@tanstack/react-query"
import { apiClient } from "@/lib/api"
import { queryKeys } from "@/queries/keys"

export const ticketQueries = {
  histories: (ticketId: string) =>
    queryOptions({
      queryKey: queryKeys.ticketHistories(ticketId),
      queryFn: () => apiClient.getTicketHistories(ticketId),
      select: (res) => res.histories,
    }),
}
