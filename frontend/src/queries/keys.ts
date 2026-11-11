/** チケット一覧の絞り込み条件。キーの一部になるため、同じ条件は同じキャッシュを指す。 */
export type TicketListFilter = {
  /** 空配列は絞り込みなし（全てのステータス）を表す。 */
  statusIds: string[]
  query: string
}

/**
 * キャッシュのキー。階層にしておくと、上位のキーで関連するデータをまとめて無効化できる。
 * 例: invalidateQueries({ queryKey: queryKeys.form(id) }) でそのフォームの配下すべてを取り直す。
 */
export const queryKeys = {
  forms: () => ["forms"] as const,
  form: (formId: string) => ["forms", formId] as const,
  formQuestions: (formId: string) => ["forms", formId, "questions"] as const,
  formStatuses: (formId: string) => ["forms", formId, "statuses"] as const,
  formMembers: (formId: string) => ["forms", formId, "members"] as const,
  formInvites: (formId: string) => ["forms", formId, "invites"] as const,
  notificationSettings: (formId: string) => ["forms", formId, "notification-settings"] as const,
  /** フォームに属するチケットの一覧と件数すべて。SSE の更新はこの配下に反映する。 */
  formTickets: (formId: string) => ["forms", formId, "tickets"] as const,
  ticketLists: (formId: string) => ["forms", formId, "tickets", "list"] as const,
  ticketList: (formId: string, filter: TicketListFilter) =>
    ["forms", formId, "tickets", "list", filter] as const,
  ticketCountsAll: (formId: string) => ["forms", formId, "tickets", "counts"] as const,
  ticketCounts: (formId: string, query: string) =>
    ["forms", formId, "tickets", "counts", { query }] as const,
  ticket: (ticketId: string) => ["tickets", ticketId] as const,
  ticketHistories: (ticketId: string) => ["tickets", ticketId, "histories"] as const,
}
