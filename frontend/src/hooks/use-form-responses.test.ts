import { act, renderHook, waitFor } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { apiClient } from "@/lib/api"
import { useFormResponses, type TicketSegment } from "@/hooks/use-form-responses"
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
  vi.spyOn(apiClient, "getNotificationSettings").mockResolvedValue({
    settings: [
      { notification_type: "status_change", mode: "off", include_detail: false },
      { notification_type: "assignee_assigned", mode: "off", include_detail: false },
    ],
    email_collection_type: "VERIFIED",
  })
  vi.spyOn(console, "error").mockImplementation(() => undefined)
})

const listSegments: TicketSegment[] = [{ key: "list", statusIds: [] }]
const kanbanSegments: TicketSegment[] = [
  { key: OPEN.id, statusIds: [OPEN.id] },
  { key: DONE.id, statusIds: [DONE.id] },
]

describe("useFormResponses", () => {
  it("最初のページを読み込み、loadMore で続きを末尾に追加する", async () => {
    const first = [ticketSummary(), ticketSummary()]
    const second = [ticketSummary()]
    const getTickets = vi
      .spyOn(apiClient, "getTickets")
      .mockResolvedValueOnce(page(first, "cursor-2"))
      .mockResolvedValueOnce(page(second))

    const { result } = renderHook(
      () => useFormResponses("form-1", { segments: listSegments, query: "", withCounts: false }),
      createQueryWrapper(),
    )

    await waitFor(() => expect(result.current.pages.list.loading).toBe(false))
    expect(result.current.pages.list.tickets).toEqual(first)
    expect(result.current.pages.list.hasMore).toBe(true)

    act(() => result.current.loadMore("list"))

    await waitFor(() => expect(result.current.pages.list.tickets).toHaveLength(3))
    expect(getTickets).toHaveBeenLastCalledWith(
      "form-1",
      expect.objectContaining({ cursor: "cursor-2" }),
    )
    expect(result.current.pages.list.hasMore).toBe(false)
  })

  it("検索条件が変わる前の応答は、後から届いても反映しない", async () => {
    let resolveStale: (value: ListTicketsResponse) => void = () => undefined
    const stale = ticketSummary({ respondent_email: "old@example.com" })
    const fresh = ticketSummary({ respondent_email: "new@example.com" })
    vi.spyOn(apiClient, "getTickets")
      .mockImplementationOnce(() => new Promise((resolve) => (resolveStale = resolve)))
      .mockResolvedValueOnce(page([fresh]))

    const { result, rerender } = renderHook(
      ({ query }) =>
        useFormResponses("form-1", { segments: listSegments, query, withCounts: false }),
      { initialProps: { query: "old" }, ...createQueryWrapper() },
    )
    rerender({ query: "new" })
    await waitFor(() => expect(result.current.pages.list.tickets).toEqual([fresh]))

    await act(() => {
      resolveStale(page([stale]))
      return Promise.resolve()
    })

    expect(result.current.pages.list.tickets).toEqual([fresh])
  })

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

    const { result } = renderHook(
      () => useFormResponses("form-1", { segments: kanbanSegments, query: "", withCounts: true }),
      createQueryWrapper(),
    )
    await waitFor(() => expect(result.current.pages[OPEN.id].tickets).toEqual([ticket]))
    await waitFor(() => expect(result.current.counts[OPEN.id]).toBe(1))

    act(() => {
      FakeEventSource.instances[0].emit("ticket_updated", {
        ticket: ticketDetail({ ...ticket, status: DONE }),
      })
    })

    expect(result.current.pages[OPEN.id].tickets).toEqual([])
    expect(result.current.pages[DONE.id].tickets.map((t) => t.id)).toEqual([ticket.id])
    expect(result.current.counts).toEqual({ [OPEN.id]: 0, [DONE.id]: 1 })
  })

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
      () => useFormResponses("form-1", { segments: listSegments, query: "", withCounts: false }),
      createQueryWrapper(),
    )
    await waitFor(() => expect(result.current.pages.list.tickets).toHaveLength(1))
    await waitFor(() => expect(notificationSettings).toHaveBeenCalled())

    await act(() => result.current.updateResponseStatus(ticket.id, DONE.id))

    expect(updateTicket).not.toHaveBeenCalled()
    expect(result.current.pendingNotification?.notificationType).toBe("status_change")

    await act(() => result.current.resolvePendingNotification(true))

    expect(updateTicket).toHaveBeenCalledWith(ticket.id, { status_id: DONE.id })
    expect(sendNotification).toHaveBeenCalledWith(ticket.id, { notification_type: "status_change" })
    expect(result.current.pendingNotification).toBeNull()
  })
})
