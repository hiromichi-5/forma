import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { LayoutList, LayoutGrid, Search, Users, ArrowLeft } from "lucide-react"
import { useNavigate } from "react-router-dom"
import type { FormStatus } from "@/types"

type FormManagementHeaderProps = {
  formTitle: string
  viewMode: "list" | "kanban"
  onViewModeChange: (mode: "list" | "kanban") => void
  searchQuery: string
  onSearchChange: (query: string) => void
  statusFilter: "all" | string
  onStatusFilterChange: (status: "all" | string) => void
  statuses: FormStatus[]
  onMembersClick: () => void
}

export function FormManagementHeader({
  formTitle,
  viewMode,
  onViewModeChange,
  searchQuery,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  statuses,
  onMembersClick,
}: FormManagementHeaderProps) {
  const navigate = useNavigate()
  const sortedStatuses = [...statuses].sort((a, b) => a.display_order - b.display_order)

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => navigate("/")}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-foreground">{formTitle}</h1>
        </div>
        <Button onClick={onMembersClick} variant="outline" className="gap-2 bg-transparent">
          <Users className="h-4 w-4" />
          メンバー管理
        </Button>
      </div>

      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="回答者名やメールアドレスで検索..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-9"
          />
        </div>

        <Select value={statusFilter} onValueChange={onStatusFilterChange}>
          <SelectTrigger className="w-full sm:w-[180px]">
            <SelectValue placeholder="ステータス絞込" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">すべて</SelectItem>
            {sortedStatuses.map((status) => (
              <SelectItem key={status.id} value={status.id}>
                {status.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <div className="flex gap-1 bg-muted p-1 rounded-md">
          <Button
            variant={viewMode === "list" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => onViewModeChange("list")}
            className="gap-2"
          >
            <LayoutList className="h-4 w-4" />
            リスト
          </Button>
          <Button
            variant={viewMode === "kanban" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => onViewModeChange("kanban")}
            className="gap-2"
          >
            <LayoutGrid className="h-4 w-4" />
            カンバン
          </Button>
        </div>
      </div>
    </div>
  )
}
