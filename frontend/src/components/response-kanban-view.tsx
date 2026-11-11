import type { FormStatus, Member, TicketSummary, TicketPriority } from "@/types"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { MessageSquare, Calendar } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { ja } from "date-fns/locale"
import { cn } from "@/lib/utils"
import {
  fallbackStatusColor,
  hexToRgba,
  priorityConfig,
  respondentEmailLabel,
  toPriorityValue,
} from "@/lib/ticket-display"
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
import { useQuery } from "@tanstack/react-query"
import { InfiniteScrollTrigger } from "@/components/infinite-scroll-trigger"
import { useTicketList } from "@/hooks/use-ticket-list"
import { ticketQueries } from "@/queries/tickets"

type ResponseKanbanViewProps = {
  formId: string
  query: string
  /** 表示する列。ステータスの絞り込みを適用済みのもの。 */
  columns: FormStatus[]
  members: Member[]
  onStatusChange: (id: string, statusId: string) => void
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: TicketPriority) => void
  onOpenDetail: (response: TicketSummary) => void
}

type DraggableCardProps = {
  response: TicketSummary
  members: Member[]
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: TicketPriority) => void
  onOpenDetail: (response: TicketSummary) => void
}

function DraggableCard({
  response,
  members,
  onAssignChange,
  onPriorityChange,
  onOpenDetail,
}: DraggableCardProps) {
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

  const email = respondentEmailLabel(response.respondent_email)

  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={cn(
        "cursor-grab border p-3 shadow-none transition-colors hover:bg-muted/30 active:cursor-grabbing",
        isDragging && "opacity-50",
      )}
      {...attributes}
      {...listeners}
    >
      <div className="space-y-2">
        <div className="flex items-start justify-between gap-2">
          <Button
            variant="ghost"
            type="button"
            className="h-auto min-w-0 flex-1 justify-start p-0 text-left hover:bg-transparent"
            onClick={() => onOpenDetail(response)}
            aria-label={`回答詳細を開く: ${response.title}`}
          >
            <div className="min-w-0">
              <h4 className="mb-0.5 truncate text-sm font-semibold text-foreground">
                {response.title}
              </h4>
              <p className="truncate text-xs text-muted-foreground">{email}</p>
            </div>
          </Button>
          <Select
            value={response.priority}
            onValueChange={(value) => onPriorityChange(response.id, toPriorityValue(value))}
          >
            <SelectTrigger
              className="h-7 w-[80px] shrink-0 border-0 text-xs shadow-none"
              onClick={(e) => e.stopPropagation()}
            >
              <div
                className="flex items-center gap-1.5 rounded px-2 py-1"
                style={{
                  backgroundColor: hexToRgba(priorityConfig[response.priority].hex, 0.1),
                }}
              >
                <span
                  className="text-xs font-medium"
                  style={{
                    color: priorityConfig[response.priority].hex,
                  }}
                >
                  {priorityConfig[response.priority].label}
                </span>
              </div>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="low">低</SelectItem>
              <SelectItem value="medium">中</SelectItem>
              <SelectItem value="high">高</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <Select
          value={response.assignee?.id ?? "unassigned"}
          onValueChange={(value) =>
            onAssignChange(response.id, value === "unassigned" ? null : value)
          }
        >
          <SelectTrigger className="h-7 w-full text-xs" onClick={(e) => e.stopPropagation()}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="unassigned">未割当</SelectItem>
            {members.map((member) => (
              <SelectItem key={member.id} value={member.id}>
                {member.display_name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <div className="flex items-center justify-between border-t pt-1.5">
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Calendar className="h-3 w-3" />
            <span>{formatDistanceToNow(new Date(response.submitted_at), { locale: ja })}</span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              onOpenDetail(response)
            }}
            className="h-7 gap-1 px-2"
            aria-label={`回答詳細を開く: ${email}`}
          >
            <MessageSquare className="h-3 w-3" />
          </Button>
        </div>
      </div>
    </Card>
  )
}

type DroppableColumnProps = {
  statusId: string
  statusName: string
  statusColor?: string | null
  children: React.ReactNode
  count: number
}

function DroppableColumn({
  statusId,
  statusName,
  statusColor,
  children,
  count,
}: DroppableColumnProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: statusId,
  })

  return (
    <div className="flex w-60 shrink-0 flex-col gap-2">
      <div
        className="flex items-center justify-between rounded-lg p-2"
        style={{
          backgroundColor: hexToRgba(statusColor, 0.1),
        }}
      >
        <div className="flex items-center gap-2">
          <div
            className="h-2 w-2 shrink-0 rounded-full"
            style={{
              backgroundColor: statusColor || fallbackStatusColor,
            }}
          />
          <h3 className="text-sm font-semibold text-foreground">{statusName}</h3>
        </div>
        <Badge variant="secondary" className="bg-card">
          {count}
        </Badge>
      </div>

      <div
        ref={setNodeRef}
        className={cn(
          "min-h-[160px] space-y-2 rounded-lg p-2 transition-colors",
          isOver && "bg-muted/50",
        )}
      >
        {children}
      </div>
    </div>
  )
}

export function ResponseKanbanView({
  formId,
  query,
  columns,
  members,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onOpenDetail,
}: ResponseKanbanViewProps) {
  const [activeResponse, setActiveResponse] = useState<TicketSummary | null>(null)
  const { data: counts } = useQuery(ticketQueries.counts(formId, query))

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
  )

  // ドラッグ中のカードは列ごとに別々の一覧にあるため、ドラッグ元のカードが持つデータを使う。
  const draggedResponse = (event: DragStartEvent | DragEndEvent) =>
    (event.active.data.current as { response?: TicketSummary } | undefined)?.response ?? null

  const handleDragStart = (event: DragStartEvent) => {
    setActiveResponse(draggedResponse(event))
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const response = draggedResponse(event)
    const newStatusId = event.over?.id
    if (response && typeof newStatusId === "string" && response.status.id !== newStatusId) {
      onStatusChange(response.id, newStatusId)
    }
    setActiveResponse(null)
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-3 overflow-x-auto pb-2">
        {columns.map((status) => (
          <KanbanColumn
            key={status.id}
            formId={formId}
            query={query}
            status={status}
            count={counts?.[status.id] ?? 0}
            members={members}
            onAssignChange={onAssignChange}
            onPriorityChange={onPriorityChange}
            onOpenDetail={onOpenDetail}
          />
        ))}
      </div>

      <DragOverlay>
        {activeResponse ? (
          <Card className="rotate-3 border bg-card p-4 shadow-lg">
            <div className="space-y-3">
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1">
                  <h4 className="mb-1 text-sm font-semibold text-foreground">
                    {activeResponse.title}
                  </h4>
                  <p className="text-xs text-muted-foreground">
                    {respondentEmailLabel(activeResponse.respondent_email)}
                  </p>
                </div>
                <span
                  className={cn(
                    "text-xs font-medium",
                    priorityConfig[activeResponse.priority].color,
                  )}
                >
                  {priorityConfig[activeResponse.priority].label}
                </span>
              </div>
            </div>
          </Card>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}

type KanbanColumnProps = {
  formId: string
  query: string
  status: FormStatus
  count: number
  members: Member[]
  onAssignChange: (id: string, userId: string | null) => void
  onPriorityChange: (id: string, priority: TicketPriority) => void
  onOpenDetail: (response: TicketSummary) => void
}

/** 1 つのステータスの列。列ごとに独立して無限スクロールする。 */
function KanbanColumn({ formId, query, status, count, ...cardProps }: KanbanColumnProps) {
  const { tickets, hasMore, loading, loadMore } = useTicketList(formId, {
    statusIds: [status.id],
    query,
  })

  return (
    <DroppableColumn
      statusId={status.id}
      statusName={status.name}
      statusColor={status.color}
      count={count}
    >
      {tickets.map((response) => (
        <DraggableCard key={response.id} response={response} {...cardProps} />
      ))}
      <InfiniteScrollTrigger hasMore={hasMore} loading={loading} onLoadMore={loadMore} />
    </DroppableColumn>
  )
}
