import { useEffect, useMemo, useState } from "react"
import { Navigate, useNavigate, useParams } from "react-router-dom"
import { AppLayout } from "@/components/app-layout"
import { FormManagementHeader } from "@/components/form-management-header"
import { ResponseTableView } from "@/components/response-table-view"
import { ResponseKanbanView } from "@/components/response-kanban-view"
import { ResponseDetail } from "@/components/response-detail"
import { MembersDialog } from "@/components/members-dialog"
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
import { useFormResponses } from "@/hooks/use-form-responses"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import type { FormResponse } from "@/types/form-response"
import type { Member, FormStatus, Form, FormQuestion } from "@/types"
import { toast } from "sonner"

export default function FormManagementPage() {
  const params = useParams()
  const navigate = useNavigate()
  const formId = params.id
  const { responses, updateResponseStatus, assignResponse, updatePriority } = useFormResponses(formId ?? null)
  const [form, setForm] = useState<Form | null>(null)
  const [questions, setQuestions] = useState<FormQuestion[]>([])
  const [members, setMembers] = useState<Member[]>([])
  const [formStatuses, setFormStatuses] = useState<FormStatus[]>([])
  const [viewMode, setViewMode] = useState<"list" | "kanban">("list")
  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState<"all" | string>("all")
  const [selectedResponse, setSelectedResponse] = useState<FormResponse | null>(null)
  const [isDetailOpen, setIsDetailOpen] = useState(false)
  const [isMembersOpen, setIsMembersOpen] = useState(false)
  const [isStatusesOpen, setIsStatusesOpen] = useState(false)
  const [isUnregisterOpen, setIsUnregisterOpen] = useState(false)
  const [isUnregistering, setIsUnregistering] = useState(false)
  const [unregisterError, setUnregisterError] = useState("")

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

  const formTitle = formResponses[0]?.formTitle || "フォーム管理"

  const handleOpenDetail = (response: FormResponse) => {
    setSelectedResponse(response)
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
      <div className="space-y-6">
        <FormManagementHeader
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
          onUnregisterClick={() => setIsUnregisterOpen(true)}
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

      {isDetailOpen && selectedResponse && (
        <ResponseDetail
          response={selectedResponse}
          onClose={() => setIsDetailOpen(false)}
          currentUserId="1"
          currentUserName="田中 太郎"
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
