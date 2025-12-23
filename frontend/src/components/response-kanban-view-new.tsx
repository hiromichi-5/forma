"use client"

import type { FormResponse, User } from "@/types/form-response"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { MessageSquare, Calendar } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { ja } from "date-fns/locale"
import { cn } from "@/lib/utils"

interface ResponseKanbanViewNewProps {
  responses: FormResponse[]
  users: User[]
  onStatusChange: (id: string, status: FormResponse["status"]) => void
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: FormResponse["priority"]) => void
  onOpenChat: (response: FormResponse) => void
}

const statusConfig = {
  new: { label: "新規", color: "bg-blue-50" },
  "in-review": { label: "対応中", color: "bg-yellow-50" },
  "needs-info": { label: "情報待ち", color: "bg-orange-50" },
  completed: { label: "完了", color: "bg-green-50" },
}

const priorityConfig = {
  low: { label: "低", color: "text-gray-600" },
  medium: { label: "中", color: "text-blue-600" },
  high: { label: "高", color: "text-red-600" },
}

export function ResponseKanbanViewNew({
  responses,
  users,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onOpenChat,
}: ResponseKanbanViewNewProps) {
  const columns: Array<FormResponse["status"]> = ["new", "in-review", "needs-info", "completed"]

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {columns.map((status) => (
        <div key={status} className="flex flex-col gap-3">
          <div className={cn("flex items-center justify-between p-3 rounded-lg", statusConfig[status].color)}>
            <h3 className="font-semibold text-sm text-foreground">{statusConfig[status].label}</h3>
            <Badge variant="secondary" className="bg-card">
              {responses.filter((r) => r.status === status).length}
            </Badge>
          </div>

          <div className="space-y-3 min-h-[500px]">
            {responses
              .filter((r) => r.status === status)
              .map((response) => (
                <Card key={response.id} className="p-4 border hover:bg-muted/30 transition-colors">
                  <div className="space-y-3">
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex-1">
                        <h4 className="font-semibold text-sm text-foreground mb-1">{response.respondentName}</h4>
                        <p className="text-xs text-muted-foreground truncate">{response.respondentEmail}</p>
                      </div>
                      <Select
                        value={response.priority}
                        onValueChange={(value) => onPriorityChange(response.id, value as FormResponse["priority"])}
                      >
                        <SelectTrigger className="w-[70px] h-7 text-xs">
                          <span className={cn("font-medium", priorityConfig[response.priority].color)}>
                            {priorityConfig[response.priority].label}
                          </span>
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="low">低</SelectItem>
                          <SelectItem value="medium">中</SelectItem>
                          <SelectItem value="high">高</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="text-xs text-muted-foreground line-clamp-2">
                      {Object.values(response.responses)[0]}
                    </div>

                    <Select
                      value={response.assignedTo || "unassigned"}
                      onValueChange={(value) => onAssignChange(response.id, value === "unassigned" ? null : value)}
                    >
                      <SelectTrigger className="w-full h-8 text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="unassigned">未割当</SelectItem>
                        {users.map((user) => (
                          <SelectItem key={user.id} value={user.id}>
                            {user.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>

                    <div className="flex items-center justify-between pt-2 border-t">
                      <div className="flex items-center gap-1 text-xs text-muted-foreground">
                        <Calendar className="h-3 w-3" />
                        <span>{formatDistanceToNow(response.submittedAt, { locale: ja })}</span>
                      </div>
                      <Button variant="ghost" size="sm" onClick={() => onOpenChat(response)} className="gap-1 h-7 px-2">
                        <MessageSquare className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                </Card>
              ))}
          </div>
        </div>
      ))}
    </div>
  )
}
