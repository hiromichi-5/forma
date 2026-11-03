import { useInfiniteQuery } from "@tanstack/react-query"
import { ticketQueries } from "@/queries/tickets"
import type { TicketListFilter } from "@/queries/keys"
import type { TicketSummary } from "@/types"

export type TicketListPage = {
  tickets: TicketSummary[]
  hasMore: boolean
  loading: boolean
  loadMore: () => void
}

/**
 * 条件に合うチケットの一覧（無限スクロール）。
 * 条件ごとに別のキャッシュになるため、条件を変えた後に古い条件の応答が届いても表示は混ざらない。
 */
export function useTicketList(formId: string, filter: TicketListFilter): TicketListPage {
  const { data, hasNextPage, isFetching, isError, fetchNextPage, isFetchingNextPage } =
    useInfiniteQuery(ticketQueries.list(formId, filter))

  return {
    tickets: data?.pages.flatMap((page) => page.tickets) ?? [],
    // 読み込みに失敗した場合は自動で続きを読み込み続けないよう、続きが無いものとして扱う。
    hasMore: hasNextPage && !isError,
    loading: isFetching,
    loadMore: () => {
      if (!isFetchingNextPage) void fetchNextPage()
    },
  }
}
