import { infiniteQueryOptions, queryOptions } from "@tanstack/react-query"
import { apiClient } from "@/lib/api"
import { queryKeys, type TicketListFilter } from "@/queries/keys"

export const TICKET_PAGE_SIZE = 50

export const ticketQueries = {
  /** カーソル方式のページング。続きは next_cursor で取得する。 */
  list: (formId: string, filter: TicketListFilter) =>
    infiniteQueryOptions({
      queryKey: queryKeys.ticketList(formId, filter),
      queryFn: ({ pageParam }) =>
        apiClient.getTickets(formId, {
          statusIds: filter.statusIds,
          query: filter.query,
          limit: TICKET_PAGE_SIZE,
          cursor: pageParam,
        }),
      initialPageParam: null as string | null,
      getNextPageParam: (lastPage) => lastPage.next_cursor ?? null,
    }),
  counts: (formId: string, query: string) =>
    queryOptions({
      queryKey: queryKeys.ticketCounts(formId, query),
      queryFn: () => apiClient.getTicketCounts(formId, query),
      select: (res): Record<string, number> =>
        Object.fromEntries(res.counts.map((c) => [c.status_id, c.count])),
    }),
  /** 回答本文と通知履歴。一覧には含まれないため、詳細を開いた時点で取得する。 */
  detail: (ticketId: string) =>
    queryOptions({
      queryKey: queryKeys.ticket(ticketId),
      queryFn: () => apiClient.getTicket(ticketId),
    }),
  histories: (ticketId: string) =>
    queryOptions({
      queryKey: queryKeys.ticketHistories(ticketId),
      queryFn: () => apiClient.getTicketHistories(ticketId),
      select: (res) => res.histories,
    }),
}
