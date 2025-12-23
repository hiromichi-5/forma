"use client"

import type { FormResponse, User } from "@/types/form-response"
import { statusConfig } from "@/lib/utils/status-config"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { X, Calendar, Mail, UserIcon, MessageSquare, AlertCircle } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { ja } from "date-fns/locale"

type ResponseDetailPanelProps = {
  response: FormResponse
  users: User[]
  onClose: () => void
  onStatusChange: (status: FormResponse["status"]) => void
  onAssignChange: (userId: string | null) => void
  onPriorityChange: (priority: FormResponse["priority"]) => void
  onOpenChat: () => void
}

export function ResponseDetailPanel({
  response,
  users,
  onClose,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onOpenChat,
}: ResponseDetailPanelProps) {
  return (
    <div className="fixed inset-y-0 right-0 w-full md:w-[500px] bg-background border-l shadow-lg z-50 overflow-y-auto">
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold">回答詳細</h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* 基本情報 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">基本情報</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="text-sm font-medium text-muted-foreground">回答者名</label>
              <p className="text-base font-semibold">{response.respondentName}</p>
            </div>

            <div>
              <label className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Mail className="h-4 w-4" />
                メールアドレス
              </label>
              <p className="text-base">{response.respondentEmail}</p>
            </div>

            <div>
              <label className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                送信日時
              </label>
              <p className="text-base">{formatDistanceToNow(response.submittedAt, { addSuffix: true, locale: ja })}</p>
              <p className="text-sm text-muted-foreground">{response.submittedAt.toLocaleString("ja-JP")}</p>
            </div>

            <div>
              <label className="text-sm font-medium text-muted-foreground">フォーム</label>
              <Badge variant="outline">{response.formTitle}</Badge>
            </div>
          </CardContent>
        </Card>

        {/* ステータス管理 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">ステータス管理</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-2 block">ステータス</label>
              <Select value={response.status} onValueChange={onStatusChange}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="new">
                    <div className="flex items-center gap-2">
                      <span className={`h-2 w-2 rounded-full bg-blue-500`} />
                      新規
                    </div>
                  </SelectItem>
                  <SelectItem value="in_progress">
                    <div className="flex items-center gap-2">
                      <span className={`h-2 w-2 rounded-full bg-yellow-500`} />
                      対応中
                    </div>
                  </SelectItem>
                  <SelectItem value="done">
                    <div className="flex items-center gap-2">
                      <span className={`h-2 w-2 rounded-full bg-green-500`} />
                      完了
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground mt-1">{statusConfig[response.status].description}</p>
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block flex items-center gap-2">
                <UserIcon className="h-4 w-4" />
                担当者
              </label>
              <Select
                value={response.assignedTo || "unassigned"}
                onValueChange={(v) => onAssignChange(v === "unassigned" ? null : v)}
              >
                <SelectTrigger>
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
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block flex items-center gap-2">
                <AlertCircle className="h-4 w-4" />
                優先度
              </label>
              <Select value={response.priority} onValueChange={onPriorityChange}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Low">低</SelectItem>
                  <SelectItem value="Medium">中</SelectItem>
                  <SelectItem value="High">高</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* 回答内容 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">回答内容</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {Object.entries(response.responses).map(([key, value], index) => (
              <div key={key} className="pb-3 border-b last:border-b-0">
                <label className="text-sm font-medium text-muted-foreground">質問 {index + 1}</label>
                <p className="text-base mt-1">{value}</p>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* アクション */}
        <Button onClick={onOpenChat} className="w-full gap-2" size="lg">
          <MessageSquare className="h-5 w-5" />
          チャットで追加質問する
        </Button>
      </div>
    </div>
  )
}
