import React, { useState, useEffect, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  DragDropContext,
  Droppable,
  Draggable,
  type DropResult,
} from "@hello-pangea/dnd";
import { motion } from "framer-motion";
import {
  ListChecks,
  Clock,
  CheckCircle2,
  AlertCircle,
  ArrowLeft,
  Users,
  Calendar,
  Hash,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { Button } from "../components/ui/Button";
import { CardContent } from "../components/ui/Card";
import { Badge } from "../components/ui/Badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../components/ui/Select";
import { apiClient, ApiError } from "../lib/api";
import { useAuth } from "../hooks/useAuth";
import type { Ticket, Member, FormSummary } from "../types";

interface KanbanColumn {
  id: Ticket["status"];
  title: string;
  icon: React.ElementType;
  color: string;
}

const columns: KanbanColumn[] = [
  {
    id: "new",
    title: "新規",
    icon: AlertCircle,
    color: "bg-blue-50 border-blue-200 border-2 border-dashed",
  },
  {
    id: "in_progress",
    title: "対応中",
    icon: Clock,
    color: "bg-amber-50 border-amber-200 border-2 border-dashed",
  },
  {
    id: "done",
    title: "完了",
    icon: CheckCircle2,
    color: "bg-emerald-50 border-emerald-200 border-2 border-dashed",
  },
];

const priorityColors = {
  1: "bg-red-100 text-red-800",
  2: "bg-orange-100 text-orange-800",
  3: "bg-yellow-100 text-yellow-800",
  4: "bg-green-100 text-green-800",
  5: "bg-gray-100 text-gray-800",
};

interface TicketCardProps {
  ticket: Ticket;
  index: number;
  members: Member[];
  canEdit: boolean;
  onUpdateTicket: (ticketId: string, status: Ticket["status"]) => void;
}

function TicketCard({
  ticket,
  index,
  members,
  canEdit,
  onUpdateTicket,
}: TicketCardProps) {
  const assignedMember = members.find((m) => m.id === ticket.assignee_id);
  const priorityColor =
    priorityColors[ticket.priority as keyof typeof priorityColors] ||
    priorityColors[5];

  const moveTicket = (direction: "left" | "right") => {
    return () => {
      const statusOrder: Ticket["status"][] = ["new", "in_progress", "done"];
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
  };

  return (
    <Draggable draggableId={ticket.id} index={index} isDragDisabled={!canEdit}>
      {(provided, snapshot) => (
        <div
          ref={provided.innerRef}
          {...provided.draggableProps}
          {...provided.dragHandleProps}
        >
          <motion.div
            layout
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            whileHover={{ scale: canEdit ? 1.02 : 1 }}
            className={`bg-white rounded-lg border shadow-sm mb-3 ${
              canEdit ? "cursor-grab active:cursor-grabbing" : "cursor-default"
            } ${snapshot.isDragging ? "shadow-lg rotate-2" : ""}`}
          >
            <CardContent className="p-4 space-y-3">
              <div className="flex items-center justify-between">
                <Badge className={`text-xs ${priorityColor}`}>
                  優先度 {ticket.priority}
                </Badge>
                {canEdit && (
                  <div className="flex space-x-1">
                    {ticket.status !== "new" && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={moveTicket("left")}
                        className="h-6 w-6 p-0"
                      >
                        <ChevronLeft className="h-3 w-3" />
                      </Button>
                    )}
                    {ticket.status !== "done" && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={moveTicket("right")}
                        className="h-6 w-6 p-0"
                      >
                        <ChevronRight className="h-3 w-3" />
                      </Button>
                    )}
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <div className="flex items-center text-sm text-gray-600">
                  <Hash className="h-3 w-3 mr-1" />
                  <span className="font-mono">{ticket.id.slice(0, 8)}</span>
                </div>

                <div className="text-xs text-gray-500 space-y-1">
                  <div>Form: {ticket.form_id}</div>
                  <div>Response: {ticket.response_id}</div>
                </div>

                <div className="flex items-center text-xs text-gray-500">
                  <Calendar className="h-3 w-3 mr-1" />
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
          </motion.div>
        </div>
      )}
    </Draggable>
  );
}

interface KanbanColumnProps {
  column: KanbanColumn;
  tickets: Ticket[];
  members: Member[];
  canEdit: boolean;
  onUpdateTicket: (ticketId: string, status: Ticket["status"]) => void;
}

function KanbanColumn({
  column,
  tickets,
  members,
  canEdit,
  onUpdateTicket,
}: KanbanColumnProps) {
  const Icon = column.icon;

  return (
    <div className={`rounded-lg p-4 ${column.color}`}>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Icon className="h-5 w-5" />
          <h3 className="font-semibold">{column.title}</h3>
        </div>
        <Badge variant="secondary" className="text-xs">
          {tickets.length}
        </Badge>
      </div>

      <Droppable droppableId={column.id}>
        {(provided, snapshot) => (
          <div
            ref={provided.innerRef}
            {...provided.droppableProps}
            className={`min-h-[300px] space-y-2 ${
              snapshot.isDraggingOver ? "bg-white/50 rounded-lg p-2" : ""
            }`}
          >
            {tickets.map((ticket, index) => (
              <TicketCard
                key={ticket.id}
                ticket={ticket}
                index={index}
                members={members}
                canEdit={canEdit}
                onUpdateTicket={onUpdateTicket}
              />
            ))}
            {provided.placeholder}
            {tickets.length === 0 && (
              <div className="text-center py-12 text-muted-foreground text-sm">
                チケットなし
              </div>
            )}
          </div>
        )}
      </Droppable>
    </div>
  );
}

export function KanbanPage() {
  const { form_id } = useParams<{ form_id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [selectedForm, setSelectedForm] = useState<string>("all");
  const [userRole, setUserRole] = useState<"admin" | "editor" | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const canEdit = userRole === "admin";

  const loadData = async (formId?: string) => {
    try {
      setIsLoading(true);
      setError(null);

      const [formsResponse, ticketsResponse] = await Promise.all([
        apiClient.getForms(),
        apiClient.getTickets(formId === "all" ? undefined : formId),
      ]);

      setForms(formsResponse.forms);
      setTickets(ticketsResponse.tickets);

      // Get user role for specific form if formId is provided and not "all"
      if (formId && formId !== "all") {
        try {
          const membersResponse = await apiClient.getMembers(formId);
          setMembers(membersResponse.members);
          const currentMember = membersResponse.members.find(
            (m) => m.id === user?.id
          );
          setUserRole(currentMember?.role || null);
        } catch (err) {
          console.error("Failed to load members:", err);
        }
      } else {
        // Reset role and members when viewing all forms
        setMembers([]);
        setUserRole(null);
      }
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.isForbidden) {
          setError("このフォームにアクセスする権限がありません");
        } else {
          setError("データの読み込みに失敗しました");
        }
      } else {
        setError("ネットワークエラーが発生しました");
      }
      console.error("Failed to load kanban data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFormSelection = (value: string) => {
    if (value === "all") {
      navigate("/kanban");
    } else {
      navigate(`/kanban/${value}`);
    }
  };

  useEffect(() => {
    if (form_id) {
      loadData(form_id);
      setSelectedForm(form_id);
    } else {
      loadData(selectedForm);
    }
  }, [form_id, selectedForm, user?.id]);

  const handleDragEnd = async (result: DropResult) => {
    const { destination, source, draggableId } = result;

    if (!destination || !canEdit) {
      return;
    }

    if (
      destination.droppableId === source.droppableId &&
      destination.index === source.index
    ) {
      return;
    }

    const ticket = tickets.find((t) => t.id === draggableId);
    if (!ticket) return;

    const newStatus = destination.droppableId as Ticket["status"];

    // Optimistically update UI
    setTickets((prev) =>
      prev.map((t) =>
        t.id === draggableId
          ? { ...t, status: newStatus, updated_at: new Date().toISOString() }
          : t
      )
    );

    try {
      await apiClient.updateTicket(draggableId, { status: newStatus });
    } catch (err) {
      // Revert on error
      setTickets((prev) =>
        prev.map((t) =>
          t.id === draggableId ? { ...t, status: ticket.status } : t
        )
      );
      setError("チケットの更新に失敗しました");
      console.error("Failed to update ticket:", err);
    }
  };

  const updateTicketStatus = async (
    ticketId: string,
    newStatus: Ticket["status"]
  ) => {
    const ticket = tickets.find((t) => t.id === ticketId);
    if (!ticket) return;

    setTickets((prev) =>
      prev.map((t) =>
        t.id === ticketId
          ? { ...t, status: newStatus, updated_at: new Date().toISOString() }
          : t
      )
    );

    try {
      await apiClient.updateTicket(ticketId, { status: newStatus });
    } catch (err) {
      setTickets((prev) =>
        prev.map((t) =>
          t.id === ticketId ? { ...t, status: ticket.status } : t
        )
      );
      setError("チケットの更新に失敗しました");
      console.error("Failed to update ticket:", err);
    }
  };

  const ticketsByStatus = useMemo(() => {
    return {
      new: tickets.filter((t) => t.status === "new"),
      in_progress: tickets.filter((t) => t.status === "in_progress"),
      done: tickets.filter((t) => t.status === "done"),
    };
  }, [tickets]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">読み込み中...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {form_id && (
            <Button
              variant="ghost"
              onClick={() => window.history.back()}
              aria-label="戻る"
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>
          )}
          <div className="rounded-lg bg-primary/10 p-2">
            <ListChecks className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">看板</h1>
            <p className="text-muted-foreground">
              チケットをドラッグ&ドロップで管理
              {form_id && ` - ${form_id}`}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-4">
          {userRole && (
            <Badge
              className={
                userRole === "admin"
                  ? "bg-purple-100 text-purple-800"
                  : "bg-blue-100 text-blue-800"
              }
            >
              {userRole === "admin" ? "管理者" : "編集者"}
            </Badge>
          )}

          {!form_id && (
            <Select value={selectedForm} onValueChange={handleFormSelection}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="フォームを選択" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">すべてのフォーム</SelectItem>
                {forms.map((form) => (
                  <SelectItem key={form.form_id} value={form.form_id}>
                    {form.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}

          {form_id && (
            <Button variant="secondary">
              <Users className="h-4 w-4 mr-2" />
              メンバー ({members.length})
            </Button>
          )}

          <Button
            onClick={() => loadData(form_id || selectedForm)}
            variant="secondary"
          >
            更新
          </Button>
        </div>
      </div>

      {!canEdit && userRole && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
          <div className="flex">
            <AlertCircle className="h-5 w-5 text-yellow-400" />
            <div className="ml-3">
              <p className="text-sm text-yellow-700">
                閲覧モードです。チケットの編集にはadmin権限が必要です。
              </p>
            </div>
          </div>
        </div>
      )}

      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <DragDropContext onDragEnd={handleDragEnd}>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {columns.map((column) => (
            <KanbanColumn
              key={column.id}
              column={column}
              tickets={ticketsByStatus[column.id]}
              members={members}
              canEdit={canEdit}
              onUpdateTicket={updateTicketStatus}
            />
          ))}
        </div>
      </DragDropContext>

      {tickets.length === 0 && !error && (
        <div className="text-center py-12">
          <ListChecks className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            チケットがありません
          </h3>
          <p className="text-muted-foreground">
            フォームからの回答があると自動的にチケットが作成されます。
          </p>
        </div>
      )}

      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>合計 {tickets.length} 件のチケット</span>
        <span>最終更新: {new Date().toLocaleString("ja-JP")}</span>
      </div>
    </div>
  );
}
