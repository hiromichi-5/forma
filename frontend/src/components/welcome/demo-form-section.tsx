import { useLayoutEffect, useMemo, useRef, useState } from "react"
import { toast } from "sonner"
import { FormManagementHeader } from "@/components/form-management-header"
import { ResponseTableView } from "@/components/response-table-view"
import { ResponseKanbanView } from "@/components/response-kanban-view"
import { ResponseDetail, type ResponseChatMessage } from "@/components/response-detail"
import { Annotation } from "@/components/welcome/annotation"
import type { DemoWorkspace } from "@/components/welcome/demo-workspace"
import {
  DemoMembersDialog,
  DemoNotificationsDialog,
  DemoStatusesDialog,
} from "@/components/welcome/demo-settings-dialogs"
import { DemoMemo, UpcomingBadge } from "@/components/welcome/demo-upcoming-panels"
import { useSignupPrompt } from "@/hooks/use-signup-prompt"
import { DEMO_VIEWER } from "@/components/welcome/demo-data"
import { sortStatuses } from "@/lib/ticket-display"
import type { NotificationType, TicketSummary } from "@/types"

type DemoFormSectionProps = {
  workspace: DemoWorkspace
  formId: string
  onSynced: () => void
}

const SYNC_DELAY_MS = 800

/** フォーム管理画面のデモ。本番の画面部品に、手元のデータをつないで動かす。 */
export function DemoFormSection({ workspace, formId, onSynced }: DemoFormSectionProps) {
  const promptSignup = useSignupPrompt()
  const { title, questions, statuses, members, titleQuestionId } = workspace.form(formId)
  const [viewMode, setViewMode] = useState<"list" | "kanban">("kanban")
  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilters, setStatusFilters] = useState<string[]>([])
  const [currentTitleQuestionId, setCurrentTitleQuestionId] = useState(titleQuestionId)
  const [isSyncing, setIsSyncing] = useState(false)
  const [selectedResponse, setSelectedResponse] = useState<TicketSummary | null>(null)
  const [isDetailOpen, setIsDetailOpen] = useState(false)
  const [chatMessages, setChatMessages] = useState<ResponseChatMessage[]>([])
  const [openSetting, setOpenSetting] = useState<"statuses" | "members" | "notifications" | null>(
    null,
  )
  const headerRef = useRef<HTMLDivElement>(null)
  const syncButtonLeft = useSyncButtonLeft(headerRef)

  const query = searchQuery.trim()
  const ticketFilter = useMemo(() => ({ statusIds: statusFilters, query }), [statusFilters, query])
  const kanbanStatuses = useMemo(
    () =>
      sortStatuses(statuses).filter(
        (status) => statusFilters.length === 0 || statusFilters.includes(status.id),
      ),
    [statuses, statusFilters],
  )

  // 条件を変える前に、その条件の一覧を用意しておく。
  const handleSearchChange = (value: string) => {
    workspace.prepare(formId, { statusIds: statusFilters, query: value.trim() })
    setSearchQuery(value)
  }

  const handleStatusFiltersChange = (statusIds: string[]) => {
    workspace.prepare(formId, { statusIds, query })
    setStatusFilters(statusIds)
  }

  const handleTitleQuestionChange = (questionId: string | null) => {
    workspace.setTitleQuestion(formId, questionId)
    setCurrentTitleQuestionId(questionId)
  }

  const handleSync = () => {
    setIsSyncing(true)
    setTimeout(() => {
      const added = workspace.sync(formId)
      setIsSyncing(false)
      onSynced()
      toast.success(
        added > 0 ? `${added}件の新しい回答を同期しました` : "新しい回答はありませんでした",
      )
    }, SYNC_DELAY_MS)
  }

  const handleSendNotification = (ticketId: string, type: NotificationType) => {
    workspace.sendNotification(formId, ticketId, type)
    toast.success("回答者に通知メールを送信しました")
    return Promise.resolve(true)
  }

  const handleSendChat = (ticketId: string, text: string) => {
    setChatMessages(workspace.sendMessage(formId, ticketId, text))
  }

  const actions = {
    onStatusChange: (id: string, statusId: string) =>
      workspace.updateTicket(formId, id, { status_id: statusId }),
    onAssignChange: (id: string, userId: string | null) =>
      workspace.updateTicket(formId, id, { assignee_id: userId }),
    onPriorityChange: (id: string, priority: TicketSummary["priority"]) =>
      workspace.updateTicket(formId, id, { priority }),
    onOpenDetail: (response: TicketSummary) => {
      setChatMessages(workspace.chat(formId, response.id))
      setSelectedResponse(response)
      setIsDetailOpen(true)
    },
  }

  return (
    <div className="space-y-6">
      <div className="relative pt-20" ref={headerRef}>
        <Annotation arrow="down" align="start" className="absolute top-0 left-2 hidden md:flex">
          ここから設定を開けます
        </Annotation>
        {syncButtonLeft !== null && (
          <Annotation
            arrow="down"
            align="start"
            className="absolute top-0 hidden md:flex"
            style={{ left: syncButtonLeft }}
          >
            Googleフォームの新しい回答を取り込めます
          </Annotation>
        )}
        <FormManagementHeader
          googleFormId={null}
          formTitle={title}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          searchQuery={searchQuery}
          onSearchChange={handleSearchChange}
          statusFilters={statusFilters}
          onStatusFiltersChange={handleStatusFiltersChange}
          statuses={statuses}
          questions={questions}
          titleQuestionId={currentTitleQuestionId}
          onTitleQuestionChange={handleTitleQuestionChange}
          onStatusManageClick={() => setOpenSetting("statuses")}
          onMembersClick={() => setOpenSetting("members")}
          onNotificationsClick={() => setOpenSetting("notifications")}
          onUnregisterClick={promptSignup}
          isSyncing={isSyncing}
          onSyncClick={handleSync}
        />
      </div>

      <div className="flex items-start justify-between gap-4">
        <div className="flex gap-16">
          <Annotation arrow="down" align="start" className="ml-6">
            {viewMode === "kanban"
              ? "カードをドラッグしてステータスを変更できます"
              : "ステータス・担当者・優先度をその場で変更できます"}
          </Annotation>
          <Annotation arrow="down" align="start" className="hidden lg:flex">
            {viewMode === "kanban" ? "カードを開くと" : "右端のアイコンから"}
            回答の内容や対応の記録を見られます
          </Annotation>
        </div>
        <Annotation arrow="up" align="end" className="-mt-4 hidden sm:flex">
          リストとカンバンを切り替えられます
        </Annotation>
      </div>

      {viewMode === "list" ? (
        <ResponseTableView
          formId={formId}
          filter={ticketFilter}
          members={members}
          statuses={statuses}
          {...actions}
        />
      ) : (
        <ResponseKanbanView
          formId={formId}
          query={query}
          columns={kanbanStatuses}
          members={members}
          {...actions}
        />
      )}

      {selectedResponse && (
        <ResponseDetail
          response={selectedResponse}
          open={isDetailOpen}
          onOpenChange={setIsDetailOpen}
          members={members}
          statuses={statuses}
          onStatusChange={actions.onStatusChange}
          onAssignChange={actions.onAssignChange}
          onPriorityChange={actions.onPriorityChange}
          onSendNotification={handleSendNotification}
          chat={{
            messages: chatMessages,
            onSend: (text) => handleSendChat(selectedResponse.id, text),
            footnote: (
              <span className="flex items-center gap-2">
                <UpcomingBadge />
                回答者へメールで届く仕組みは今後追加予定です
              </span>
            ),
          }}
          memo={
            <DemoMemo
              key={selectedResponse.id}
              defaultValue={workspace.memo(formId, selectedResponse.id)}
              onChange={(memo) => workspace.setMemo(formId, selectedResponse.id, memo)}
            />
          }
        />
      )}

      <DemoStatusesDialog
        statuses={statuses}
        open={openSetting === "statuses"}
        onOpenChange={(open) => !open && setOpenSetting(null)}
        onBlockedAction={promptSignup}
      />
      <DemoMembersDialog
        open={openSetting === "members"}
        onOpenChange={(open) => !open && setOpenSetting(null)}
        onBlockedAction={promptSignup}
      />
      <DemoNotificationsDialog
        emailSample={{
          formTitle: title,
          statusName: statuses[statuses.length - 1].name,
          assigneeName: DEMO_VIEWER.display_name,
        }}
        open={openSetting === "notifications"}
        onOpenChange={(open) => !open && setOpenSetting(null)}
        onBlockedAction={promptSignup}
      />
    </div>
  )
}

/** 同期ボタンはフォーム名の右に並ぶため、フォーム名の幅から注釈の位置を決める。 */
function useSyncButtonLeft(containerRef: React.RefObject<HTMLDivElement | null>) {
  const [left, setLeft] = useState<number | null>(null)

  useLayoutEffect(() => {
    const container = containerRef.current
    const titleButton = container?.querySelector("h1")?.closest("button")
    if (!container || !titleButton) return

    const update = () =>
      setLeft(
        titleButton.getBoundingClientRect().right - container.getBoundingClientRect().left + 12,
      )
    update()
    if (typeof ResizeObserver === "undefined") return
    const observer = new ResizeObserver(update)
    observer.observe(container)
    observer.observe(titleButton)
    return () => observer.disconnect()
  }, [containerRef])

  return left
}
