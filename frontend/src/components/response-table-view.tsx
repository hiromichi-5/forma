"use client"

import { Fragment, useState } from "react"
import type { FormResponse, User } from "@/types/form-response"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { MessageSquare } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { ja } from "date-fns/locale"
import { cn } from "@/lib/utils"

type ResponseTableViewProps = {
  responses: FormResponse[]
  users: User[]
  onStatusChange: (id: string, status: FormResponse["status"]) => void
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: FormResponse["priority"]) => void
  onOpenChat: (response: FormResponse) => void
}

const statusConfig = {
  new: { label: "新規", color: "bg-blue-50 text-blue-700 border-blue-200" },
  in_progress: { label: "対応中", color: "bg-yellow-50 text-yellow-700 border-yellow-200" },
  done: { label: "完了", color: "bg-green-50 text-green-700 border-green-200" },
}

const priorityConfig = {
  Low: { label: "低", color: "text-gray-600" },
  Medium: { label: "中", color: "text-blue-600" },
  High: { label: "高", color: "text-red-600" },
}

const isStatusValue = (value: string): value is FormResponse["status"] =>
  value === "new" || value === "in_progress" || value === "done"

const isPriorityValue = (value: string): value is FormResponse["priority"] =>
  value === "Low" || value === "Medium" || value === "High"

const toPriorityValue = (value: string): FormResponse["priority"] =>
  isPriorityValue(value) ? value : "Medium"

const toStatusValue = (value: string): FormResponse["status"] =>
  isStatusValue(value) ? value : "new"

export function ResponseTableView({
  responses,
  users,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onOpenChat,
}: ResponseTableViewProps) {
  const [expandedRow, setExpandedRow] = useState<string | null>(null)

  return (
    <div className="bg-card rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow className="bg-muted/50 hover:bg-muted/50">
            <TableHead className="w-[200px]">回答者</TableHead>
            <TableHead className="w-[150px]">ステータス</TableHead>
            <TableHead className="w-[150px]">担当者</TableHead>
            <TableHead className="w-[100px]">優先度</TableHead>
            <TableHead className="w-[150px]">送信日時</TableHead>
            <TableHead className="w-[100px]">アクション</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {responses.map((response) => (
            <Fragment key={response.id}>
              <TableRow
                className="cursor-pointer hover:bg-muted/50 transition-colors"
                onClick={() => setExpandedRow(expandedRow === response.id ? null : response.id)}
              >
                <TableCell>
                  <div>
                    <p className="font-medium text-foreground">{response.respondentName}</p>
                    <p className="text-sm text-muted-foreground">{response.respondentEmail}</p>
                  </div>
                </TableCell>
                <TableCell onClick={(e) => e.stopPropagation()}>
                  <Select
                    value={response.status}
                    onValueChange={(value) => onStatusChange(response.id, toStatusValue(value))}
                  >
                    <SelectTrigger className="w-[130px] h-8">
                      <Badge variant="outline" className={cn("border", statusConfig[response.status].color)}>
                        {statusConfig[response.status].label}
                      </Badge>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="new">新規</SelectItem>
                      <SelectItem value="in_progress">対応中</SelectItem>
                      <SelectItem value="done">完了</SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
                <TableCell onClick={(e) => e.stopPropagation()}>
                  <Select
                    value={response.assignedTo || "unassigned"}
                    onValueChange={(value) => onAssignChange(response.id, value === "unassigned" ? null : value)}
                  >
                    <SelectTrigger className="w-[130px] h-8">
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
                </TableCell>
                <TableCell onClick={(e) => e.stopPropagation()}>
                  <Select
                    value={response.priority}
                    onValueChange={(value) => onPriorityChange(response.id, toPriorityValue(value))}
                  >
                    <SelectTrigger className="w-[90px] h-8">
                      <span className={cn("font-medium", priorityConfig[response.priority].color)}>
                        {priorityConfig[response.priority].label}
                      </span>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Low">低</SelectItem>
                      <SelectItem value="Medium">中</SelectItem>
                      <SelectItem value="High">高</SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
                <TableCell>
                  <span className="text-sm text-muted-foreground">
                    {formatDistanceToNow(response.submittedAt, { addSuffix: true, locale: ja })}
                  </span>
                </TableCell>
                <TableCell onClick={(e) => e.stopPropagation()}>
                  <Button variant="ghost" size="sm" onClick={() => onOpenChat(response)} className="gap-2 h-8">
                    <MessageSquare className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
              {expandedRow === response.id && (
                <TableRow className="hover:bg-muted/20">
                  <TableCell colSpan={6} className="bg-muted/20 p-4">
                    <div className="space-y-3">
                      <h4 className="font-semibold text-sm text-foreground">回答内容</h4>
                      {Object.entries(response.responses).map(([key, value], index) => (
                        <div key={key} className="text-sm">
                          <span className="text-muted-foreground">質問 {index + 1}: </span>
                          <span className="text-foreground">{value}</span>
                        </div>
                      ))}
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </Fragment>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
