import { queryOptions } from "@tanstack/react-query"
import { apiClient } from "@/lib/api"
import { queryKeys } from "@/queries/keys"

/**
 * 画面から使う取得処理の定義。useQuery(formQueries.list()) のように使う。
 * キーと取得関数を一緒に定義しておくことで、同じデータを別のキーで二重に持つことを防ぐ。
 */
export const formQueries = {
  list: () =>
    queryOptions({
      queryKey: queryKeys.forms(),
      queryFn: () => apiClient.getForms(),
      select: (res) => res.forms,
    }),
  detail: (formId: string) =>
    queryOptions({
      queryKey: queryKeys.form(formId),
      queryFn: () => apiClient.getForm(formId),
    }),
  questions: (formId: string) =>
    queryOptions({
      queryKey: queryKeys.formQuestions(formId),
      queryFn: () => apiClient.getFormQuestions(formId),
      select: (res) => res.questions,
    }),
  statuses: (formId: string) =>
    queryOptions({
      queryKey: queryKeys.formStatuses(formId),
      queryFn: () => apiClient.getFormStatuses(formId),
      select: (res) => res.statuses,
    }),
  members: (formId: string) =>
    queryOptions({
      queryKey: queryKeys.formMembers(formId),
      queryFn: () => apiClient.getMembers(formId),
      select: (res) => res.members,
    }),
  invites: (formId: string) =>
    queryOptions({
      queryKey: queryKeys.formInvites(formId),
      queryFn: () => apiClient.listInvites(formId),
      select: (res) => res.invites,
    }),
  notificationSettings: (formId: string) =>
    queryOptions({
      queryKey: queryKeys.notificationSettings(formId),
      queryFn: () => apiClient.getNotificationSettings(formId),
    }),
}
