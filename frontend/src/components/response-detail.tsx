import { useEffect, useMemo, useRef, useState, type FormEvent, type ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { ticketQueries } from "@/queries/tickets"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { History, Mail, Send } from "lucide-react"
import { formatDistanceToNow, format } from "date-fns"
import { ja } from "date-fns/locale"
import { useTicketHistories } from "@/hooks/use-ticket-histories"
import { cn } from "@/lib/utils"
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
  TicketHistory,
  TicketPriority,
  TicketSummary,
} from "@/types"

type ResponseDetailProps = {
  /** 開いた時点の一覧の値。詳細の取得が終わるまでの表示に使う。 */
  response: TicketSummary
  open: boolean
  onOpenChange: (open: boolean) => void
  members: Member[]
  statuses: FormStatus[]
  onStatusChange: (id: string, statusId: string) => void
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: TicketPriority) => void
  onSendNotification?: (responseId: string, notificationType: NotificationType) => Promise<boolean>
  /** 回答者とのやりとり。渡すと対応履歴と時系列に混ぜて表示し、送信欄を出す。 */
  chat?: ResponseChat
  /** 元の回答の下に表示する、チーム内のメモ欄。 */
  memo?: ReactNode
}

export type ResponseChatMessage = {
  id: string
  from: "staff" | "respondent"
  authorName: string
  text: string
  sentAt: string
}

export type ResponseChat = {
  messages: ResponseChatMessage[]
  onSend: (text: string) => void
  /** 送信欄の下に添える補足。 */
  footnote?: ReactNode
}

export function ResponseDetail({
  response: initialResponse,
  open,
  onOpenChange,
  members,
  statuses,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onSendNotification,
  chat,
  memo,
}: ResponseDetailProps) {
  // 回答本文と通知履歴は一覧に含まれないため、開いた時点で取得する。
  // 変更や SSE による更新もこのキャッシュに反映されるため、表示は常に最新になる。
  const { data: detail } = useQuery(ticketQueries.detail(initialResponse.id))
  const response = detail ?? initialResponse
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
            {memo && <div className="flex h-1/3 flex-col border-t p-4">{memo}</div>}
          </div>

          {chat ? (
            <ChatTimeline chat={chat} histories={histories} describe={formatHistoryChange} />
          ) : (
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
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

type ChatTimelineProps = {
  chat: ResponseChat
  histories: TicketHistory[]
  describe: (history: TicketHistory) => string
}

type TimelineItem =
  | { kind: "message"; at: number; message: ResponseChatMessage }
  | { kind: "history"; at: number; history: TicketHistory }

/** 回答者とのやりとりと対応履歴を、古い順に一続きで表示する。 */
function ChatTimeline({ chat, histories, describe }: ChatTimelineProps) {
  const [draft, setDraft] = useState("")
  const endRef = useRef<HTMLDivElement>(null)
  const items = useMemo(
    () =>
      [
        ...chat.messages.map((message): TimelineItem => ({
          kind: "message",
          at: new Date(message.sentAt).getTime(),
          message,
        })),
        ...histories.map((history): TimelineItem => ({
          kind: "history",
          at: new Date(history.created_at).getTime(),
          history,
        })),
      ].sort((a, b) => a.at - b.at),
    [chat.messages, histories],
  )

  useEffect(() => {
    endRef.current?.scrollIntoView({ block: "end" })
  }, [items.length])

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    const text = draft.trim()
    if (!text) return
    chat.onSend(text)
    setDraft("")
  }

  return (
    <div className="flex w-1/2 flex-col">
      <div className="flex-1 space-y-4 overflow-y-auto p-4">
        {items.map((item) =>
          item.kind === "message" ? (
            <div
              key={`message-${item.message.id}`}
              className={cn(
                "flex flex-col gap-1",
                item.message.from === "staff" ? "items-end" : "items-start",
              )}
            >
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span className="font-medium">{item.message.authorName}</span>
                <span>{formatDistanceToNow(item.at, { addSuffix: true, locale: ja })}</span>
              </div>
              <p
                className={cn(
                  "max-w-[80%] rounded-2xl px-3 py-2 text-sm whitespace-pre-wrap",
                  item.message.from === "staff" ? "bg-primary text-primary-foreground" : "bg-muted",
                )}
              >
                {item.message.text}
              </p>
            </div>
          ) : (
            <div key={`history-${item.history.id}`} className="flex justify-center">
              <p className="rounded-full bg-muted/60 px-3 py-1 text-xs text-muted-foreground">
                {item.history.changed_by_name}が{describe(item.history)}
                <span className="ml-2">
                  {formatDistanceToNow(item.at, { addSuffix: true, locale: ja })}
                </span>
              </p>
            </div>
          ),
        )}
        <div ref={endRef} />
      </div>
      <form onSubmit={handleSubmit} className="border-t p-4">
        <div className="flex gap-2">
          <Input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            placeholder="メッセージを入力..."
            aria-label="メッセージ"
            className="flex-1"
          />
          <Button type="submit" disabled={!draft.trim()} className="gap-2">
            <Send className="h-4 w-4" />
            送信
          </Button>
        </div>
        {chat.footnote && <div className="mt-2 text-xs text-muted-foreground">{chat.footnote}</div>}
      </form>
    </div>
  )
}
