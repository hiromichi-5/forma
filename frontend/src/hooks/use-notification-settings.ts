import { useCallback } from "react"
import { useQuery } from "@tanstack/react-query"
import type { NotificationMode, NotificationType } from "@/types"
import { formQueries } from "@/queries/forms"

/**
 * 通知の送信モード。通知設定ダイアログと同じキャッシュを読むため、設定を保存するとすぐに反映される。
 * 取得できない間は既定値（すべて off）として扱い、確認ダイアログを出さない。
 */
export function useNotificationSettings(formId: string | null) {
  const { data } = useQuery({
    ...formQueries.notificationSettings(formId ?? ""),
    enabled: Boolean(formId),
  })

  const modeOf = useCallback(
    (notificationType: NotificationType): NotificationMode =>
      data?.settings.find((s) => s.notification_type === notificationType)?.mode ?? "off",
    [data],
  )

  return { modeOf }
}

/** メールアドレスを収集していないフォームでは通知が届かない。 */
export function isEmailCollectionDisabled(emailCollectionType: string | null): boolean {
  return emailCollectionType === "DO_NOT_COLLECT"
}
