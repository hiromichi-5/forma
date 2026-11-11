import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { History, Mail } from "lucide-react"
import { formatDistanceToNow, format } from "date-fns"
import { ja } from "date-fns/locale"
import { useTicketHistories } from "@/hooks/use-ticket-histories"
import {
  fallbackStatusColor,
  hexToRgba,
  priorityConfig,
  respondentEmailLabel,
  sortStatuses,
  statusById,
  toPriorityValue,
} from "@/lib/ticket-display"
import type {
  FormStatus,
  Member,
  NotificationType,
  TicketDetail,
  TicketHistory,
  TicketPriority,
  TicketSummary,
} from "@/types"

type ResponseDetailProps = {
  response: TicketSummary
  /** 回答本文と通知履歴。一覧には含まれないため、読み込みが終わるまで undefined。 */
  detail: TicketDetail | undefined
  open: boolean
  onOpenChange: (open: boolean) => void
  members: Member[]
  statuses: FormStatus[]
  onStatusChange: (id: string, statusId: string) => void
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: TicketPriority) => void
  onSendNotification?: (responseId: string, notificationType: NotificationType) => Promise<boolean>
}

export function ResponseDetail({
  response,
  detail,
  open,
  onOpenChange,
  members,
  statuses,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onSendNotification,
}: ResponseDetailProps) {
  const sortedStatuses = sortStatuses(statuses)
  const email = respondentEmailLabel(response.respondent_email)
  const status = statusById(statuses).get(response.status.id) ?? response.status
  const [sendingNotification, setSendingNotification] = useState<NotificationType | null>(null)

  const handleSendNotification = async (notificationType: NotificationType) => {
    if (!onSendNotification) return
    setSendingNotification(notificationType)
    try {
      await onSendNotification(response.id, notificationType)
    } finally {
      setSendingNotification(null)
    }
  }

  const getLastSentAt = (notificationType: NotificationType): string | null =>
    detail?.notifications.find((n) => n.notification_type === notificationType)?.last_sent_at ??
    null

  const formatLastSentAt = (sentAt: string | null): string =>
    sentAt ? `最終送信: ${format(new Date(sentAt), "MM/dd HH:mm")}` : "未送信"
  const {
    histories,
    isLoading: isHistoriesLoading,
    error: historiesError,
  } = useTicketHistories(response.id)

  const getFieldLabel = (fieldName: string): string => {
    switch (fieldName) {
      case "status":
        return "ステータス"
      case "assignee":
        return "担当者"
      case "priority":
        return "優先度"
      default:
        return fieldName
    }
  }

  const getPriorityLabel = (priority: string): string =>
    priorityConfig[priority as keyof typeof priorityConfig]?.label ?? priority

  const formatHistoryChange = (history: TicketHistory): string => {
    const fieldLabel = getFieldLabel(history.field_name)
    const oldValue = history.old_value
      ? history.field_name === "priority"
        ? getPriorityLabel(history.old_value)
        : history.old_value
      : "なし"
    const newValue = history.new_value
      ? history.field_name === "priority"
        ? getPriorityLabel(history.new_value)
        : history.new_value
      : "なし"

    return `${fieldLabel}を「${oldValue}」から「${newValue}」に変更しました`
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex h-[90vh] max-w-7xl flex-col gap-0 overflow-hidden p-0 sm:max-w-7xl">
        <div className="border-b p-4 pr-12">
          <DialogHeader className="flex-row items-center justify-between gap-3 text-left">
            <DialogTitle>{email}の詳細</DialogTitle>
            {onSendNotification && response.respondent_email && (
              <div className="flex items-center gap-3">
                <div className="flex flex-col items-center gap-1">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleSendNotification("status_change")}
                    disabled={sendingNotification !== null}
                  >
                    <Mail className="h-4 w-4" />
                    対応状況を通知
                  </Button>
                  <span className="text-xs whitespace-nowrap text-muted-foreground">
                    {formatLastSentAt(getLastSentAt("status_change"))}
                  </span>
                </div>
                <div className="flex flex-col items-center gap-1">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleSendNotification("assignee_assigned")}
                    disabled={sendingNotification !== null || !response.assignee}
                  >
                    <Mail className="h-4 w-4" />
                    担当者を通知
                  </Button>
                  <span className="text-xs whitespace-nowrap text-muted-foreground">
                    {formatLastSentAt(getLastSentAt("assignee_assigned"))}
                  </span>
                </div>
              </div>
            )}
          </DialogHeader>

          <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2">
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-muted-foreground">ステータス</span>
              <Select
                value={response.status.id}
                onValueChange={(value) => onStatusChange(response.id, value)}
              >
                <SelectTrigger
                  className="h-8 w-[150px] border-0 shadow-none"
                  aria-label="ステータス"
                >
                  <div
                    className="flex items-center gap-2 rounded px-2 py-1"
                    style={{
                      backgroundColor: hexToRgba(status.color, 0.1),
                    }}
                  >
                    <div
                      className="h-2 w-2 shrink-0 rounded-full"
                      style={{
                        backgroundColor: status.color || fallbackStatusColor,
                      }}
                    />
                    <span className="text-sm">{status.name}</span>
                  </div>
                </SelectTrigger>
                <SelectContent>
                  {sortedStatuses.map((status) => (
                    <SelectItem key={status.id} value={status.id}>
                      {status.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center gap-1.5">
              <span className="text-xs text-muted-foreground">担当者</span>
              <Select
                value={response.assignee?.id ?? "unassigned"}
                onValueChange={(value) =>
                  onAssignChange(response.id, value === "unassigned" ? null : value)
                }
              >
                <SelectTrigger className="h-8 w-[150px] border-0 shadow-none" aria-label="担当者">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="unassigned">未割当</SelectItem>
                  {members.map((member) => (
                    <SelectItem key={member.id} value={member.id}>
                      {member.display_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center gap-1.5">
              <span className="text-xs text-muted-foreground">優先度</span>
              <Select
                value={response.priority}
                onValueChange={(value) => onPriorityChange(response.id, toPriorityValue(value))}
              >
                <SelectTrigger className="h-8 w-[90px] border-0 shadow-none" aria-label="優先度">
                  <div
                    className="flex items-center gap-2 rounded px-2 py-1"
                    style={{
                      backgroundColor: hexToRgba(priorityConfig[response.priority].hex, 0.1),
                    }}
                  >
                    <span
                      className="text-sm font-medium"
                      style={{ color: priorityConfig[response.priority].hex }}
                    >
                      {priorityConfig[response.priority].label}
                    </span>
                  </div>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">低</SelectItem>
                  <SelectItem value="medium">中</SelectItem>
                  <SelectItem value="high">高</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        <div className="flex flex-1 overflow-hidden">
          <div className="flex w-1/2 flex-col border-r">
            <div className="flex-1 overflow-y-auto p-4">
              <div className="mb-3 flex items-center gap-2">
                <Badge variant="outline">元の回答</Badge>
                <span className="text-xs text-muted-foreground">
                  {formatDistanceToNow(new Date(response.submitted_at), {
                    addSuffix: true,
                    locale: ja,
                  })}
                </span>
                <span className="text-xs text-muted-foreground">
                  ({format(new Date(response.submitted_at), "yyyy/MM/dd HH:mm")})
                </span>
              </div>
              <div className="space-y-3">
                {detail ? (
                  detail.answers.map((answer, index) => (
                    <div
                      key={`${answer.question_id}-${index}`}
                      className="rounded-lg bg-muted/50 p-3"
                    >
                      <p className="mb-1 text-xs font-medium text-muted-foreground">
                        {answer.question_title}
                      </p>
                      <p className="text-sm">{answer.display_value}</p>
                    </div>
                  ))
                ) : (
                  <p className="text-sm text-muted-foreground">読み込み中...</p>
                )}
              </div>
            </div>
          </div>

          <div className="flex w-1/2 flex-col overflow-y-auto p-4">
            <div className="mb-3 flex items-center gap-2">
              <Badge variant="outline">対応履歴</Badge>
            </div>
            {isHistoriesLoading ? (
              <p className="py-8 text-center text-sm text-muted-foreground">読み込み中...</p>
            ) : historiesError ? (
              <p className="py-8 text-center text-sm text-muted-foreground">
                変更履歴を読み込めませんでした
              </p>
            ) : histories.length === 0 ? (
              <div className="flex flex-1 flex-col items-center justify-center text-center">
                <History className="mb-3 h-12 w-12 text-muted-foreground" />
                <p className="text-muted-foreground">まだ変更履歴がありません</p>
                <p className="text-sm text-muted-foreground">
                  ステータス・担当者・優先度を変更すると、ここに記録されます
                </p>
              </div>
            ) : (
              <ol className="space-y-3">
                {histories.map((history) => (
                  <li
                    key={history.id}
                    className="rounded-lg border border-muted bg-muted/30 px-4 py-2"
                  >
                    <div className="mb-1 flex items-center gap-2 text-xs text-muted-foreground">
                      <span className="font-medium">{history.changed_by_name}</span>
                      <span>
                        {formatDistanceToNow(new Date(history.created_at), {
                          addSuffix: true,
                          locale: ja,
                        })}
                      </span>
                      <span>({format(new Date(history.created_at), "yyyy/MM/dd HH:mm")})</span>
                    </div>
                    <p className="text-sm">{formatHistoryChange(history)}</p>
                  </li>
                ))}
              </ol>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
