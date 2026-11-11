import { act, renderHook, waitFor } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { apiClient } from "@/lib/api"
import { useTicketActions } from "@/hooks/use-ticket-actions"
import { useTicketList } from "@/hooks/use-ticket-list"
import { useTicketStream } from "@/hooks/use-ticket-stream"
import { ticketQueries } from "@/queries/tickets"
import { ticketDetail, ticketSummary } from "@/test/factories"
import { createQueryWrapper } from "@/test/query"
import type { ListTicketsResponse, TicketSummary } from "@/types"

/** SSE の購読先。テストからイベントを流し込む。 */
class FakeEventSource {
  static instances: FakeEventSource[] = []
  listeners = new Map<string, (e: MessageEvent) => void>()
  constructor() {
    FakeEventSource.instances.push(this)
  }
  addEventListener(type: string, listener: (e: MessageEvent) => void) {
    this.listeners.set(type, listener)
  }
  close() {}
  emit(type: string, data: unknown) {
    this.listeners.get(type)?.(new MessageEvent(type, { data: JSON.stringify(data) }))
  }
}

const FORM = "form-1"
const OPEN = { id: "status-open", name: "未対応" }
const DONE = { id: "status-done", name: "完了" }

const page = (
  tickets: TicketSummary[],
  next_cursor: string | null = null,
): ListTicketsResponse => ({
  tickets,
  next_cursor,
})

beforeEach(() => {
  FakeEventSource.instances = []
  vi.stubGlobal("EventSource", FakeEventSource)
  vi.spyOn(console, "error").mockImplementation(() => undefined)
})

describe("useTicketList", () => {
  it("最初のページを読み込み、loadMore で続きを末尾に追加する", async () => {
    const first = [ticketSummary(), ticketSummary()]
    const second = [ticketSummary()]
    const getTickets = vi
      .spyOn(apiClient, "getTickets")
      .mockResolvedValueOnce(page(first, "cursor-2"))
      .mockResolvedValueOnce(page(second))

    const { result } = renderHook(
      () => useTicketList(FORM, { statusIds: [], query: "" }),
      createQueryWrapper(),
    )

    await waitFor(() => expect(result.current.tickets).toEqual(first))
    expect(result.current.hasMore).toBe(true)

    act(() => result.current.loadMore())

    await waitFor(() => expect(result.current.tickets).toHaveLength(3))
    expect(getTickets).toHaveBeenLastCalledWith(
      FORM,
      expect.objectContaining({ cursor: "cursor-2" }),
    )
    expect(result.current.hasMore).toBe(false)
  })

  it("続きの読み込みに失敗しても、再試行で読み込みを再開できる", async () => {
    const first = [ticketSummary()]
    const second = [ticketSummary()]
    vi.spyOn(apiClient, "getTickets")
      .mockResolvedValueOnce(page(first, "cursor-2"))
      .mockRejectedValueOnce(new Error("network"))
      .mockResolvedValueOnce(page(second))

    const { result } = renderHook(
      () => useTicketList(FORM, { statusIds: [], query: "" }),
      createQueryWrapper(),
    )
    await waitFor(() => expect(result.current.tickets).toEqual(first))

    act(() => result.current.loadMore())
    await waitFor(() => expect(result.current.failed).toBe(true))
    expect(result.current.hasMore).toBe(true)

    act(() => result.current.retry())
    await waitFor(() => expect(result.current.tickets).toHaveLength(2))
    expect(result.current.failed).toBe(false)
  })

  it("裏での取り直しの間は読み込み中として扱わない", async () => {
    vi.spyOn(apiClient, "getTickets")
      .mockResolvedValueOnce(page([ticketSummary()], "cursor-2"))
      .mockImplementationOnce(() => new Promise(() => undefined))
    const { client, wrapper } = createQueryWrapper()

    const { result } = renderHook(() => useTicketList(FORM, { statusIds: [], query: "" }), {
      wrapper,
    })
    await waitFor(() => expect(result.current.tickets).toHaveLength(1))

    void client.invalidateQueries()
    await waitFor(() => expect(client.isFetching()).toBe(1))
    expect(result.current.loading).toBe(false)
  })

  it("検索条件が変わる前の応答は、後から届いても反映しない", async () => {
    let resolveStale: (value: ListTicketsResponse) => void = () => undefined
    const stale = ticketSummary({ respondent_email: "old@example.com" })
    const fresh = ticketSummary({ respondent_email: "new@example.com" })
    vi.spyOn(apiClient, "getTickets")
      .mockImplementationOnce(() => new Promise((resolve) => (resolveStale = resolve)))
      .mockResolvedValueOnce(page([fresh]))

    const { result, rerender } = renderHook(
      ({ query }) => useTicketList(FORM, { statusIds: [], query }),
      { initialProps: { query: "old" }, ...createQueryWrapper() },
    )
    rerender({ query: "new" })
    await waitFor(() => expect(result.current.tickets).toEqual([fresh]))

    await act(() => {
      resolveStale(page([stale]))
      return Promise.resolve()
    })

    expect(result.current.tickets).toEqual([fresh])
  })
})

describe("useTicketStream", () => {
  it("SSE で届いた更新に合わせて、カンバンの列と件数を移す", async () => {
    const ticket = ticketSummary({ status: OPEN })
    vi.spyOn(apiClient, "getTickets").mockImplementation((_formId, options) =>
      Promise.resolve(page(options?.statusIds?.[0] === OPEN.id ? [ticket] : [])),
    )
    vi.spyOn(apiClient, "getTicketCounts").mockResolvedValue({
      counts: [
        { status_id: OPEN.id, count: 1 },
        { status_id: DONE.id, count: 0 },
      ],
    })
    const { client, wrapper } = createQueryWrapper()

    const { result } = renderHook(
      () => {
        useTicketStream(FORM)
        return {
          open: useTicketList(FORM, { statusIds: [OPEN.id], query: "" }),
          done: useTicketList(FORM, { statusIds: [DONE.id], query: "" }),
        }
      },
      { wrapper },
    )
    await client.fetchQuery(ticketQueries.counts(FORM, ""))
    await waitFor(() => expect(result.current.open.tickets).toEqual([ticket]))

    act(() => {
      FakeEventSource.instances[0].emit("ticket_updated", {
        ticket: ticketDetail({ ...ticket, status: DONE }),
      })
    })

    // Query は変更の通知をまとめて行うため、再描画を待つ。
    await waitFor(() => expect(result.current.open.tickets).toEqual([]))
    expect(result.current.done.tickets.map((t) => t.id)).toEqual([ticket.id])
    expect(client.getQueryData(ticketQueries.counts(FORM, "").queryKey)).toEqual({
      counts: [
        { status_id: OPEN.id, count: 0 },
        { status_id: DONE.id, count: 1 },
      ],
    })
  })
})

describe("useTicketActions", () => {
  it("通知が confirm モードなら、変更の前に確認を挟み、承認後に更新と通知を行う", async () => {
    const ticket = ticketSummary({ status: OPEN })
    vi.spyOn(apiClient, "getTickets").mockResolvedValue(page([ticket]))
    const notificationSettings = vi.spyOn(apiClient, "getNotificationSettings").mockResolvedValue({
      settings: [{ notification_type: "status_change", mode: "confirm", include_detail: false }],
      email_collection_type: "VERIFIED",
    })
    const updateTicket = vi
      .spyOn(apiClient, "updateTicket")
      .mockResolvedValue({ ...ticketDetail({ ...ticket, status: DONE }), notification_results: [] })
    const sendNotification = vi
      .spyOn(apiClient, "sendTicketNotification")
      .mockResolvedValue({ notification_type: "status_change", sent_at: "2026-11-01T00:00:00Z" })

    const { result } = renderHook(
      () => ({
        list: useTicketList(FORM, { statusIds: [], query: "" }),
        actions: useTicketActions(FORM),
      }),
      createQueryWrapper(),
    )
    await waitFor(() => expect(result.current.list.tickets).toHaveLength(1))
    await waitFor(() => expect(notificationSettings).toHaveBeenCalled())

    await act(() => result.current.actions.updateResponseStatus(ticket.id, DONE.id))

    expect(updateTicket).not.toHaveBeenCalled()
    expect(result.current.actions.pendingNotification?.notificationType).toBe("status_change")

    await act(() => result.current.actions.resolvePendingNotification(true))

    expect(updateTicket).toHaveBeenCalledWith(ticket.id, { status_id: DONE.id })
    expect(sendNotification).toHaveBeenCalledWith(ticket.id, { notification_type: "status_change" })
    expect(result.current.actions.pendingNotification).toBeNull()
    // 変更結果は取り直さずに一覧へ反映される。
    await waitFor(() => expect(result.current.list.tickets[0].status).toEqual(DONE))
  })
})
