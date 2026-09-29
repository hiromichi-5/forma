"use client"

import { useCallback, useEffect, useMemo, useRef, useState } from "react"
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

const PAGE_SIZE = 50

export type TicketSegment = {
  key: string
  statusIds: string[]
}

export type SegmentPage = {
  tickets: TicketSummary[]
  hasMore: boolean
  loading: boolean
}

type SegmentState = {
  ids: string[]
  nextCursor: string | null
  loading: boolean
  failed: boolean
}

type ListState = {
  byId: Record<string, TicketSummary>
  segments: Record<string, SegmentState>
}

const EMPTY_LIST: ListState = { byId: {}, segments: {} }
const EMPTY_SEGMENT: SegmentState = { ids: [], nextCursor: null, loading: false, failed: false }
const LOADING_PAGE: SegmentPage = { tickets: [], hasMore: false, loading: true }

const compareTickets = (a: TicketSummary, b: TicketSummary) =>
  b.submitted_at.localeCompare(a.submitted_at) || b.id.localeCompare(a.id)

const matchesQuery = (ticket: TicketSummary, query: string) =>
  !query || (ticket.respondent_email ?? "").toLowerCase().includes(query.toLowerCase())

const matchesSegment = (ticket: TicketSummary, segment: TicketSegment, query: string) =>
  (segment.statusIds.length === 0 || segment.statusIds.includes(ticket.status.id)) &&
  matchesQuery(ticket, query)

// 読み込み済みの範囲より後ろに並ぶチケットは、続きのページで取得されるため挿入しない。
function placeTicket(
  state: SegmentState,
  ticket: TicketSummary,
  byId: Record<string, TicketSummary>,
  matches: boolean
): SegmentState {
  const included = state.ids.includes(ticket.id)
  if (!matches) {
    return included ? { ...state, ids: state.ids.filter((id) => id !== ticket.id) } : state
  }
  if (included) return state
  if (state.loading && state.ids.length === 0) return state

  const last = byId[state.ids[state.ids.length - 1]]
  if (state.nextCursor !== null && (!last || compareTickets(ticket, last) > 0)) return state

  const ids = [...state.ids]
  const index = ids.findIndex((id) => compareTickets(ticket, byId[id]) < 0)
  ids.splice(index === -1 ? ids.length : index, 0, ticket.id)
  return { ...state, ids }
}

type UseFormResponsesOptions = {
  segments: TicketSegment[]
  query: string
  withCounts: boolean
}

export function useFormResponses(
  formId: string | null,
  { segments, query, withCounts }: UseFormResponsesOptions
) {
  const [list, setList] = useState<ListState>(EMPTY_LIST)
  const [counts, setCounts] = useState<Record<string, number>>({})
  const [details, setDetails] = useState<Record<string, TicketDetail>>({})
  const [pendingNotification, setPendingNotification] = useState<PendingNotification | null>(null)
  const { modeOf } = useNotificationSettings(formId)

  const segmentsKey = segments.map((s) => `${s.key}:${s.statusIds.join(",")}`).join("|")
  const contextRef = useRef({ formId, segments, query, withCounts })
  const listRef = useRef(list)
  useEffect(() => {
    contextRef.current = { formId, segments, query, withCounts }
    listRef.current = list
  })

  // 条件を変えた後に古い条件の応答が届いても反映しないよう、読み込みの世代を区別する。
  const generationRef = useRef(0)
  const inFlightRef = useRef(new Set<string>())

  const loadPage = useCallback(async (segment: TicketSegment, cursor: string | null) => {
    const { formId, query } = contextRef.current
    if (!formId || inFlightRef.current.has(segment.key)) return
    const generation = generationRef.current
    inFlightRef.current.add(segment.key)

    const updateSegment = (update: (current: SegmentState) => SegmentState) =>
      setList((prev) => ({
        ...prev,
        segments: {
          ...prev.segments,
          [segment.key]: update(prev.segments[segment.key] ?? EMPTY_SEGMENT),
        },
      }))

    updateSegment((current) => ({ ...current, loading: true, failed: false }))
    try {
      const res = await apiClient.getTickets(formId, {
        statusIds: segment.statusIds,
        query,
        limit: PAGE_SIZE,
        cursor,
      })
      if (generation !== generationRef.current) return
      setList((prev) => {
        const byId = { ...prev.byId }
        res.tickets.forEach((ticket) => {
          byId[ticket.id] = ticket
        })
        const current = prev.segments[segment.key] ?? EMPTY_SEGMENT
        const known = new Set(current.ids)
        const ids = [...current.ids, ...res.tickets.map((t) => t.id).filter((id) => !known.has(id))]
        return {
          byId,
          segments: {
            ...prev.segments,
            [segment.key]: { ids, nextCursor: res.next_cursor, loading: false, failed: false },
          },
        }
      })
    } catch (error) {
      if (generation !== generationRef.current) return
      console.error("Failed to load responses:", error)
      updateSegment((current) => ({ ...current, loading: false, failed: true }))
    } finally {
      if (generation === generationRef.current) inFlightRef.current.delete(segment.key)
    }
  }, [])

  const loadCounts = useCallback(async () => {
    const { formId, query } = contextRef.current
    if (!formId) return
    const generation = generationRef.current
    try {
      const res = await apiClient.getTicketCounts(formId, query)
      if (generation !== generationRef.current) return
      setCounts(Object.fromEntries(res.counts.map((c) => [c.status_id, c.count])))
    } catch (error) {
      console.error("Failed to load ticket counts:", error)
    }
  }, [])

  const refetch = useCallback(() => {
    generationRef.current += 1
    inFlightRef.current = new Set()
    setList(EMPTY_LIST)
    setCounts({})
    setDetails({})
    const { segments, withCounts } = contextRef.current
    segments.forEach((segment) => loadPage(segment, null))
    if (withCounts) loadCounts()
  }, [loadPage, loadCounts])

  useEffect(() => {
    refetch()
  }, [formId, segmentsKey, query, withCounts, refetch])

  const loadMore = useCallback(
    (key: string) => {
      const segment = contextRef.current.segments.find((s) => s.key === key)
      const state = listRef.current.segments[key]
      if (!segment || !state || state.nextCursor === null) return
      loadPage(segment, state.nextCursor)
    },
    [loadPage]
  )

  const applyTicket = useCallback(
    (ticket: TicketSummary) => {
      const { segments, query, withCounts } = contextRef.current
      const previous = listRef.current.byId[ticket.id]

      setList((prev) => {
        const byId = { ...prev.byId, [ticket.id]: ticket }
        const next: Record<string, SegmentState> = { ...prev.segments }
        segments.forEach((segment) => {
          const state = prev.segments[segment.key]
          if (!state) return
          next[segment.key] = placeTicket(state, ticket, byId, matchesSegment(ticket, segment, query))
        })
        return { byId, segments: next }
      })

      if (!withCounts) return
      if (!previous) {
        loadCounts()
        return
      }
      if (previous.status.id !== ticket.status.id && matchesQuery(ticket, query)) {
        setCounts((prev) => ({
          ...prev,
          [previous.status.id]: (prev[previous.status.id] ?? 1) - 1,
          [ticket.status.id]: (prev[ticket.status.id] ?? 0) + 1,
        }))
      }
    },
    [loadCounts]
  )

  // 詳細は一覧の要約も含むため、両方に反映して表示のずれを防ぐ。
  const cacheDetail = useCallback(
    (detail: TicketDetail) => {
      applyTicket(detail)
      setDetails((prev) => ({ ...prev, [detail.id]: detail }))
    },
    [applyTicket]
  )

  useTicketStream(formId, cacheDetail)

  const pages = useMemo(() => {
    const result: Record<string, SegmentPage> = {}
    segments.forEach((segment) => {
      const state = list.segments[segment.key]
      result[segment.key] = state
        ? {
            tickets: state.ids.map((id) => list.byId[id]),
            hasMore: state.nextCursor !== null && !state.failed,
            loading: state.loading,
          }
        : LOADING_PAGE
    })
    return result
  }, [list, segments])

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
    return Boolean(list.byId[id]?.respondent_email)
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
    pages,
    ticketsById: list.byId,
    counts,
    loadMore,
    details,
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
