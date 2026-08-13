"use client"

import { useCallback, useEffect, useState } from "react"
import type { NotificationMode, NotificationSetting, NotificationType } from "@/types"
import { apiClient } from "@/lib/api"

const NOTIFICATION_TYPES: NotificationType[] = ["status_change", "assignee_assigned"]

const defaultSettings = (): NotificationSetting[] =>
  NOTIFICATION_TYPES.map((notification_type) => ({
    notification_type,
    mode: "off" as NotificationMode,
    include_detail: false,
  }))

export function useNotificationSettings(formId: string | null) {
  const [settings, setSettings] = useState<NotificationSetting[]>(defaultSettings)
  const [emailCollectionType, setEmailCollectionType] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const refetch = useCallback(async () => {
    if (!formId) return
    setLoading(true)
    try {
      const res = await apiClient.getNotificationSettings(formId)
      setSettings(res.settings)
      setEmailCollectionType(res.email_collection_type ?? null)
    } catch (error) {
      // 通知設定を取得できない場合は既定値（すべて off）として扱い、確認ダイアログを出さない。
      console.error("Failed to load notification settings:", error)
      setSettings(defaultSettings())
    } finally {
      setLoading(false)
    }
  }, [formId])

  useEffect(() => {
    refetch()
  }, [refetch])

  const modeOf = useCallback(
    (notificationType: NotificationType): NotificationMode =>
      settings.find((s) => s.notification_type === notificationType)?.mode ?? "off",
    [settings]
  )

  return {
    settings,
    setSettings,
    emailCollectionType,
    loading,
    refetch,
    modeOf,
  }
}

/** メールアドレスを収集していないフォームでは通知が届かない。 */
export function isEmailCollectionDisabled(emailCollectionType: string | null): boolean {
  return emailCollectionType === "DO_NOT_COLLECT"
}
