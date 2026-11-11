import { describe, expect, it } from "vitest"
import { placeTicket } from "@/lib/ticket-cache"
import { ticketSummary } from "@/test/factories"
import type { ListTicketsResponse, TicketSummary } from "@/types"

const at = (day: number, overrides: Partial<TicketSummary> = {}) =>
  ticketSummary({ submitted_at: `2026-10-${String(day).padStart(2, "0")}T00:00:00Z`, ...overrides })

const pages = (...list: [TicketSummary[], string | null][]) => ({
  pages: list.map(([tickets, next_cursor]): ListTicketsResponse => ({ tickets, next_cursor })),
  pageParams: list.map((_, i) => (i === 0 ? null : `cursor-${i}`)),
})

const ids = (data: ReturnType<typeof pages>) => data.pages.map((p) => p.tickets.map((t) => t.id))

const all = { statusIds: [], query: "" }
const OPEN = { id: "open", name: "未対応" }
const DONE = { id: "done", name: "完了" }

describe("placeTicket", () => {
  it("条件に合わなくなったチケットを取り除く", () => {
    const t = at(5, { status: OPEN })
    const data = pages([[t], null])

    const next = placeTicket(data, { ...t, status: DONE }, { statusIds: [OPEN.id], query: "" })

    expect(ids(next)).toEqual([[]])
  })

  it("既に一覧にあるチケットは、位置を変えずに内容を差し替える", () => {
    const [a, b] = [at(9), at(8)]
    const next = placeTicket(pages([[a, b], null]), { ...b, priority: "high" }, all)

    expect(next.pages[0].tickets[1].priority).toBe("high")
    expect(ids(next)).toEqual([[a.id, b.id]])
  })

  it("新たに条件に合ったチケットを、並び順の位置に挿入する", () => {
    const [a, c] = [at(9), at(5)]
    const b = at(7)

    const next = placeTicket(pages([[a], "c1"], [[c], null]), b, all)

    expect(ids(next)).toEqual([[a.id], [b.id, c.id]])
  })

  it("読み込み済みの範囲より後ろに並ぶチケットは、続きのページで届くため挿入しない", () => {
    const a = at(9)
    const older = at(1)
    const data = pages([[a], "c1"])

    expect(placeTicket(data, older, all)).toBe(data)
  })

  it("続きのページが無ければ、末尾に追加する", () => {
    const a = at(9)
    const older = at(1)

    expect(ids(placeTicket(pages([[a], null]), older, all))).toEqual([[a.id, older.id]])
  })

  it("メールアドレスの検索条件に合わないチケットは挿入しない", () => {
    const data = pages([[], null])
    const t = at(3, { respondent_email: "foo@example.com" })

    expect(placeTicket(data, t, { statusIds: [], query: "bar" })).toBe(data)
  })
})
