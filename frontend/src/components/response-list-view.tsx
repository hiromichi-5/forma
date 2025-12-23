"use client"

import type { FormResponse, User } from "@/types/form-response"
import { statusConfig, priorityConfig } from "@/lib/utils/status-config"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MessageSquare, UserIcon, Calendar } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { ja } from "date-fns/locale"

interface ResponseListViewProps {
  responses: FormResponse[]
  users: User[]
  onSelectResponse: (response: FormResponse) => void
}

export function ResponseListView({ responses, users, onSelectResponse }: ResponseListViewProps) {
  const getUserName = (userId: string | null) => {
    if (!userId) return "未割当"
    return users.find((u) => u.id === userId)?.name || "不明"
  }

  return (
    <div className="space-y-3">
      {responses.map((response) => (
        <div
          key={response.id}
          className="bg-card border rounded-lg p-4 hover:border-primary/50 transition-colors cursor-pointer"
          onClick={() => onSelectResponse(response)}
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 space-y-2">
              <div className="flex items-center gap-2 flex-wrap">
                <h3 className="font-semibold text-lg">{response.respondentName}</h3>
                <Badge variant="outline" className="text-xs">
                  {response.formTitle}
                </Badge>
                <Badge className={statusConfig[response.status].color}>{statusConfig[response.status].label}</Badge>
                <span className={`text-sm font-medium ${priorityConfig[response.priority].color}`}>
                  優先度: {priorityConfig[response.priority].label}
                </span>
              </div>

              <div className="text-sm text-muted-foreground space-y-1">
                <p className="flex items-center gap-2">
                  <UserIcon className="h-4 w-4" />
                  {response.respondentEmail}
                </p>
                <p className="flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  {formatDistanceToNow(response.submittedAt, { addSuffix: true, locale: ja })}
                </p>
              </div>

              <div className="text-sm">
                {Object.entries(response.responses)
                  .slice(0, 2)
                  .map(([key, value]) => (
                    <p key={key} className="text-muted-foreground truncate">
                      {value}
                    </p>
                  ))}
              </div>
            </div>

            <div className="flex flex-col items-end gap-2">
              <div className="flex items-center gap-2 text-sm">
                <UserIcon className="h-4 w-4" />
                <span className="font-medium">{getUserName(response.assignedTo)}</span>
              </div>
              <Button variant="outline" size="sm" className="gap-2 bg-transparent">
                <MessageSquare className="h-4 w-4" />
                チャット
              </Button>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
