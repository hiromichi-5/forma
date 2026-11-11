import type { InfiniteData, QueryClient, QueryKey } from "@tanstack/react-query"
import { queryKeys, type TicketListFilter } from "@/queries/keys"
import type {
  ListTicketCountsResponse,
  ListTicketsResponse,
  TicketDetail,
  TicketSummary,
} from "@/types"

type TicketPages = InfiniteData<ListTicketsResponse, string | null>

/** 一覧の並び順（提出日時の新しい順、同時刻は id の降順）。サーバーの並びと一致させる。 */
export const compareTickets = (a: TicketSummary, b: TicketSummary) =>
  b.submitted_at.localeCompare(a.submitted_at) || b.id.localeCompare(a.id)

const matchesQuery = (ticket: TicketSummary, query: string) =>
  !query || (ticket.respondent_email ?? "").toLowerCase().includes(query.toLowerCase())

export const matchesFilter = (ticket: TicketSummary, filter: TicketListFilter) =>
  (filter.statusIds.length === 0 || filter.statusIds.includes(ticket.status.id)) &&
  matchesQuery(ticket, filter.query)

const mapPages = (
  data: TicketPages,
  fn: (tickets: TicketSummary[], index: number) => TicketSummary[],
) => ({
  ...data,
  pages: data.pages.map((page, index) => ({ ...page, tickets: fn(page.tickets, index) })),
})

/**
 * 更新されたチケットを、読み込み済みの一覧に反映する。
 * - 条件に合わなくなったら取り除き、合っていれば内容を差し替える
 * - 新たに条件に合ったら並び順の位置に挿入する。ただし読み込み済みの範囲より後ろに並ぶものは、
 *   続きのページで取得されるため挿入しない（挿入すると、続きを読んだときに重複する）
 */
export function placeTicket(
  data: TicketPages,
  ticket: TicketSummary,
  filter: TicketListFilter,
): TicketPages {
  const included = data.pages.some((page) => page.tickets.some((t) => t.id === ticket.id))

  if (!matchesFilter(ticket, filter)) {
    return included ? mapPages(data, (tickets) => tickets.filter((t) => t.id !== ticket.id)) : data
  }
  if (included) {
    return mapPages(data, (tickets) => tickets.map((t) => (t.id === ticket.id ? ticket : t)))
  }

  const all = data.pages.flatMap((page) => page.tickets)
  const hasMore = data.pages[data.pages.length - 1]?.next_cursor != null
  const last = all[all.length - 1]
  if (hasMore && (!last || compareTickets(ticket, last) > 0)) return data

  // 挿入位置の直後にくるチケットを探し、そのページに入れる。見つからなければ最後のページの末尾。
  const next = all.find((t) => compareTickets(ticket, t) < 0)
  const lastPageIndex = data.pages.length - 1
  return mapPages(data, (tickets, index) => {
    if (next) {
      const at = tickets.findIndex((t) => t.id === next.id)
      return at === -1 ? tickets : [...tickets.slice(0, at), ticket, ...tickets.slice(at)]
    }
    return index === lastPageIndex ? [...tickets, ticket] : tickets
  })
}

/** 読み込み済みのキャッシュから、更新前のチケットを探す。 */
export function findCachedTicket(
  queryClient: QueryClient,
  formId: string,
  ticketId: string,
): TicketSummary | undefined {
  const detail = queryClient.getQueryData<TicketDetail>(queryKeys.ticket(ticketId))
  if (detail) return detail
  for (const [, data] of queryClient.getQueriesData<TicketPages>({
    queryKey: queryKeys.ticketLists(formId),
  })) {
    const found = data?.pages.flatMap((page) => page.tickets).find((t) => t.id === ticketId)
    if (found) return found
  }
  return undefined
}

/**
 * チケットの更新（自分の操作でも SSE で届いた他人の操作でも）を、関係するキャッシュすべてに反映する。
 * 一覧を取り直すとスクロール位置や読み込み済みのページが失われるため、手元で差し替える。
 */
export function applyTicketUpdate(queryClient: QueryClient, ticket: TicketDetail) {
  const formId = ticket.form_id
  const previous = findCachedTicket(queryClient, formId, ticket.id)

  for (const [key, data] of queryClient.getQueriesData<TicketPages>({
    queryKey: queryKeys.ticketLists(formId),
  })) {
    if (!data) continue
    const filter = key[key.length - 1] as TicketListFilter
    queryClient.setQueryData<TicketPages>(key, placeTicket(data, ticket, filter))
    refetchIfInFlight(queryClient, key)
  }

  for (const [key, data] of queryClient.getQueriesData<ListTicketCountsResponse>({
    queryKey: queryKeys.ticketCountsAll(formId),
  })) {
    if (!data) continue
    const { query } = key[key.length - 1] as { query: string }
    if (!previous) {
      // 新しいチケットは件数の増減を手元で判断できないため取り直す。
      void queryClient.invalidateQueries({ queryKey: key, exact: true })
      continue
    }
    if (previous.status.id === ticket.status.id || !matchesQuery(ticket, query)) continue
    queryClient.setQueryData<ListTicketCountsResponse>(key, {
      counts: adjustCounts(data.counts, previous.status.id, ticket.status.id),
    })
    refetchIfInFlight(queryClient, key)
  }

  queryClient.setQueryData(queryKeys.ticket(ticket.id), ticket)
  // 更新のたびに変更履歴が増えるため、詳細画面の履歴も取り直す。
  void queryClient.invalidateQueries({ queryKey: queryKeys.ticketHistories(ticket.id) })
}

// 反映前に始まった取り直しは更新前の状態を返しうるため、取り直しをやり直して上書きを防ぐ。
function refetchIfInFlight(queryClient: QueryClient, queryKey: QueryKey) {
  if (queryClient.isFetching({ queryKey, exact: true }) > 0) {
    void queryClient.invalidateQueries({ queryKey, exact: true })
  }
}

function adjustCounts(
  counts: ListTicketCountsResponse["counts"],
  from: string,
  to: string,
): ListTicketCountsResponse["counts"] {
  const next = counts.map((c) =>
    c.status_id === from
      ? { ...c, count: Math.max(c.count - 1, 0) }
      : c.status_id === to
        ? { ...c, count: c.count + 1 }
        : c,
  )
  return next.some((c) => c.status_id === to) ? next : [...next, { status_id: to, count: 1 }]
}
