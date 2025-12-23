import { useState } from "react"
import { useParams } from "react-router-dom"
import { AppLayout } from "@/components/app-layout"
import { FormManagementHeader } from "@/components/form-management-header"
import { ResponseTableView } from "@/components/response-table-view"
import { ResponseKanbanViewNew } from "@/components/response-kanban-view-new"
import { ChatInterface } from "@/components/chat-interface"
import { MembersDialog } from "@/components/members-dialog"
import { useFormResponses } from "@/hooks/use-form-responses"
import { mockUsers } from "@/lib/mock-data"
import type { FormResponse } from "@/types/form-response"

export default function FormManagementPage() {
  const params = useParams()
  const formId = params.id as string
  const { responses, updateResponseStatus, assignResponse, updatePriority } = useFormResponses()
  const [viewMode, setViewMode] = useState<"list" | "kanban">("list")
  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [selectedResponse, setSelectedResponse] = useState<FormResponse | null>(null)
  const [isChatOpen, setIsChatOpen] = useState(false)
  const [isMembersOpen, setIsMembersOpen] = useState(false)

  const formResponses = responses.filter((r) => r.formId === formId)

  const filteredResponses = formResponses.filter((response) => {
    const matchesSearch =
      response.respondentName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      response.respondentEmail.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = statusFilter === "all" || response.status === statusFilter

    return matchesSearch && matchesStatus
  })

  const formTitle = formResponses[0]?.formTitle || "フォーム管理"

  const handleOpenChat = (response: FormResponse) => {
    setSelectedResponse(response)
    setIsChatOpen(true)
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
          onMembersClick={() => setIsMembersOpen(true)}
        />

        {viewMode === "list" ? (
          <ResponseTableView
            responses={filteredResponses}
            users={mockUsers}
            onStatusChange={updateResponseStatus}
            onAssignChange={assignResponse}
            onPriorityChange={updatePriority}
            onOpenChat={handleOpenChat}
          />
        ) : (
          <ResponseKanbanViewNew
            responses={filteredResponses}
            users={mockUsers}
            onStatusChange={updateResponseStatus}
            onAssignChange={assignResponse}
            onPriorityChange={updatePriority}
            onOpenChat={handleOpenChat}
          />
        )}
      </div>

      {isChatOpen && selectedResponse && (
        <ChatInterface
          response={selectedResponse}
          onClose={() => setIsChatOpen(false)}
          currentUserId="1"
          currentUserName="田中 太郎"
        />
      )}

      {isMembersOpen && <MembersDialog formId={formId} onClose={() => setIsMembersOpen(false)} />}
    </AppLayout>
  )
}
