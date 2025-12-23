import { useEffect, useState } from "react"
import { Navigate, useParams } from "react-router-dom"
import { AppLayout } from "@/components/app-layout"
import { FormManagementHeader } from "@/components/form-management-header"
import { ResponseTableView } from "@/components/response-table-view"
import { ResponseKanbanView } from "@/components/response-kanban-view"
import { ChatInterface } from "@/components/chat-interface"
import { MembersDialog } from "@/components/members-dialog"
import { useFormResponses } from "@/hooks/use-form-responses"
import { apiClient } from "@/lib/api"
import type { FormResponse } from "@/types/form-response"
import type { Member } from "@/types"

export default function FormManagementPage() {
  const params = useParams()
  const formId = params.id
  const { responses, updateResponseStatus, assignResponse, updatePriority } = useFormResponses(formId ?? null)
  const [members, setMembers] = useState<Member[]>([])
  const [viewMode, setViewMode] = useState<"list" | "kanban">("list")
  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState<"all" | FormResponse["status"]>("all")
  const [selectedResponse, setSelectedResponse] = useState<FormResponse | null>(null)
  const [isChatOpen, setIsChatOpen] = useState(false)
  const [isMembersOpen, setIsMembersOpen] = useState(false)

  const formResponses = responses

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

  useEffect(() => {
    if (!formId) return
    let isActive = true

    const loadMembers = async () => {
      try {
        // 外部API(バックエンド)との同期のための処理
        const response = await apiClient.getMembers(formId)
        if (!isActive) return
        setMembers(response.members)
      } catch (error) {
        if (!isActive) return
        console.error("Failed to load members:", error)
      }
    }

    // 外部API(バックエンド)との同期のための処理
    loadMembers()

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
            onStatusChange={updateResponseStatus}
            onAssignChange={assignResponse}
            onPriorityChange={updatePriority}
            onOpenChat={handleOpenChat}
          />
        ) : (
          <ResponseKanbanView
            responses={filteredResponses}
            users={members.map((member) => ({
              id: member.id,
              name: member.display_name,
              email: member.email,
            }))}
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
