import { useMemo, useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { Navigate, useNavigate, useParams } from "react-router"
import { AppLayout } from "@/components/app-layout"
import { FormManagementHeader } from "@/components/form-management-header"
import { ResponseTableView } from "@/components/response-table-view"
import { ResponseKanbanView } from "@/components/response-kanban-view"
import { ResponseDetail } from "@/components/response-detail"
import { MembersDialog } from "@/components/members-dialog"
import { NotificationsDialog } from "@/components/notifications-dialog"
import { StatusesDialog } from "@/components/statuses-dialog"
import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { notificationLabel, useTicketActions } from "@/hooks/use-ticket-actions"
import { useTicketStream } from "@/hooks/use-ticket-stream"
import { useDebouncedValue } from "@/hooks/use-debounced-value"
import { useAuth } from "@/hooks/use-auth"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage, toastApiError } from "@/lib/api-error"
import { FALLBACK_ASSIGNEE_NAME, FALLBACK_STATUS_NAME } from "@/lib/notification-email-preview"
import { sortStatuses } from "@/lib/ticket-display"
import type { FormStatus, TicketSummary } from "@/types"
import { formQueries } from "@/queries/forms"
import { queryKeys } from "@/queries/keys"
import { toast } from "sonner"

const EMPTY_STATUSES: FormStatus[] = []

export default function FormManagementPage() {
  const params = useParams()
  const navigate = useNavigate()
  const formId = params.id
  const queryClient = useQueryClient()
  const enabled = Boolean(formId)
  const { data: form = null } = useQuery({ ...formQueries.detail(formId ?? ""), enabled })
  const { data: questions = [] } = useQuery({ ...formQueries.questions(formId ?? ""), enabled })
  const { data: members = [] } = useQuery({ ...formQueries.members(formId ?? ""), enabled })
  const { data: formStatuses = EMPTY_STATUSES } = useQuery({
    ...formQueries.statuses(formId ?? ""),
    enabled,
  })
  const [viewMode, setViewMode] = useState<"list" | "kanban">("list")
  const [searchQuery, setSearchQuery] = useState("")
  const debouncedQuery = useDebouncedValue(searchQuery.trim(), 300)
  // 空配列は絞り込みなし（全てのステータス）を表す。
  const [statusFilters, setStatusFilters] = useState<string[]>([])

  const ticketFilter = useMemo(
    () => ({ statusIds: statusFilters, query: debouncedQuery }),
    [statusFilters, debouncedQuery],
  )
  // カンバンは絞り込みで選んだステータスの列だけを表示する。
  const kanbanStatuses = useMemo(
    () =>
      sortStatuses(formStatuses).filter(
        (status) => statusFilters.length === 0 || statusFilters.includes(status.id),
      ),
    [formStatuses, statusFilters],
  )

  useTicketStream(formId ?? null)
  const {
    updateResponseStatus,
    assignResponse,
    updatePriority,
    pendingNotification,
    resolvePendingNotification,
    cancelPendingNotification,
    sendNotification,
  } = useTicketActions(formId ?? "")
  const { user } = useAuth()
  const [selectedResponse, setSelectedResponse] = useState<TicketSummary | null>(null)
  const [isDetailOpen, setIsDetailOpen] = useState(false)
  const [isMembersOpen, setIsMembersOpen] = useState(false)
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false)
  const [isStatusesOpen, setIsStatusesOpen] = useState(false)
  const [isUnregisterOpen, setIsUnregisterOpen] = useState(false)
  const [isUnregistering, setIsUnregistering] = useState(false)
  const [unregisterError, setUnregisterError] = useState("")
  const [isSyncing, setIsSyncing] = useState(false)

  const isAdmin = useMemo(
    () => members.some((member) => member.id === user?.id && member.role === "admin"),
    [members, user?.id],
  )

  const formTitle = form?.title || "フォーム管理"

  const notificationEmailSample = useMemo(
    () => ({
      formTitle: form?.title ?? formTitle,
      // 先頭のステータスは新規チケットの初期値であり変更後の値になりにくいため、末尾のものを使う。
      statusName: formStatuses[formStatuses.length - 1]?.name ?? FALLBACK_STATUS_NAME,
      // 閲覧者はフォームのメンバーであり実際に割り当てられうる人物のため、サンプルとして実態に近い。
      assigneeName:
        members.find((member) => member.id === user?.id)?.display_name ?? FALLBACK_ASSIGNEE_NAME,
    }),
    [form?.title, formTitle, formStatuses, members, user?.id],
  )

  const handleOpenDetail = (response: TicketSummary) => {
    setSelectedResponse(response)
    setIsDetailOpen(true)
  }

  const handleTitleQuestionChange = async (questionId: string | null) => {
    if (!formId) return
    try {
      await apiClient.updateForm(formId, { title_question_id: questionId })
      queryClient.setQueryData(queryKeys.form(formId), (prev) =>
        prev ? { ...prev, title_question_id: questionId } : prev,
      )
      // チケットのタイトルは選んだ質問の回答から作られるため、一覧を取り直す。
      await queryClient.invalidateQueries({ queryKey: queryKeys.formTickets(formId) })
    } catch (error) {
      console.error("Failed to update title question:", error)
      toastApiError(
        error,
        {
          VALIDATION_ERROR: "タイトルに設定する質問を確認してください",
          RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
          FORM_NOT_FOUND: "フォームが見つかりません",
        },
        "タイトル質問の更新に失敗しました",
      )
    }
  }

  const handleSync = async () => {
    if (!formId || isSyncing) return
    setIsSyncing(true)
    try {
      const result = await apiClient.syncForm(formId)
      await queryClient.invalidateQueries({ queryKey: queryKeys.formTickets(formId) })
      // フォーム一覧の回答数も変わる。
      void queryClient.invalidateQueries({ queryKey: queryKeys.forms(), exact: true })
      toast.success(
        result.new_tickets > 0
          ? `${result.new_tickets}件の新しい回答を同期しました`
          : "新しい回答はありませんでした",
      )
    } catch (error) {
      console.error("Failed to sync form:", error)
      toastApiError(
        error,
        {
          RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
          FORM_NOT_FOUND: "フォームが見つかりません",
          FORM_NOT_SHARED:
            "フォームがサービスアカウントに共有されていません。共有設定を確認してください",
        },
        "フォームの同期に失敗しました",
      )
    } finally {
      setIsSyncing(false)
    }
  }

  const handleUnregister = async () => {
    if (!formId) return
    setIsUnregistering(true)
    setUnregisterError("")
    try {
      await apiClient.deleteForm(formId)
      // 一覧とサイドバーに削除したフォームが残らないよう、遷移前に取り直す。
      await queryClient.invalidateQueries({ queryKey: queryKeys.forms(), exact: true })
      void navigate("/")
    } catch (error) {
      console.error("Failed to delete form:", error)
      setUnregisterError(
        getApiErrorMessage(
          error,
          {
            FORBIDDEN: "この操作を行う権限がありません（管理者のみ）",
            RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
          },
          "フォームの登録解除に失敗しました",
        ),
      )
    } finally {
      setIsUnregistering(false)
    }
  }

  if (!formId) {
    return <Navigate to="/" replace />
  }

  return (
    <AppLayout>
      <div className="space-y-6 pt-4">
        <FormManagementHeader
          googleFormId={form?.form_id ?? null}
          formTitle={formTitle}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          statusFilters={statusFilters}
          onStatusFiltersChange={setStatusFilters}
          statuses={formStatuses}
          questions={questions}
          titleQuestionId={form?.title_question_id ?? null}
          onTitleQuestionChange={handleTitleQuestionChange}
          onStatusManageClick={() => setIsStatusesOpen(true)}
          onMembersClick={() => setIsMembersOpen(true)}
          onNotificationsClick={() => setIsNotificationsOpen(true)}
          onUnregisterClick={() => setIsUnregisterOpen(true)}
          isSyncing={isSyncing}
          onSyncClick={handleSync}
        />

        {viewMode === "list" ? (
          <ResponseTableView
            formId={formId}
            filter={ticketFilter}
            members={members}
            statuses={formStatuses}
            onStatusChange={updateResponseStatus}
            onAssignChange={assignResponse}
            onPriorityChange={updatePriority}
            onOpenDetail={handleOpenDetail}
          />
        ) : (
          <ResponseKanbanView
            formId={formId}
            query={debouncedQuery}
            columns={kanbanStatuses}
            members={members}
            onStatusChange={updateResponseStatus}
            onAssignChange={assignResponse}
            onPriorityChange={updatePriority}
            onOpenDetail={handleOpenDetail}
          />
        )}
      </div>

      {selectedResponse && (
        <ResponseDetail
          response={selectedResponse}
          open={isDetailOpen}
          onOpenChange={setIsDetailOpen}
          members={members}
          statuses={formStatuses}
          onStatusChange={updateResponseStatus}
          onAssignChange={assignResponse}
          onPriorityChange={updatePriority}
          onSendNotification={sendNotification}
        />
      )}

      {formId && (
        <StatusesDialog formId={formId} open={isStatusesOpen} onOpenChange={setIsStatusesOpen} />
      )}

      <MembersDialog formId={formId} open={isMembersOpen} onOpenChange={setIsMembersOpen} />

      <NotificationsDialog
        formId={formId}
        open={isNotificationsOpen}
        onOpenChange={setIsNotificationsOpen}
        canEdit={isAdmin}
        emailSample={notificationEmailSample}
      />

      <AlertDialog
        open={pendingNotification !== null}
        onOpenChange={(open) => {
          if (!open) cancelPendingNotification()
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>回答者に通知しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              {pendingNotification
                ? `${notificationLabel(pendingNotification.notificationType)}を回答者にメールで通知します。通知しない場合も変更は保存されます。`
                : ""}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>キャンセル</AlertDialogCancel>
            <Button variant="outline" onClick={() => resolvePendingNotification(false)}>
              通知せず変更
            </Button>
            <Button onClick={() => resolvePendingNotification(true)}>通知して変更</Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog
        open={isUnregisterOpen}
        onOpenChange={(open) => {
          if (!open) {
            setIsUnregisterOpen(false)
            setUnregisterError("")
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>フォームの登録を解除しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              「{formTitle}
              」を登録解除すると、担当者の割り当てなどを含むすべてのデータが削除され元に戻せません。この操作は管理者のみが実行できます。Googleフォームのフォームやデータ自体は削除されません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {unregisterError && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {unregisterError}
            </div>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isUnregistering}>キャンセル</AlertDialogCancel>
            <Button onClick={handleUnregister} disabled={isUnregistering} variant="destructive">
              登録解除する
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AppLayout>
  )
}
