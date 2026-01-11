import { useEffect, useState } from "react"
import { Navigate, useParams } from "react-router-dom"
import { AppLayout } from "@/components/app-layout"
import { FormManagementHeader } from "@/components/form-management-header"
import { ResponseTableView } from "@/components/response-table-view"
import { ResponseKanbanView } from "@/components/response-kanban-view"
import { ResponseDetail } from "@/components/response-detail"
import { MembersDialog } from "@/components/members-dialog"
import { useFormResponses } from "@/hooks/use-form-responses"
import { apiClient } from "@/lib/api"
import type { FormResponse } from "@/types/form-response"
import type { Member, FormStatus, Form, FormQuestion } from "@/types"

export default function FormManagementPage() {
  const params = useParams()
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

  const formResponses = responses

  const filteredResponses = formResponses.filter((response) => {
    const matchesSearch =
      response.respondentEmail.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = statusFilter === "all" || response.status === statusFilter

    return matchesSearch && matchesStatus
  })

  const formTitle = formResponses[0]?.formTitle || "フォーム管理"

  const handleOpenDetail = (response: FormResponse) => {
    setSelectedResponse(response)
    setIsDetailOpen(true)
  }

  const handleTitleQuestionChange = async (questionId: string | null) => {
    if (!formId) return
    try {
      await apiClient.updateFormTitleQuestion(formId, questionId)
      setForm((prev) => (prev ? { ...prev, title_question_id: questionId } : null))
    } catch (error) {
      console.error("Failed to update title question:", error)
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
          onMembersClick={() => setIsMembersOpen(true)}
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

      <MembersDialog formId={formId} open={isMembersOpen} onOpenChange={setIsMembersOpen} />
    </AppLayout>
  )
}
