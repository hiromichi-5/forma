"use client";

import type { User } from "@/types/form-response";
import type { FormStatus, TicketDetail, TicketPriority } from "@/types";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MessageSquare, Calendar } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { ja } from "date-fns/locale";
import { cn } from "@/lib/utils";
import {
  fallbackStatusColor,
  hexToRgba,
  priorityConfig,
  respondentEmailLabel,
  sortStatuses,
  toPriorityValue,
} from "@/lib/ticket-display";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  closestCorners,
  useDroppable,
  useDraggable,
} from "@dnd-kit/core";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import { useState } from "react";

type ResponseKanbanViewProps = {
  responses: TicketDetail[];
  users: User[];
  statuses: FormStatus[];
  onStatusChange: (id: string, statusId: string) => void;
  onAssignChange: (id: string, userId: string | null) => void;
  onPriorityChange: (id: string, priority: TicketPriority) => void;
  onOpenDetail: (response: TicketDetail) => void;
};

type DraggableCardProps = {
  response: TicketDetail;
  users: User[];
  onAssignChange: (id: string, userId: string | null) => void;
  onPriorityChange: (id: string, priority: TicketPriority) => void;
  onOpenDetail: (response: TicketDetail) => void;
};

function DraggableCard({
  response,
  users,
  onAssignChange,
  onPriorityChange,
  onOpenDetail,
}: DraggableCardProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({
      id: response.id,
      data: {
        response,
      },
    });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
      }
    : undefined;

  const email = respondentEmailLabel(response.respondent_email);

  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={cn(
        "cursor-grab border p-3 shadow-none transition-colors hover:bg-muted/30 active:cursor-grabbing",
        isDragging && "opacity-50"
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
            onValueChange={(value) =>
              onPriorityChange(response.id, toPriorityValue(value))
            }
          >
            <SelectTrigger
              className="w-[80px] h-7 text-xs shrink-0 border-0 shadow-none"
              onClick={(e) => e.stopPropagation()}
            >
              <div
                className="flex items-center gap-1.5 px-2 py-1 rounded"
                style={{
                  backgroundColor: hexToRgba(
                    priorityConfig[response.priority].hex,
                    0.1
                  ),
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
          <SelectTrigger
            className="w-full h-7 text-xs"
            onClick={(e) => e.stopPropagation()}
          >
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

        <div className="flex items-center justify-between pt-1.5 border-t">
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Calendar className="h-3 w-3" />
            <span>
              {formatDistanceToNow(new Date(response.submitted_at), { locale: ja })}
            </span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation();
              onOpenDetail(response);
            }}
            className="gap-1 h-7 px-2"
            aria-label={`回答詳細を開く: ${email}`}
          >
            <MessageSquare className="h-3 w-3" />
          </Button>
        </div>
      </div>
    </Card>
  );
}

type DroppableColumnProps = {
  statusId: string;
  statusName: string;
  statusColor?: string | null;
  children: React.ReactNode;
  count: number;
};

function DroppableColumn({
  statusId,
  statusName,
  statusColor,
  children,
  count,
}: DroppableColumnProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: statusId,
  });

  return (
    <div className="flex w-60 shrink-0 flex-col gap-2">
      <div
        className="flex items-center justify-between p-2 rounded-lg"
        style={{
          backgroundColor: hexToRgba(statusColor, 0.1),
        }}
      >
        <div className="flex items-center gap-2">
          <div
            className="w-2 h-2 rounded-full shrink-0"
            style={{
              backgroundColor: statusColor || fallbackStatusColor,
            }}
          />
          <h3 className="font-semibold text-sm text-foreground">
            {statusName}
          </h3>
        </div>
        <Badge variant="secondary" className="bg-card">
          {count}
        </Badge>
      </div>

      <div
        ref={setNodeRef}
        className={cn(
          "space-y-2 min-h-[160px] p-2 rounded-lg transition-colors",
          isOver && "bg-muted/50"
        )}
      >
        {children}
      </div>
    </div>
  );
}

export function ResponseKanbanView({
  responses,
  users,
  statuses,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onOpenDetail,
}: ResponseKanbanViewProps) {
  const [activeId, setActiveId] = useState<string | null>(null);

  const sortedStatuses = sortStatuses(statuses);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over) {
      setActiveId(null);
      return;
    }

    const responseId = active.id as string;
    const newStatusId = over.id as string;

    const response = responses.find((r) => r.id === responseId);
    if (response && response.status.id !== newStatusId) {
      onStatusChange(responseId, newStatusId);
    }

    setActiveId(null);
  };

  const activeResponse = activeId
    ? responses.find((r) => r.id === activeId)
    : null;

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-3 overflow-x-auto pb-2">
        {sortedStatuses.map((status) => (
          <DroppableColumn
            key={status.id}
            statusId={status.id}
            statusName={status.name}
            statusColor={status.color}
            count={responses.filter((r) => r.status.id === status.id).length}
          >
            {responses
              .filter((r) => r.status.id === status.id)
              .map((response) => (
                <DraggableCard
                  key={response.id}
                  response={response}
                  users={users}
                  onAssignChange={onAssignChange}
                  onPriorityChange={onPriorityChange}
                  onOpenDetail={onOpenDetail}
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
                  <h4 className="font-semibold text-sm text-foreground mb-1">
                    {activeResponse.title}
                  </h4>
                  <p className="text-xs text-muted-foreground">
                    {respondentEmailLabel(activeResponse.respondent_email)}
                  </p>
                </div>
                <span
                  className={cn(
                    "font-medium text-xs",
                    priorityConfig[activeResponse.priority].color
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
  );
}
