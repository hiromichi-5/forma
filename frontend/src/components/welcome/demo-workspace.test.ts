import { describe, expect, it } from "vitest"
import type { InfiniteData } from "@tanstack/react-query"
import { createDemoQueryClient, createDemoWorkspace } from "@/components/welcome/demo-workspace"
import { demoStatusId } from "@/components/welcome/demo-data"
import { queryKeys } from "@/queries/keys"
import type { ListTicketCountsResponse, ListTicketsResponse, TicketDetail } from "@/types"

const FORM = "demo-inquiry"
const OPEN = demoStatusId(FORM, "open")
const DONE = demoStatusId(FORM, "done")

function setup() {
  const queryClient = createDemoQueryClient()
  const workspace = createDemoWorkspace(queryClient, new Date("2026-11-11T00:00:00Z"))
  const list = (statusIds: string[], query = "") =>
    queryClient
      .getQueryData<InfiniteData<ListTicketsResponse>>(
        queryKeys.ticketList(FORM, { statusIds, query }),
      )
      ?.pages.flatMap((page) => page.tickets) ?? []
  const count = (statusId: string, query = "") =>
    queryClient
      .getQueryData<ListTicketCountsResponse>(queryKeys.ticketCounts(FORM, query))
      ?.counts.find((c) => c.status_id === statusId)?.count
  return { queryClient, workspace, list, count }
}

describe("createDemoWorkspace", () => {
  it("ステータスを変えると、カンバンの列と件数の間でチケットが移る", () => {
    const { workspace, list, count } = setup()
    const ticket = list([OPEN])[0]
    const [openBefore, doneBefore] = [count(OPEN)!, count(DONE)!]

    workspace.updateTicket(FORM, ticket.id, { status_id: DONE })

    expect(list([OPEN]).map((t) => t.id)).not.toContain(ticket.id)
    expect(list([DONE]).map((t) => t.id)).toContain(ticket.id)
    expect(count(OPEN)).toBe(openBefore - 1)
    expect(count(DONE)).toBe(doneBefore + 1)
  })

  it("変更を詳細と変更履歴にも反映する", () => {
    const { queryClient, workspace, list } = setup()
    const ticket = list([])[0]

    workspace.updateTicket(FORM, ticket.id, { priority: "low" })

    expect(queryClient.getQueryData<TicketDetail>(queryKeys.ticket(ticket.id))?.priority).toBe(
      "low",
    )
    expect(
      queryClient.getQueryData<{ histories: unknown[] }>(queryKeys.ticketHistories(ticket.id))
        ?.histories,
    ).toHaveLength(1)
  })

  it("検索の条件を用意すると、条件に合う回答だけの一覧になる", () => {
    const { workspace, list } = setup()

    workspace.prepare(FORM, { statusIds: [], query: "yamada" })

    expect(list([], "yamada").map((t) => t.respondent_email)).toEqual(["yamada@example.com"])
  })

  it("同期すると届いた回答が先頭に加わり、2回目は何も届かない", () => {
    const { workspace, list } = setup()

    expect(workspace.sync(FORM)).toBe(1)
    expect(list([])[0].respondent_email).toBe("matsumoto@example.com")
    expect(workspace.sync(FORM)).toBe(0)
  })

  it("タイトル質問を変えると、チケットのタイトルがその回答になる", () => {
    const { workspace, list } = setup()

    workspace.setTitleQuestion(FORM, `${FORM}-q2`)

    expect(list([]).find((t) => t.respondent_email === "yamada@example.com")?.title).toBe(
      "山田 太郎",
    )
  })

  it("回答ごとのやりとりとメモを持ち、変更履歴は新しい順に並ぶ", () => {
    const { queryClient, workspace, list } = setup()
    const ticket = list([]).find((t) => t.respondent_email === "ito@example.com")!

    const histories = queryClient.getQueryData<{ histories: { created_at: string }[] }>(
      queryKeys.ticketHistories(ticket.id),
    )!.histories
    const chat = workspace.chat(FORM, ticket.id)

    expect(histories.length).toBeGreaterThan(0)
    expect(histories.map((h) => h.created_at)).toEqual(
      [...histories.map((h) => h.created_at)].sort().reverse(),
    )
    expect(chat.find((m) => m.from === "respondent")?.authorName).toBe("伊藤 大輔")
    expect(workspace.memo(FORM, ticket.id)).not.toBe("")
  })

  it("API から取得しない", () => {
    const { queryClient } = setup()

    expect(queryClient.getDefaultOptions().queries?.enabled).toBe(false)
  })
})
