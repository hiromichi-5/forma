import { useInfiniteQuery } from "@tanstack/react-query"
import { ticketQueries } from "@/queries/tickets"
import type { TicketListFilter } from "@/queries/keys"
import type { TicketSummary } from "@/types"

export type TicketListPage = {
  tickets: TicketSummary[]
  hasMore: boolean
  loading: boolean
  failed: boolean
  loadMore: () => void
  retry: () => void
}

/**
 * 条件に合うチケットの一覧（無限スクロール）。
 * 条件ごとに別のキャッシュになるため、条件を変えた後に古い条件の応答が届いても表示は混ざらない。
 */
export function useTicketList(formId: string, filter: TicketListFilter): TicketListPage {
  const { data, hasNextPage, isPending, isError, fetchNextPage, isFetchingNextPage, refetch } =
    useInfiniteQuery(ticketQueries.list(formId, filter))

  return {
    tickets: data?.pages.flatMap((page) => page.tickets) ?? [],
    hasMore: hasNextPage,
    // 画面に戻ったときなどの裏での取り直しでは、末尾の表示も続きの読み込みも止めない。
    loading: isPending || isFetchingNextPage,
    failed: isError,
    loadMore: () => {
      if (!isFetchingNextPage) void fetchNextPage()
    },
    retry: () => void (data ? fetchNextPage() : refetch()),
  }
}
