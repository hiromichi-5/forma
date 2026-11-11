import type { TicketDetail, TicketSummary } from "@/types"

let seq = 0

/** テスト用のチケット。必要な項目だけ上書きして使う。 */
export function ticketSummary(overrides: Partial<TicketSummary> = {}): TicketSummary {
  seq += 1
  return {
    id: `ticket-${seq}`,
    form_id: "form-1",
    form_title: "フォーム",
    response_id: `response-${seq}`,
    respondent_email: `user${seq}@example.com`,
    status: { id: "status-open", name: "未対応", color: "#EF4444" },
    priority: "medium",
    title: `回答 ${seq}`,
    assignee: null,
    submitted_at: `2026-10-${String(seq).padStart(2, "0")}T00:00:00Z`,
    created_at: "2026-10-01T00:00:00Z",
    ...overrides,
  }
}

export function ticketDetail(overrides: Partial<TicketDetail> = {}): TicketDetail {
  return { ...ticketSummary(overrides), answers: [], notifications: [], ...overrides }
}
