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
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  closestCorners,
  useDroppable,
  useDraggable,
} from "@dnd-kit/core"
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core"
import { useState } from "react"

type ResponseKanbanViewProps = {
  responses: FormResponse[]
  users: User[]
  onStatusChange: (id: string, status: FormResponse["status"]) => void
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: FormResponse["priority"]) => void
  onOpenChat: (response: FormResponse) => void
}

const statusConfig = {
  new: { label: "新規", color: "bg-blue-50" },
  in_progress: { label: "対応中", color: "bg-yellow-50" },
  done: { label: "完了", color: "bg-green-50" },
}

const priorityConfig = {
  Low: { label: "低", color: "text-gray-600" },
  Medium: { label: "中", color: "text-blue-600" },
  High: { label: "高", color: "text-red-600" },
}

const isPriorityValue = (value: string): value is FormResponse["priority"] =>
  value === "Low" || value === "Medium" || value === "High"

const toPriorityValue = (value: string): FormResponse["priority"] =>
  isPriorityValue(value) ? value : "Medium"

type DraggableCardProps = {
  response: FormResponse
  users: User[]
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: FormResponse["priority"]) => void
  onOpenChat: (response: FormResponse) => void
}

function DraggableCard({ response, users, onAssignChange, onPriorityChange, onOpenChat }: DraggableCardProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: response.id,
    data: {
      response,
    },
  })

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
      }
    : undefined

  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={cn(
        "p-4 border hover:bg-muted/30 transition-colors cursor-grab active:cursor-grabbing",
        isDragging && "opacity-50"
      )}
      {...attributes}
      {...listeners}
    >
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1">
            <h4 className="font-semibold text-sm text-foreground mb-1">{response.respondentName}</h4>
            <p className="text-xs text-muted-foreground truncate">{response.respondentEmail}</p>
          </div>
          <Select
            value={response.priority}
            onValueChange={(value) => onPriorityChange(response.id, toPriorityValue(value))}
          >
            <SelectTrigger className="w-[80px] h-7 text-xs shrink-0" onClick={(e) => e.stopPropagation()}>
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
        </div>

        <div className="text-xs text-muted-foreground line-clamp-2">{Object.values(response.responses)[0]}</div>

        <Select
          value={response.assignedTo || "unassigned"}
          onValueChange={(value) => onAssignChange(response.id, value === "unassigned" ? null : value)}
        >
          <SelectTrigger className="w-full h-8 text-xs" onClick={(e) => e.stopPropagation()}>
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
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              onOpenChat(response)
            }}
            className="gap-1 h-7 px-2"
          >
            <MessageSquare className="h-3 w-3" />
          </Button>
        </div>
      </div>
    </Card>
  )
}

type DroppableColumnProps = {
  status: FormResponse["status"]
  children: React.ReactNode
  count: number
}

function DroppableColumn({ status, children, count }: DroppableColumnProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: status,
  })

  return (
    <div className="flex flex-col gap-3">
      <div className={cn("flex items-center justify-between p-3 rounded-lg", statusConfig[status].color)}>
        <h3 className="font-semibold text-sm text-foreground">{statusConfig[status].label}</h3>
        <Badge variant="secondary" className="bg-card">
          {count}
        </Badge>
      </div>

      <div
        ref={setNodeRef}
        className={cn("space-y-3 min-h-[500px] p-2 rounded-lg transition-colors", isOver && "bg-muted/50")}
      >
        {children}
      </div>
    </div>
  )
}

export function ResponseKanbanView({
  responses,
  users,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onOpenChat,
}: ResponseKanbanViewProps) {
  const columns: Array<FormResponse["status"]> = ["new", "in_progress", "done"]
  const [activeId, setActiveId] = useState<string | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  )

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string)
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event

    if (!over) {
      setActiveId(null)
      return
    }

    const responseId = active.id as string
    const newStatus = over.id as FormResponse["status"]

    const response = responses.find((r) => r.id === responseId)
    if (response && response.status !== newStatus) {
      onStatusChange(responseId, newStatus)
    }

    setActiveId(null)
  }

  const activeResponse = activeId ? responses.find((r) => r.id === activeId) : null

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-w-7xl mx-auto">
        {columns.map((status) => (
          <DroppableColumn key={status} status={status} count={responses.filter((r) => r.status === status).length}>
            {responses
              .filter((r) => r.status === status)
              .map((response) => (
                <DraggableCard
                  key={response.id}
                  response={response}
                  users={users}
                  onAssignChange={onAssignChange}
                  onPriorityChange={onPriorityChange}
                  onOpenChat={onOpenChat}
                />
              ))}
          </DroppableColumn>
        ))}
      </div>

      <DragOverlay>
        {activeResponse ? (
          <Card className="p-4 border bg-card shadow-lg rotate-3">
            <div className="space-y-3">
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1">
                  <h4 className="font-semibold text-sm text-foreground mb-1">{activeResponse.respondentName}</h4>
                  <p className="text-xs text-muted-foreground truncate">{activeResponse.respondentEmail}</p>
                </div>
                <span className={cn("font-medium text-xs", priorityConfig[activeResponse.priority].color)}>
                  {priorityConfig[activeResponse.priority].label}
                </span>
              </div>
              <div className="text-xs text-muted-foreground line-clamp-2">
                {Object.values(activeResponse.responses)[0]}
              </div>
            </div>
          </Card>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}
