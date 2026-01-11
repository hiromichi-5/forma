"use client";

import type { FormResponse, User } from "@/types/form-response";
import type { FormStatus } from "@/types";
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
  responses: FormResponse[];
  users: User[];
  statuses: FormStatus[];
  onStatusChange: (id: string, statusId: string) => void;
  onAssignChange: (id: string, userId: string | null) => void;
  onPriorityChange: (id: string, priority: FormResponse["priority"]) => void;
  onOpenDetail: (response: FormResponse) => void;
};

const priorityConfig = {
  low: { label: "低", color: "text-gray-600", hex: "#6B7280" },
  medium: { label: "中", color: "text-blue-600", hex: "#2563EB" },
  high: { label: "高", color: "text-red-600", hex: "#DC2626" },
};

const isPriorityValue = (value: string): value is FormResponse["priority"] =>
  value === "low" || value === "medium" || value === "high";

const toPriorityValue = (value: string): FormResponse["priority"] =>
  isPriorityValue(value) ? value : "medium";

const hexToRgba = (hex: string | null | undefined, alpha: number): string => {
  if (!hex) return `rgba(156, 163, 175, ${alpha})`;
  const cleanHex = hex.replace("#", "");
  const r = parseInt(cleanHex.slice(0, 2), 16);
  const g = parseInt(cleanHex.slice(2, 4), 16);
  const b = parseInt(cleanHex.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
};

type DraggableCardProps = {
  response: FormResponse;
  users: User[];
  onAssignChange: (id: string, userId: string | null) => void;
  onPriorityChange: (id: string, priority: FormResponse["priority"]) => void;
  onOpenDetail: (response: FormResponse) => void;
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

  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={cn(
        "p-4 border hover:bg-muted/30 transition-colors cursor-pointer",
        isDragging && "opacity-50"
      )}
      onClick={() => onOpenDetail(response)}
      {...attributes}
      {...listeners}
    >
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1">
            <h4 className="font-semibold text-sm text-foreground mb-1">
              {response.respondentEmail}
            </h4>
          </div>
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

        <div className="text-xs text-muted-foreground line-clamp-2">
          {Object.values(response.responses)[0]}
        </div>

        <Select
          value={response.assignedTo || "unassigned"}
          onValueChange={(value) =>
            onAssignChange(response.id, value === "unassigned" ? null : value)
          }
        >
          <SelectTrigger
            className="w-full h-8 text-xs"
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

        <div className="flex items-center justify-between pt-2 border-t">
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Calendar className="h-3 w-3" />
            <span>
              {formatDistanceToNow(response.submittedAt, { locale: ja })}
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
    <div className="flex flex-col gap-3">
      <div
        className="flex items-center justify-between p-3 rounded-lg"
        style={{
          backgroundColor: hexToRgba(statusColor, 0.1),
        }}
      >
        <div className="flex items-center gap-2">
          <div
            className="w-2 h-2 rounded-full shrink-0"
            style={{
              backgroundColor: statusColor || "#9CA3AF",
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
          "space-y-3 min-h-[500px] p-2 rounded-lg transition-colors",
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

  const sortedStatuses = [...statuses].sort(
    (a, b) => a.display_order - b.display_order
  );

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
    if (response && response.status !== newStatusId) {
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
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-w-7xl mx-auto">
        {sortedStatuses.map((status) => (
          <DroppableColumn
            key={status.id}
            statusId={status.id}
            statusName={status.name}
            statusColor={status.color}
            count={responses.filter((r) => r.status === status.id).length}
          >
            {responses
              .filter((r) => r.status === status.id)
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
                    {activeResponse.respondentEmail}
                  </h4>
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
              <div className="text-xs text-muted-foreground line-clamp-2">
                {Object.values(activeResponse.responses)[0]}
              </div>
            </div>
          </Card>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
