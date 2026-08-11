import { useEffect, useMemo, useState } from "react"
import { Navigate, useNavigate, useParams } from "react-router-dom"
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
import { notificationLabel, useFormResponses } from "@/hooks/use-form-responses"
import { useAuth } from "@/hooks/useAuth"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import {
  FALLBACK_ASSIGNEE_NAME,
  FALLBACK_STATUS_NAME,
} from "@/lib/notification-email-preview"
import type { FormResponse } from "@/types/form-response"
import type { Member, FormStatus, Form, FormQuestion } from "@/types"
import { toast } from "sonner"

export default function FormManagementPage() {
  const params = useParams()
  const navigate = useNavigate()
  const formId = params.id
  const {
    responses,
    updateResponseStatus,
    assignResponse,
    updatePriority,
    refetch,
    pendingNotification,
    resolvePendingNotification,
    cancelPendingNotification,
    sendNotification,
  } = useFormResponses(formId ?? null)
  const { user } = useAuth()
  const [form, setForm] = useState<Form | null>(null)
  const [questions, setQuestions] = useState<FormQuestion[]>([])
  const [members, setMembers] = useState<Member[]>([])
  const [formStatuses, setFormStatuses] = useState<FormStatus[]>([])
  const [viewMode, setViewMode] = useState<"list" | "kanban">("list")
  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState<"all" | string>("all")
  const [selectedResponseId, setSelectedResponseId] = useState<string | null>(null)
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
    [members, user?.id]
  )

  const statusMap = useMemo(() => new Map(formStatuses.map((status) => [status.id, status])), [formStatuses])
  const formResponses = useMemo(
    () =>
      responses.map((response) => {
        const status = statusMap.get(response.status)
        if (!status) return response
        if (response.statusName === status.name && response.statusColor === status.color) {
          return response
        }
        return {
          ...response,
          statusName: status.name,
          statusColor: status.color,
        }
      }),
    [responses, statusMap]
  )

  const filteredResponses = useMemo(
    () =>
      formResponses.filter((response) => {
        const matchesSearch = response.respondentEmail.toLowerCase().includes(searchQuery.toLowerCase())
        const matchesStatus = statusFilter === "all" || response.status === statusFilter
        return matchesSearch && matchesStatus
      }),
    [formResponses, searchQuery, statusFilter]
  )

  // 更新後の値をダイアログに反映するため、開いた時点のスナップショットではなく最新の一覧から引く。
  const selectedResponse = useMemo(
    () => formResponses.find((response) => response.id === selectedResponseId) ?? null,
    [formResponses, selectedResponseId]
  )

  const formTitle = formResponses[0]?.formTitle || "フォーム管理"

  const notificationEmailSample = useMemo(
    () => ({
      formTitle: form?.title ?? formTitle,
      // 先頭のステータスは新規チケットの初期値であり変更後の値になりにくいため、末尾のものを使う。
      statusName: formStatuses[formStatuses.length - 1]?.name ?? FALLBACK_STATUS_NAME,
      // 閲覧者はフォームのメンバーであり実際に割り当てられうる人物のため、サンプルとして実態に近い。
      assigneeName:
        members.find((member) => member.id === user?.id)?.display_name ?? FALLBACK_ASSIGNEE_NAME,
    }),
    [form?.title, formTitle, formStatuses, members, user?.id]
  )

  const handleOpenDetail = (response: FormResponse) => {
    setSelectedResponseId(response.id)
    setIsDetailOpen(true)
  }

  const handleTitleQuestionChange = async (questionId: string | null) => {
    if (!formId) return
    try {
      await apiClient.updateForm(formId, { title_question_id: questionId })
      setForm((prev) => (prev ? { ...prev, title_question_id: questionId } : null))
    } catch (error) {
      console.error("Failed to update title question:", error)
      toast.error(
        getApiErrorMessage(
          error,
          {
            VALIDATION_ERROR: "タイトルに設定する質問を確認してください",
            RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
            FORM_NOT_FOUND: "フォームが見つかりません",
            INVALID_SESSION: "セッションの有効期限が切れました。ログインし直してください",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "タイトル質問の更新に失敗しました"
        )
      )
    }
  }

  const handleSync = async () => {
    if (!formId || isSyncing) return
    setIsSyncing(true)
    try {
      const result = await apiClient.syncForm(formId)
      await refetch()
      toast.success(
        result.new_tickets > 0
          ? `${result.new_tickets}件の新しい回答を同期しました`
          : "新しい回答はありませんでした"
      )
    } catch (error) {
      console.error("Failed to sync form:", error)
      toast.error(
        getApiErrorMessage(
          error,
          {
            RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
            FORM_NOT_FOUND: "フォームが見つかりません",
            FORM_NOT_SHARED:
              "フォームがサービスアカウントに共有されていません。共有設定を確認してください",
            INVALID_SESSION: "セッションの有効期限が切れました。ログインし直してください",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "フォームの同期に失敗しました"
        )
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
      navigate("/")
    } catch (error) {
      console.error("Failed to delete form:", error)
      setUnregisterError(
        getApiErrorMessage(
          error,
          {
            FORBIDDEN: "この操作を行う権限がありません（管理者のみ）",
            RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
            INVALID_SESSION: "セッションの有効期限が切れました。ログインし直してください",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "フォームの登録解除に失敗しました"
        )
      )
    } finally {
      setIsUnregistering(false)
    }
  }

  useEffect(() => {
    if (!formId) return
    let isActive = true

    const loadData = async () => {
      try {
        const [formResponse, questionsResponse, membersResponse, statusesResponse] = await Promise.all([
          apiClient.getForm(formId),
          apiClient.getFormQuestions(formId),
          apiClient.getMembers(formId),
          apiClient.getFormStatuses(formId),
        ])
        if (!isActive) return
        setForm(formResponse)
        setQuestions(questionsResponse.questions)
        setMembers(membersResponse.members)
        setFormStatuses(statusesResponse.statuses)
      } catch (error) {
        if (!isActive) return
        console.error("Failed to load data:", error)
      }
    }

    loadData()

    return () => {
      isActive = false
    }
  }, [formId])

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
          statusFilter={statusFilter}
          onStatusFilterChange={setStatusFilter}
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
            responses={filteredResponses}
            users={members.map((member) => ({
              id: member.id,
              name: member.display_name,
              email: member.email,
            }))}
            statuses={formStatuses}
            titleQuestionId={form?.title_question_id ?? null}
            onStatusChange={updateResponseStatus}
            onAssignChange={assignResponse}
            onPriorityChange={updatePriority}
            onOpenDetail={handleOpenDetail}
          />
        ) : (
          <ResponseKanbanView
            responses={filteredResponses}
            users={members.map((member) => ({
              id: member.id,
              name: member.display_name,
              email: member.email,
            }))}
            statuses={formStatuses}
            titleQuestionId={form?.title_question_id ?? null}
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
          currentUserId="1"
          currentUserName="田中 太郎"
          onSendNotification={sendNotification}
        />
      )}

      {formId && (
        <StatusesDialog
          formId={formId}
          open={isStatusesOpen}
          onOpenChange={setIsStatusesOpen}
          statuses={formStatuses}
          onStatusesChange={setFormStatuses}
        />
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
              「{formTitle}」を登録解除すると、担当者の割り当てなどを含むすべてのデータが削除され元に戻せません。この操作は管理者のみが実行できます。Googleフォームのフォームやデータ自体は削除されません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {unregisterError && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {unregisterError}
            </div>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isUnregistering}>キャンセル</AlertDialogCancel>
            <Button
              onClick={handleUnregister}
              disabled={isUnregistering}
              variant="destructive"
            >
              登録解除する
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AppLayout>
  )
}
