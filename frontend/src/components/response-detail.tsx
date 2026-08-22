"use client";

import type React from "react";

import { useState, useRef, useEffect, useMemo } from "react";
import type { User as MemberUser } from "@/types/form-response";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Send, User, Bot, Mail } from "lucide-react";
import { formatDistanceToNow, format } from "date-fns";
import { ja } from "date-fns/locale";
import { useChatMessages } from "@/hooks/use-chat-messages";
import { useTicketHistories } from "@/hooks/use-ticket-histories";
import { cn } from "@/lib/utils";
import {
  fallbackStatusColor,
  hexToRgba,
  priorityConfig,
  respondentEmailLabel,
  sortStatuses,
  statusById,
  toPriorityValue,
} from "@/lib/ticket-display";
import type {
  FormStatus,
  NotificationType,
  TicketDetail,
  TicketHistory,
  TicketPriority,
  TicketSummary,
} from "@/types";

type TimelineItem =
  | {
      type: "message";
      data: {
        id: string;
        message: string;
        senderName: string;
        senderType: "staff" | "respondent";
        timestamp: Date;
      };
    }
  | {
      type: "history";
      data: TicketHistory;
    };

type ResponseDetailProps = {
  response: TicketSummary;
  /** 回答本文と通知履歴。一覧には含まれないため、読み込みが終わるまで undefined。 */
  detail: TicketDetail | undefined;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentUserId: string;
  currentUserName: string;
  users: MemberUser[];
  statuses: FormStatus[];
  onStatusChange: (id: string, statusId: string) => void;
  onAssignChange: (id: string, userId: string | null) => void;
  onPriorityChange: (id: string, priority: TicketPriority) => void;
  onSendNotification?: (
    responseId: string,
    notificationType: NotificationType
  ) => Promise<boolean>;
};

export function ResponseDetail({
  response,
  detail,
  open,
  onOpenChange,
  currentUserId,
  currentUserName,
  users,
  statuses,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onSendNotification,
}: ResponseDetailProps) {
  const sortedStatuses = sortStatuses(statuses);
  const email = respondentEmailLabel(response.respondent_email);
  const status =
    statusById(statuses).get(response.status.id) ?? response.status;
  const [sendingNotification, setSendingNotification] =
    useState<NotificationType | null>(null);

  const handleSendNotification = async (notificationType: NotificationType) => {
    if (!onSendNotification) return;
    setSendingNotification(notificationType);
    try {
      await onSendNotification(response.id, notificationType);
    } finally {
      setSendingNotification(null);
    }
  };

  const getLastSentAt = (notificationType: NotificationType): string | null =>
    detail?.notifications.find((n) => n.notification_type === notificationType)
      ?.last_sent_at ?? null;

  const formatLastSentAt = (sentAt: string | null): string =>
    sentAt ? `最終送信: ${format(new Date(sentAt), "MM/dd HH:mm")}` : "未送信";
  const { messages, sendMessage } = useChatMessages(response.id);
  const { histories } = useTicketHistories(response.id);
  const [inputValue, setInputValue] = useState("");
  const [memoValue, setMemoValue] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const timelineItems = useMemo(() => {
    const items: TimelineItem[] = [
      ...messages.map(
        (msg): TimelineItem => ({
          type: "message",
          data: {
            id: msg.id,
            message: msg.message,
            senderName: msg.senderName,
            senderType: msg.senderType,
            timestamp: msg.timestamp,
          },
        })
      ),
      ...histories.map(
        (history): TimelineItem => ({
          type: "history",
          data: history,
        })
      ),
    ];

    return items.sort((a, b) => {
      const timeA =
        a.type === "message"
          ? a.data.timestamp.getTime()
          : new Date(a.data.created_at).getTime();
      const timeB =
        b.type === "message"
          ? b.data.timestamp.getTime()
          : new Date(b.data.created_at).getTime();
      return timeA - timeB;
    });
  }, [messages, histories]);

  useEffect(() => {
    scrollToBottom();
  }, [timelineItems]);

  const handleSend = () => {
    if (inputValue.trim()) {
      sendMessage(inputValue.trim(), currentUserId, currentUserName, email);
      setInputValue("");
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const getFieldLabel = (fieldName: string): string => {
    switch (fieldName) {
      case "status":
        return "ステータス";
      case "assignee":
        return "担当者";
      case "priority":
        return "優先度";
      default:
        return fieldName;
    }
  };

  const getPriorityLabel = (priority: string): string => {
    switch (priority) {
      case "high":
        return "高";
      case "medium":
        return "中";
      case "low":
        return "低";
      default:
        return priority;
    }
  };

  const formatHistoryChange = (history: TicketHistory): string => {
    const fieldLabel = getFieldLabel(history.field_name);
    const oldValue = history.old_value
      ? history.field_name === "priority"
        ? getPriorityLabel(history.old_value)
        : history.old_value
      : "なし";
    const newValue = history.new_value
      ? history.field_name === "priority"
        ? getPriorityLabel(history.new_value)
        : history.new_value
      : "なし";

    return `${fieldLabel}を「${oldValue}」から「${newValue}」に変更しました`;
  };

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
                  <span className="text-xs text-muted-foreground whitespace-nowrap">
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
                  <span className="text-xs text-muted-foreground whitespace-nowrap">
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
                  className="w-[150px] h-8 border-0 shadow-none"
                  aria-label="ステータス"
                >
                  <div
                    className="flex items-center gap-2 px-2 py-1 rounded"
                    style={{
                      backgroundColor: hexToRgba(status.color, 0.1),
                    }}
                  >
                    <div
                      className="w-2 h-2 rounded-full shrink-0"
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
                  onAssignChange(
                    response.id,
                    value === "unassigned" ? null : value
                  )
                }
              >
                <SelectTrigger
                  className="w-[150px] h-8 border-0 shadow-none"
                  aria-label="担当者"
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="unassigned">未割当</SelectItem>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      {user.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center gap-1.5">
              <span className="text-xs text-muted-foreground">優先度</span>
              <Select
                value={response.priority}
                onValueChange={(value) =>
                  onPriorityChange(response.id, toPriorityValue(value))
                }
              >
                <SelectTrigger
                  className="w-[90px] h-8 border-0 shadow-none"
                  aria-label="優先度"
                >
                  <div
                    className="flex items-center gap-2 px-2 py-1 rounded"
                    style={{
                      backgroundColor: hexToRgba(
                        priorityConfig[response.priority].hex,
                        0.1
                      ),
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

        <div className="flex-1 flex overflow-hidden">
          <div className="w-1/2 border-r flex flex-col">
            <div className="flex-1 overflow-y-auto p-4 border-b">
              <div className="flex items-center gap-2 mb-3">
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
                      className="bg-muted/50 p-3 rounded-lg"
                    >
                      <p className="font-medium text-muted-foreground text-xs mb-1">
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

            <div className="h-1/3 flex flex-col p-4">
              <div className="flex items-center gap-2 mb-2">
                <Badge variant="secondary">メモ</Badge>
                <span className="text-xs text-muted-foreground">
                  メモ（回答者には表示されません）
                  フォームの担当者内で共有できます
                </span>
              </div>
              <Textarea
                placeholder="メモを入力..."
                value={memoValue}
                onChange={(e) => setMemoValue(e.target.value)}
                className="flex-1 resize-none"
              />
            </div>
          </div>

          {/* 右側: チャット欄 */}
          <div className="w-1/2 flex flex-col">
            {/* チャットメッセージエリア */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              {timelineItems.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-center">
                  <Bot className="h-12 w-12 text-muted-foreground mb-3" />
                  <p className="text-muted-foreground">
                    まだメッセージがありません
                  </p>
                  <p className="text-sm text-muted-foreground">
                    追加の質問や不備の確認を送信してください
                  </p>
                </div>
              ) : (
                timelineItems.map((item) => {
                  if (item.type === "message") {
                    const message = item.data;
                    return (
                      <div
                        key={`message-${message.id}`}
                        className={cn(
                          "flex gap-3",
                          message.senderType === "staff"
                            ? "justify-end"
                            : "justify-start"
                        )}
                      >
                        <div
                          className={cn(
                            "flex gap-3 max-w-[80%]",
                            message.senderType === "staff"
                              ? "flex-row-reverse"
                              : "flex-row"
                          )}
                        >
                          <div
                            className={cn(
                              "flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center",
                              message.senderType === "staff"
                                ? "bg-primary text-primary-foreground"
                                : "bg-muted"
                            )}
                          >
                            {message.senderType === "staff" ? (
                              <User className="h-4 w-4" />
                            ) : (
                              <Bot className="h-4 w-4" />
                            )}
                          </div>

                          <div className="flex flex-col gap-1">
                            <div className="flex items-center gap-2">
                              <span className="text-xs font-medium">
                                {message.senderName}
                              </span>
                              <span className="text-xs text-muted-foreground">
                                {formatDistanceToNow(message.timestamp, {
                                  addSuffix: true,
                                  locale: ja,
                                })}
                              </span>
                              <span className="text-xs text-muted-foreground">
                                ({format(message.timestamp, "yyyy/MM/dd HH:mm")})
                              </span>
                            </div>

                            <div
                              className={cn(
                                "rounded-lg p-3",
                                message.senderType === "staff"
                                  ? "bg-primary text-primary-foreground"
                                  : "bg-muted"
                              )}
                            >
                              <p className="text-sm whitespace-pre-wrap">
                                {message.message}
                              </p>
                            </div>
                          </div>
                        </div>
                      </div>
                    );
                  }

                  const history = item.data;
                  return (
                    <div
                      key={`history-${history.id}`}
                      className="flex justify-center"
                    >
                      <div className="bg-muted/30 border border-muted rounded-lg px-4 py-2 max-w-[90%]">
                        <div className="flex items-center gap-2 text-xs text-muted-foreground mb-1">
                          <span className="font-medium">
                            {history.changed_by_name}
                          </span>
                          <span>
                            {formatDistanceToNow(new Date(history.created_at), {
                              addSuffix: true,
                              locale: ja,
                            })}
                          </span>
                          <span>
                            ({format(new Date(history.created_at), "yyyy/MM/dd HH:mm")})
                          </span>
                        </div>
                        <p className="text-sm text-center">
                          {formatHistoryChange(history)}
                        </p>
                      </div>
                    </div>
                  );
                })
              )}
              <div ref={messagesEndRef} />
            </div>

            {/* メッセージ入力欄 */}
            <div className="p-4 border-t">
              <div className="flex gap-2">
                <Input
                  placeholder="メッセージを入力..."
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  onKeyPress={handleKeyPress}
                  className="flex-1"
                />
                <Button
                  onClick={handleSend}
                  disabled={!inputValue.trim()}
                  className="gap-2"
                >
                  <Send className="h-4 w-4" />
                  送信
                </Button>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Enter で送信 / Shift + Enter で改行 /
                回答者にメール通知が送信されます
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
