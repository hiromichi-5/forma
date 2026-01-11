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
import type { Member, FormStatus } from "@/types"

export default function FormManagementPage() {
  const params = useParams()
  const formId = params.id
  const { responses, updateResponseStatus, assignResponse, updatePriority } = useFormResponses(formId ?? null)
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

  useEffect(() => {
    if (!formId) return
    let isActive = true

    const loadData = async () => {
      try {
        const [membersResponse, statusesResponse] = await Promise.all([
          apiClient.getMembers(formId),
          apiClient.getFormStatuses(formId),
        ])
        if (!isActive) return
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
