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
  ticketHistories: (ticketId: string) => ["tickets", ticketId, "histories"] as const,
}
