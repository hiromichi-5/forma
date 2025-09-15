import { useState, useEffect, useMemo } from "react";
import { useParams, Navigate } from "react-router-dom";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { Layout } from "@/components/Layout";
import { Button } from "@/components/ui/Button";
import { Loader, EmptyState, Toast } from "@/components/ui/Common";
import { Card, CardContent } from "@/components/ui/Card";
import { Icon } from "@/components/ui/Icon";
import { apiClient, ApiError } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import type { Ticket, Member, Role, ToastMessage, TicketStatus } from "@/types";
import {
  ArrowLeft,
  Users,
  AlertCircle,
  ChevronRight,
  ChevronLeft,
  Calendar,
  Hash,
} from "lucide-react";

export function KanbanPage() {
  const { form_id } = useParams<{ form_id: string }>();
  const { user } = useAuth();
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [userRole, setUserRole] = useState<Role | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  if (!form_id) {
    return <Navigate to="/" replace />;
  }

  const addToast = (toast: Omit<ToastMessage, "id">) => {
    const id = Math.random().toString(36).substr(2, 9);
    setToasts((prev) => [...prev, { ...toast, id }]);
  };

  const removeToast = (id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  };

  const loadData = async () => {
    try {
      setIsLoading(true);

      const [ticketsResponse, membersResponse] = await Promise.all([
        apiClient.getTickets(form_id),
        apiClient.getMembers(form_id),
      ]);

      setTickets(ticketsResponse.tickets);
      setMembers(membersResponse.members);

      const currentMember = membersResponse.members.find(
        (member) => member.id === user?.id
      );
      setUserRole(currentMember?.role || null);
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.isForbidden) {
          addToast({
            type: "error",
            title: "アクセス権限がありません",
            message: "このフォームにアクセスする権限がありません",
          });
        } else {
          addToast({
            type: "error",
            title: "データの読み込みに失敗しました",
            message: "ページを更新してもう一度お試しください",
          });
        }
      }
    } finally {
      setIsLoading(false);
    }
  };

  const updateTicketStatus = async (
    ticketId: string,
    newStatus: TicketStatus
  ) => {
    try {
      const updatedTicket = await apiClient.updateTicket(ticketId, {
        status: newStatus,
      });

      setTickets((prev) =>
        prev.map((ticket) => (ticket.id === ticketId ? updatedTicket : ticket))
      );

      addToast({
        type: "success",
        title: "ステータスを更新しました",
      });
    } catch (error) {
      if (error instanceof ApiError) {
        addToast({
          type: "error",
          title: "更新に失敗しました",
          message: error.isForbidden
            ? "権限がありません"
            : "ステータスの更新に失敗しました",
        });
      }
    }
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);

    if (!over || !userRole || userRole !== "admin") {
      return;
    }

    const ticketId = active.id as string;
    const newStatus = over.id as TicketStatus;

    const currentTicket = tickets.find((t) => t.id === ticketId);
    if (!currentTicket || currentTicket.status === newStatus) {
      return;
    }

    updateTicketStatus(ticketId, newStatus);
  };

  const ticketsByStatus = useMemo(() => {
    const groups = {
      new: tickets.filter((t) => t.status === "new"),
      in_progress: tickets.filter((t) => t.status === "in_progress"),
      done: tickets.filter((t) => t.status === "done"),
    };
    return groups;
  }, [tickets]);

  useEffect(() => {
    loadData();
  }, [form_id]);

  if (isLoading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-96">
          <Loader size="lg" />
        </div>
      </Layout>
    );
  }

  const columns: Array<{
    id: TicketStatus;
    title: string;
    count: number;
    tickets: Ticket[];
  }> = [
    {
      id: "new",
      title: "新規",
      count: ticketsByStatus.new.length,
      tickets: ticketsByStatus.new,
    },
    {
      id: "in_progress",
      title: "対応中",
      count: ticketsByStatus.in_progress.length,
      tickets: ticketsByStatus.in_progress,
    },
    {
      id: "done",
      title: "完了",
      count: ticketsByStatus.done.length,
      tickets: ticketsByStatus.done,
    },
  ];

  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button
              variant="ghost"
              onClick={() => window.history.back()}
              aria-label="戻る"
            >
              <Icon icon={ArrowLeft} />
            </Button>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">フォーム看板</h1>
              <p className="text-sm text-gray-600">Form ID: {form_id}</p>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            {userRole && (
              <span
                className={`px-2 py-1 text-xs font-medium rounded-full ${
                  userRole === "admin"
                    ? "bg-purple-100 text-purple-800"
                    : "bg-blue-100 text-blue-800"
                }`}
              >
                {userRole === "admin" ? "管理者" : "編集者"}
              </span>
            )}
            <Button variant="secondary" aria-label="メンバー管理">
              <Icon icon={Users} size="sm" className="mr-2" />
              メンバー ({members.length})
            </Button>
          </div>
        </div>

        {userRole !== "admin" && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
            <div className="flex">
              <Icon icon={AlertCircle} size="md" className="text-yellow-400" />
              <div className="ml-3">
                <p className="text-sm text-yellow-700">
                  閲覧モードです。チケットの編集にはadmin権限が必要です。
                </p>
              </div>
            </div>
          </div>
        )}

        <DndContext
          sensors={sensors}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {columns.map((column) => (
              <KanbanColumn
                key={column.id}
                column={column}
                canEdit={userRole === "admin"}
                members={members}
                onUpdateTicket={updateTicketStatus}
              />
            ))}
          </div>

          <DragOverlay>
            {activeId ? (
              <TicketCard
                ticket={tickets.find((t) => t.id === activeId)!}
                canEdit={false}
                members={members}
                onUpdateTicket={() => {}}
                isDragging
              />
            ) : null}
          </DragOverlay>
        </DndContext>

        {tickets.length === 0 && (
          <EmptyState
            title="チケットがありません"
            description="まだチケットが作成されていません。フォームからの回答があると自動的にチケットが作成されます。"
          />
        )}
      </div>

      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          type={toast.type}
          title={toast.title}
          message={toast.message}
          onClose={() => removeToast(toast.id)}
          duration={toast.duration}
        />
      ))}
    </Layout>
  );
}

interface KanbanColumnProps {
  column: {
    id: TicketStatus;
    title: string;
    count: number;
    tickets: Ticket[];
  };
  canEdit: boolean;
  members: Member[];
  onUpdateTicket: (ticketId: string, status: TicketStatus) => void;
}

function KanbanColumn({
  column,
  canEdit,
  members,
  onUpdateTicket,
}: KanbanColumnProps) {
  return (
    <div className="bg-gray-50 rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-medium text-gray-900">{column.title}</h3>
        <span className="bg-gray-200 text-gray-600 px-2 py-1 rounded-full text-sm">
          {column.count}
        </span>
      </div>

      <SortableContext
        items={column.tickets.map((t) => t.id)}
        strategy={verticalListSortingStrategy}
        id={column.id}
      >
        <div className="space-y-3 min-h-32">
          {column.tickets.map((ticket) => (
            <TicketCard
              key={ticket.id}
              ticket={ticket}
              canEdit={canEdit}
              members={members}
              onUpdateTicket={onUpdateTicket}
            />
          ))}
        </div>
      </SortableContext>
    </div>
  );
}

interface TicketCardProps {
  ticket: Ticket;
  canEdit: boolean;
  members: Member[];
  onUpdateTicket: (ticketId: string, status: TicketStatus) => void;
  isDragging?: boolean;
}

function TicketCard({
  ticket,
  canEdit,
  members,
  onUpdateTicket,
  isDragging = false,
}: TicketCardProps) {
  const assignedMember = members.find((m) => m.id === ticket.assignee_id);

  const priorityColors = {
    1: "bg-red-100 text-red-800",
    2: "bg-orange-100 text-orange-800",
    3: "bg-yellow-100 text-yellow-800",
    4: "bg-green-100 text-green-800",
    5: "bg-gray-100 text-gray-800",
  };

  const moveTicket = (direction: "left" | "right") => {
    const statusOrder: TicketStatus[] = ["new", "in_progress", "done"];
    const currentIndex = statusOrder.indexOf(ticket.status);

    let newIndex: number;
    if (direction === "left") {
      newIndex = Math.max(0, currentIndex - 1);
    } else {
      newIndex = Math.min(statusOrder.length - 1, currentIndex + 1);
    }

    if (newIndex !== currentIndex) {
      onUpdateTicket(ticket.id, statusOrder[newIndex]);
    }
  };

  return (
    <Card
      className={`cursor-pointer hover:shadow-md transition-shadow ${
        isDragging ? "opacity-50" : ""
      }`}
    >
      <CardContent className="p-4 space-y-3">
        <div className="flex items-center justify-between">
          <span
            className={`px-2 py-1 rounded-full text-xs font-medium ${
              priorityColors[ticket.priority as keyof typeof priorityColors] ||
              priorityColors[5]
            }`}
          >
            優先度 {ticket.priority}
          </span>
          {canEdit && (
            <div className="flex space-x-1">
              {ticket.status !== "new" && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => moveTicket("left")}
                  aria-label="左に移動"
                >
                  <Icon icon={ChevronLeft} size="sm" />
                </Button>
              )}
              {ticket.status !== "done" && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => moveTicket("right")}
                  aria-label="右に移動"
                >
                  <Icon icon={ChevronRight} size="sm" />
                </Button>
              )}
            </div>
          )}
        </div>

        <div className="space-y-2">
          <div className="flex items-center text-sm text-gray-600">
            <Icon icon={Hash} size="sm" className="mr-1" />
            <span className="font-mono">{ticket.id.slice(0, 8)}</span>
          </div>

          <div className="text-xs text-gray-500 space-y-1">
            <div>Form: {ticket.form_id}</div>
            <div>Response: {ticket.response_id}</div>
          </div>

          <div className="flex items-center text-xs text-gray-500">
            <Icon icon={Calendar} size="sm" className="mr-1" />
            {new Date(ticket.updated_at).toLocaleDateString("ja-JP")}
          </div>
        </div>

        {assignedMember && (
          <div className="flex items-center text-sm">
            <div className="w-6 h-6 bg-gray-300 rounded-full flex items-center justify-center text-xs font-medium text-gray-600 mr-2">
              {assignedMember.email.charAt(0).toUpperCase()}
            </div>
            <span className="text-gray-700 truncate">
              {assignedMember.email}
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
