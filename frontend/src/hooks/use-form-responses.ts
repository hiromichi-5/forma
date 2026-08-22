"use client"

import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"
import type {
  NotificationResult,
  NotificationType,
  TicketDetail,
  TicketPriority,
  TicketSummary,
  TicketUpdateResponse,
  UpdateTicketRequest,
} from "@/types"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import { useTicketStream } from "@/hooks/use-ticket-stream"
import { useNotificationSettings } from "@/hooks/use-notification-settings"

/** confirm モードで、変更の実行前に操作者へ通知の可否を確認する対象。 */
export type PendingNotification = {
  notificationType: NotificationType
  responseId: string
  request: UpdateTicketRequest
}

const NOTIFICATION_LABELS: Record<NotificationType, string> = {
  status_change: "対応状況の変更",
  assignee_assigned: "担当者の割り当て",
}

export const notificationLabel = (type: NotificationType): string =>
  NOTIFICATION_LABELS[type]

export function useFormResponses(formId: string | null) {
  const [responses, setResponses] = useState<TicketSummary[]>([])
  const [details, setDetails] = useState<Record<string, TicketDetail>>({})
  const [loading, setLoading] = useState(false)
  const [pendingNotification, setPendingNotification] = useState<PendingNotification | null>(null)
  const { modeOf } = useNotificationSettings(formId)

  // 詳細は一覧の要約も含むため、両方に反映して表示のずれを防ぐ。
  const cacheDetail = useCallback((detail: TicketDetail) => {
    setResponses((prev) => prev.map((r) => (r.id === detail.id ? detail : r)))
    setDetails((prev) => ({ ...prev, [detail.id]: detail }))
  }, [])

  useTicketStream(formId, cacheDetail)

  // 回答本文と通知履歴は一覧に含まれないため、チケットを開いた時点で取得する。
  const loadDetail = useCallback(
    async (id: string) => {
      try {
        cacheDetail(await apiClient.getTicket(id))
      } catch (error) {
        console.error("Failed to load ticket detail:", error)
      }
    },
    [cacheDetail]
  )

  const refetch = useCallback(async () => {
    if (!formId) return
    setLoading(true)
    setDetails({})
    try {
      const ticketList = await apiClient.getTickets(formId)
      setResponses(ticketList.tickets)
    } catch (error) {
      console.error("Failed to load responses:", error)
      setResponses([])
    } finally {
      setLoading(false)
    }
  }, [formId])

  useEffect(() => {
    refetch()
  }, [refetch])

  // 自動送信（always）の失敗はチケット更新自体を妨げないため、警告として伝えるにとどめる。
  const warnFailedNotifications = (results: NotificationResult[]) => {
    results
      .filter((r) => r.result === "failed")
      .forEach((r) => {
        toast.warning("回答者への通知メールの送信に失敗しました", {
          description: `${notificationLabel(r.notification_type)}は保存されています。`,
        })
      })
  }

  const applyUpdate = (updated: TicketUpdateResponse) => {
    cacheDetail(updated)
    warnFailedNotifications(updated.notification_results)
  }

  const updateTicket = async (
    id: string,
    request: UpdateTicketRequest,
    failureMessage: string
  ): Promise<boolean> => {
    try {
      applyUpdate(await apiClient.updateTicket(id, request))
      return true
    } catch (error) {
      console.error(failureMessage, error)
      toast.error(getApiErrorMessage(error, {}, failureMessage))
      return false
    }
  }

  // confirm モードかつ回答者のメールアドレスがある場合のみ、変更前に確認する。
  const needsConfirmation = (id: string, notificationType: NotificationType): boolean => {
    if (modeOf(notificationType) !== "confirm") return false
    return Boolean(responses.find((r) => r.id === id)?.respondent_email)
  }

  const updateResponseStatus = async (id: string, statusId: string) => {
    const request: UpdateTicketRequest = { status_id: statusId }
    if (needsConfirmation(id, "status_change")) {
      setPendingNotification({
        notificationType: "status_change",
        responseId: id,
        request,
      })
      return
    }
    await updateTicket(id, request, "対応状況の変更に失敗しました")
  }

  const assignResponse = async (id: string, userId: string | null) => {
    const request: UpdateTicketRequest = { assignee_id: userId }
    // 担当者の解除は通知の対象外。
    if (userId !== null && needsConfirmation(id, "assignee_assigned")) {
      setPendingNotification({
        notificationType: "assignee_assigned",
        responseId: id,
        request,
      })
      return
    }
    await updateTicket(id, request, "担当者の変更に失敗しました")
  }

  const updatePriority = async (id: string, priority: TicketPriority) => {
    await updateTicket(id, { priority }, "優先度の変更に失敗しました")
  }

  // confirm モードでの送信と、詳細画面からの再送の両方で使う。
  const sendNotification = async (id: string, notificationType: NotificationType) => {
    try {
      await apiClient.sendTicketNotification(id, { notification_type: notificationType })
      toast.success("回答者に通知メールを送信しました")
      return true
    } catch (error) {
      console.error("Failed to send notification:", error)
      toast.error(
        getApiErrorMessage(
          error,
          {
            NOTIFICATION_RATE_LIMITED:
              "直前に通知を送信しています。しばらくしてから再度お試しください",
            NOTIFICATION_DISABLED: "この通知は無効に設定されています",
            RESPONDENT_EMAIL_MISSING: "回答者のメールアドレスが登録されていません",
          },
          "通知メールの送信に失敗しました"
        )
      )
      return false
    }
  }

  const resolvePendingNotification = async (notify: boolean) => {
    const pending = pendingNotification
    setPendingNotification(null)
    if (!pending) return

    const updated = await updateTicket(
      pending.responseId,
      pending.request,
      "変更に失敗しました"
    )
    if (!updated || !notify) return

    await sendNotification(pending.responseId, pending.notificationType)
  }

  const cancelPendingNotification = () => setPendingNotification(null)

  return {
    responses,
    details,
    loading,
    loadDetail,
    updateResponseStatus,
    assignResponse,
    updatePriority,
    refetch,
    pendingNotification,
    resolvePendingNotification,
    cancelPendingNotification,
    sendNotification,
  }
}
