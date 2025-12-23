"use client"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { LayoutList, LayoutGrid, Search, RefreshCw } from "lucide-react"
import { SyncFormsDialog } from "./sync-forms-dialog"

interface DashboardHeaderProps {
  viewMode: "list" | "kanban"
  onViewModeChange: (mode: "list" | "kanban") => void
  searchQuery: string
  onSearchChange: (query: string) => void
  statusFilter: string
  onStatusFilterChange: (status: string) => void
  onRefresh: () => void
}

export function DashboardHeader({
  viewMode,
  onViewModeChange,
  searchQuery,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  onRefresh,
}: DashboardHeaderProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">フォーム回答管理</h1>
        <div className="flex gap-2">
          <SyncFormsDialog />
          <Button onClick={onRefresh} variant="outline" size="sm" className="gap-2 bg-transparent">
            <RefreshCw className="h-4 w-4" />
            更新
          </Button>
        </div>
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
            <SelectItem value="new">新規</SelectItem>
            <SelectItem value="in-review">対応中</SelectItem>
            <SelectItem value="needs-info">情報待ち</SelectItem>
            <SelectItem value="completed">完了</SelectItem>
          </SelectContent>
        </Select>

        <div className="flex gap-1 bg-muted p-1 rounded-lg">
          <Button
            variant={viewMode === "list" ? "default" : "ghost"}
            size="sm"
            onClick={() => onViewModeChange("list")}
            className="gap-2"
          >
            <LayoutList className="h-4 w-4" />
            リスト
          </Button>
          <Button
            variant={viewMode === "kanban" ? "default" : "ghost"}
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
