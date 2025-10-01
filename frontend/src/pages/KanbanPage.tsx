import React, { useState, useEffect, useMemo, useCallback } from "react";
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
import type {
  TicketSummary,
  TicketDetail,
  Member,
  FormSummary,
  TicketStatus,
  FormQuestion,
} from "../types";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "../components/ui/Dialog";
import { Label } from "../components/ui/Label";

interface KanbanColumn {
  id: TicketStatus;
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

const UNASSIGNED_VALUE = "__UNASSIGNED__";
const AUTO_TITLE_VALUE = "__AUTO_TITLE__";

interface TicketCardProps {
  ticket: TicketSummary;
  index: number;
  members: Member[];
  canEdit: boolean;
  onUpdateTicket: (ticketId: string, status: TicketStatus) => void;
  onOpenDetail: (ticketId: string) => void;
  onAssigneeChange: (ticketId: string, assigneeId: string | null) => void;
  onPriorityChange: (ticketId: string, priority: number) => void;
  isUpdating: boolean;
}

function TicketCard({
  ticket,
  index,
  members,
  canEdit,
  onUpdateTicket,
  onOpenDetail,
  onAssigneeChange,
  onPriorityChange,
  isUpdating,
}: TicketCardProps) {
  const baseAssignee = ticket.assignee ?? null;
  const assignedMember = baseAssignee
    ? members.find((m) => m.id === baseAssignee.id) || baseAssignee
    : null;
  const priorityColor =
    priorityColors[ticket.priority as keyof typeof priorityColors] ||
    priorityColors[5];

  const moveTicket = (direction: "left" | "right") => {
    return () => {
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
                <div className="text-xs text-gray-500 uppercase tracking-wide">
                  {ticket.form_title}
                </div>

                <button
                  type="button"
                  className="text-left text-sm font-semibold text-gray-800 line-clamp-2 hover:text-primary"
                  onClick={() => onOpenDetail(ticket.id)}
                >
                  {ticket.title || "（タイトル未設定）"}
                </button>

                <div className="flex items-center text-xs text-gray-500">
                  <Calendar className="h-3 w-3 mr-1" />
                  {new Date(ticket.submitted_at).toLocaleDateString("ja-JP")}
                </div>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center text-sm text-gray-600 flex-1">
                    <Hash className="h-3 w-3 mr-1" />
                    <span className="font-mono">{ticket.id.slice(0, 8)}</span>
                  </div>
                </div>

                <div className="text-sm text-gray-700">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-xs uppercase text-gray-500">
                      担当者
                    </span>
                    {canEdit ? (
                      <Select
                        disabled={isUpdating}
                        value={assignedMember?.id ?? UNASSIGNED_VALUE}
                        onValueChange={(value) =>
                          onAssigneeChange(
                            ticket.id,
                            value === UNASSIGNED_VALUE ? null : value
                          )
                        }
                      >
                        <SelectTrigger className="h-8 w-36">
                          <SelectValue placeholder="未割り当て" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value={UNASSIGNED_VALUE}>
                            未割り当て
                          </SelectItem>
                          {members.map((member) => (
                            <SelectItem key={member.id} value={member.id}>
                              {member.display_name || member.email}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <span>
                        {assignedMember?.display_name ||
                          assignedMember?.email ||
                          "未割り当て"}
                      </span>
                    )}
                  </div>

                  <div className="flex items-center justify-between gap-2 mt-3">
                    <span className="text-xs uppercase text-gray-500">
                      優先度
                    </span>
                    {canEdit ? (
                      <Select
                        disabled={isUpdating}
                        value={String(ticket.priority)}
                        onValueChange={(value) =>
                          onPriorityChange(ticket.id, Number(value))
                        }
                      >
                        <SelectTrigger className="h-8 w-24">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {[1, 2, 3, 4, 5].map((level) => (
                            <SelectItem key={level} value={String(level)}>
                              {level}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <span>{ticket.priority}</span>
                    )}
                  </div>
                </div>
              </div>
            </CardContent>
          </motion.div>
        </div>
      )}
    </Draggable>
  );
}

interface KanbanColumnProps {
  column: KanbanColumn;
  tickets: TicketSummary[];
  members: Member[];
  canEdit: boolean;
  onUpdateTicket: (ticketId: string, status: TicketStatus) => void;
  onOpenDetail: (ticketId: string) => void;
  onAssigneeChange: (ticketId: string, assigneeId: string | null) => void;
  onPriorityChange: (ticketId: string, priority: number) => void;
  updatingTickets: Record<string, boolean>;
}

function KanbanColumn({
  column,
  tickets,
  members,
  canEdit,
  onUpdateTicket,
  onOpenDetail,
  onAssigneeChange,
  onPriorityChange,
  updatingTickets,
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
                onOpenDetail={onOpenDetail}
                onAssigneeChange={onAssigneeChange}
                onPriorityChange={onPriorityChange}
                isUpdating={Boolean(updatingTickets[ticket.id])}
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
  const [tickets, setTickets] = useState<TicketSummary[]>([]);
  const [ticketDetails, setTicketDetails] = useState<
    Record<string, TicketDetail>
  >({});
  const [selectedTicketId, setSelectedTicketId] = useState<string | null>(null);
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [isDetailLoading, setIsDetailLoading] = useState(false);
  const [members, setMembers] = useState<Member[]>([]);
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [selectedForm, setSelectedForm] = useState<string>("all");
  const [userRole, setUserRole] = useState<"admin" | "editor" | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [formQuestionsCache, setFormQuestionsCache] = useState<
    Record<string, FormQuestion[]>
  >({});
  const [formQuestions, setFormQuestions] = useState<FormQuestion[]>([]);
  const [isQuestionsLoading, setIsQuestionsLoading] = useState(false);
  const [updatingTickets, setUpdatingTickets] = useState<
    Record<string, boolean>
  >({});
  const [isUpdatingTitleQuestion, setIsUpdatingTitleQuestion] = useState(false);
  const [titleQuestionSelection, setTitleQuestionSelection] = useState<
    string | null
  >(null);

  const canEdit = userRole === "admin";

  const ensureFormQuestions = useCallback(
    async (formId: string) => {
      if (formQuestionsCache[formId]) {
        setFormQuestions(formQuestionsCache[formId]);
        return;
      }
      setIsQuestionsLoading(true);
      try {
        const response = await apiClient.getFormQuestions(formId);
        setFormQuestions(response.questions);
        setFormQuestionsCache((prev) => ({
          ...prev,
          [formId]: response.questions,
        }));
      } catch (err) {
        console.error("Failed to load form questions:", err);
        setFormQuestions([]);
      } finally {
        setIsQuestionsLoading(false);
      }
    },
    [formQuestionsCache]
  );

  const loadData = useCallback(
    async (formId?: string) => {
      try {
        setIsLoading(true);
        setError(null);

        const [formsResponse, ticketsResponse] = await Promise.all([
          apiClient.getForms(),
          apiClient.getTickets(formId === "all" ? undefined : formId),
        ]);

        setForms(formsResponse.forms);
        setTickets(ticketsResponse.tickets);
        setTicketDetails({});

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

          await ensureFormQuestions(formId);
        } else {
          setMembers([]);
          setUserRole(null);
          setFormQuestions([]);
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
    },
    [ensureFormQuestions, user?.id]
  );

  const handleFormSelection = (value: string) => {
    if (value === "all") {
      navigate("/kanban");
    } else {
      navigate(`/kanban/${value}`);
    }
  };

  const applyTicketDetail = useCallback((detail: TicketDetail) => {
    const summary: TicketSummary = {
      id: detail.id,
      form_id: detail.form_id,
      form_title: detail.form_title,
      response_id: detail.response_id,
      status: detail.status,
      priority: detail.priority,
      title_question_id: detail.title_question_id,
      title: detail.title,
      assignee: detail.assignee ?? null,
      submitted_at: detail.submitted_at,
      updated_at: detail.updated_at,
    };
    setTickets((prev) => {
      let found = false;
      const next = prev.map((ticket) => {
        if (ticket.id === summary.id) {
          found = true;
          return summary;
        }
        return ticket;
      });
      if (!found) {
        next.push(summary);
      }
      return next;
    });
    setTicketDetails((prev) => ({
      ...prev,
      [detail.id]: detail,
    }));
  }, []);

  const setTicketUpdating = useCallback((ticketId: string, value: boolean) => {
    setUpdatingTickets((prev) => {
      if (value) {
        return { ...prev, [ticketId]: true };
      }
      const copy = { ...prev };
      delete copy[ticketId];
      return copy;
    });
  }, []);

  useEffect(() => {
    if (form_id) {
      loadData(form_id);
      setSelectedForm(form_id);
    } else {
      loadData(selectedForm);
    }
  }, [form_id, selectedForm, loadData]);

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

    const newStatus = destination.droppableId as TicketStatus;
    if (ticket.status === newStatus) {
      return;
    }

    let snapshot: TicketSummary | null = null;
    setTickets((prev) =>
      prev.map((t) => {
        if (t.id === draggableId) {
          snapshot = t;
          return {
            ...t,
            status: newStatus,
            updated_at: new Date().toISOString(),
          };
        }
        return t;
      })
    );

    if (!snapshot) {
      return;
    }

    setTicketUpdating(draggableId, true);

    try {
      const detail = await apiClient.updateTicket(draggableId, {
        status: newStatus,
      });
      applyTicketDetail(detail);
    } catch (err) {
      setTickets((prev) =>
        prev.map((t) => (t.id === draggableId ? snapshot! : t))
      );
      setError("チケットの更新に失敗しました");
      console.error("Failed to update ticket:", err);
    } finally {
      setTicketUpdating(draggableId, false);
    }
  };

  const updateTicketStatus = useCallback(
    async (ticketId: string, newStatus: TicketStatus) => {
      const ticket = tickets.find((t) => t.id === ticketId);
      if (!ticket || ticket.status === newStatus) return;

      const snapshot: TicketSummary = ticket;
      setTickets((prev) =>
        prev.map((t) =>
          t.id === ticketId
            ? {
                ...t,
                status: newStatus,
                updated_at: new Date().toISOString(),
              }
            : t
        )
      );

      setTicketUpdating(ticketId, true);
      try {
        const detail = await apiClient.updateTicket(ticketId, {
          status: newStatus,
        });
        applyTicketDetail(detail);
      } catch (err) {
        setTickets((prev) =>
          prev.map((t) => (t.id === ticketId ? snapshot : t))
        );
        setError("チケットの更新に失敗しました");
        console.error("Failed to update ticket:", err);
      } finally {
        setTicketUpdating(ticketId, false);
      }
    },
    [tickets, applyTicketDetail, setTicketUpdating]
  );

  const handleAssigneeChange = useCallback(
    async (ticketId: string, assigneeId: string | null) => {
      const ticket = tickets.find((t) => t.id === ticketId);
      if (!ticket) return;

      const snapshot = ticket;
      setTickets((prev) =>
        prev.map((t) =>
          t.id === ticketId
            ? {
                ...t,
                assignee: assigneeId
                  ? (() => {
                      const member = members.find((m) => m.id === assigneeId);
                      if (member) {
                        return {
                          id: member.id,
                          display_name: member.display_name || member.email,
                          email: member.email,
                        };
                      }
                      return t.assignee ?? null;
                    })()
                  : null,
              }
            : t
        )
      );

      setTicketUpdating(ticketId, true);
      try {
        const detail = await apiClient.updateTicket(ticketId, {
          assignee_id: assigneeId,
        });
        applyTicketDetail(detail);
      } catch (err) {
        setTickets((prev) =>
          prev.map((t) => (t.id === ticketId ? snapshot : t))
        );
        setError("担当者の更新に失敗しました");
        console.error("Failed to update assignee:", err);
      } finally {
        setTicketUpdating(ticketId, false);
      }
    },
    [tickets, members, applyTicketDetail, setTicketUpdating]
  );

  const handlePriorityChange = useCallback(
    async (ticketId: string, priority: number) => {
      const ticket = tickets.find((t) => t.id === ticketId);
      if (!ticket || ticket.priority === priority) return;

      const snapshot = ticket;
      setTickets((prev) =>
        prev.map((t) => (t.id === ticketId ? { ...t, priority } : t))
      );

      setTicketUpdating(ticketId, true);
      try {
        const detail = await apiClient.updateTicket(ticketId, { priority });
        applyTicketDetail(detail);
      } catch (err) {
        setTickets((prev) =>
          prev.map((t) => (t.id === ticketId ? snapshot : t))
        );
        setError("優先度の更新に失敗しました");
        console.error("Failed to update priority:", err);
      } finally {
        setTicketUpdating(ticketId, false);
      }
    },
    [tickets, applyTicketDetail, setTicketUpdating]
  );

  const handleOpenDetail = useCallback(
    async (ticketId: string) => {
      setSelectedTicketId(ticketId);
      setIsDetailOpen(true);
      const summary = tickets.find((t) => t.id === ticketId);
      if (summary) {
        if (formQuestionsCache[summary.form_id]) {
          setFormQuestions(formQuestionsCache[summary.form_id]);
        } else {
          await ensureFormQuestions(summary.form_id);
        }
      }

      if (ticketDetails[ticketId]) {
        return;
      }

      setIsDetailLoading(true);
      try {
        const detail = await apiClient.getTicket(ticketId);
        applyTicketDetail(detail);
      } catch (err) {
        setError("チケット詳細の取得に失敗しました");
        console.error("Failed to load ticket detail:", err);
      } finally {
        setIsDetailLoading(false);
      }
    },
    [
      tickets,
      ticketDetails,
      formQuestionsCache,
      ensureFormQuestions,
      applyTicketDetail,
    ]
  );

  const closeDetail = () => {
    setIsDetailOpen(false);
    setSelectedTicketId(null);
  };

  const selectedDetail =
    selectedTicketId && ticketDetails[selectedTicketId]
      ? ticketDetails[selectedTicketId]
      : null;

  const handleTitleQuestionUpdate = useCallback(
    async (questionId: string | null) => {
      if (!form_id) return;
      const previous = titleQuestionSelection ?? null;
      if (previous === (questionId ?? null)) {
        return;
      }

      setTitleQuestionSelection(questionId ?? null);
      setIsUpdatingTitleQuestion(true);
      try {
        await apiClient.updateFormTitleQuestion(form_id, questionId);
        const refreshed = await apiClient.getTickets(form_id);
        setTickets(refreshed.tickets);
        setTicketDetails({});
        await ensureFormQuestions(form_id);
        if (selectedTicketId) {
          try {
            const detail = await apiClient.getTicket(selectedTicketId);
            applyTicketDetail(detail);
          } catch (detailErr) {
            console.error("Failed to refresh ticket detail:", detailErr);
          }
        }
      } catch (err) {
        setTitleQuestionSelection(previous);
        setError("タイトル用の質問更新に失敗しました");
        console.error("Failed to update title question:", err);
      } finally {
        setIsUpdatingTitleQuestion(false);
      }
    },
    [
      form_id,
      titleQuestionSelection,
      ensureFormQuestions,
      selectedTicketId,
      applyTicketDetail,
    ]
  );

  const ticketsByStatus = useMemo(() => {
    return {
      new: tickets.filter((t) => t.status === "new"),
      in_progress: tickets.filter((t) => t.status === "in_progress"),
      done: tickets.filter((t) => t.status === "done"),
    };
  }, [tickets]);

  useEffect(() => {
    if (form_id) {
      const current = tickets.find((t) => t.form_id === form_id);
      setTitleQuestionSelection(current?.title_question_id ?? null);
    } else {
      setTitleQuestionSelection(null);
    }
  }, [form_id, tickets]);

  const questionOptions = useMemo<FormQuestion[]>(() => {
    if (formQuestions.length > 0) {
      return formQuestions;
    }
    if (selectedDetail) {
      return selectedDetail.answers.map((answer) => ({
        form_id: selectedDetail.form_id,
        question_id: answer.question_id,
        title: answer.question_title,
        question_type: answer.question_type,
      }));
    }
    return [];
  }, [formQuestions, selectedDetail]);

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
            <Button
              variant="secondary"
              onClick={() => navigate(`/members/${form_id}`)}
            >
              <Users className="h-4 w-4 mr-2" />
              メンバー ({members.length})
            </Button>
          )}

          <Button
            onClick={() => void loadData(form_id || selectedForm)}
            variant="secondary"
          >
            更新
          </Button>
        </div>
      </div>

      {form_id && canEdit && (
        <div className="flex items-center gap-3 text-sm">
          <Label className="text-xs uppercase text-gray-500">
            カードタイトル質問
          </Label>
          <Select
            disabled={
              isUpdatingTitleQuestion ||
              isQuestionsLoading ||
              questionOptions.length === 0
            }
            value={titleQuestionSelection ?? AUTO_TITLE_VALUE}
            onValueChange={(value) =>
              handleTitleQuestionUpdate(
                value === AUTO_TITLE_VALUE ? null : value
              )
            }
          >
            <SelectTrigger className="w-64">
              <SelectValue placeholder="質問を選択" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={AUTO_TITLE_VALUE}>自動選択</SelectItem>
              {questionOptions.map((question) => (
                <SelectItem
                  key={question.question_id}
                  value={question.question_id}
                >
                  {question.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {isQuestionsLoading && (
            <span className="text-xs text-muted-foreground">
              質問を読み込み中...
            </span>
          )}
        </div>
      )}

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
              onOpenDetail={handleOpenDetail}
              onAssigneeChange={handleAssigneeChange}
              onPriorityChange={handlePriorityChange}
              updatingTickets={updatingTickets}
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

      <Dialog
        open={isDetailOpen}
        onOpenChange={(open) => (!open ? closeDetail() : null)}
      >
        <DialogContent className="sm:max-w-2xl bg-white">
          {isDetailLoading || !selectedDetail ? (
            <div className="py-12 text-center text-muted-foreground">
              読み込み中...
            </div>
          ) : (
            <div className="space-y-6">
              <DialogHeader>
                <DialogTitle>
                  {selectedDetail.title || "チケット詳細"}
                </DialogTitle>
                <DialogDescription>
                  {selectedDetail.form_title} / 回答ID:{" "}
                  {selectedDetail.response_id}
                </DialogDescription>
              </DialogHeader>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-3">
                  <div>
                    <Label className="text-xs text-gray-500">担当者</Label>
                    <Select
                      disabled={
                        !canEdit || Boolean(updatingTickets[selectedDetail.id])
                      }
                      value={selectedDetail.assignee?.id ?? UNASSIGNED_VALUE}
                      onValueChange={(value) =>
                        handleAssigneeChange(
                          selectedDetail.id,
                          value === UNASSIGNED_VALUE ? null : value
                        )
                      }
                    >
                      <SelectTrigger className="mt-1">
                        <SelectValue placeholder="未割り当て" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value={UNASSIGNED_VALUE}>
                          未割り当て
                        </SelectItem>
                        {members.map((member) => (
                          <SelectItem key={member.id} value={member.id}>
                            {member.display_name || member.email}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div>
                    <Label className="text-xs text-gray-500">優先度</Label>
                    <Select
                      disabled={
                        !canEdit || Boolean(updatingTickets[selectedDetail.id])
                      }
                      value={String(selectedDetail.priority)}
                      onValueChange={(value) =>
                        handlePriorityChange(selectedDetail.id, Number(value))
                      }
                    >
                      <SelectTrigger className="mt-1">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {[1, 2, 3, 4, 5].map((level) => (
                          <SelectItem key={level} value={String(level)}>
                            {level}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="space-y-2 text-sm text-gray-600">
                  <div>
                    <span className="font-medium text-gray-700">
                      ステータス:
                    </span>{" "}
                    {selectedDetail.status}
                  </div>
                  <div>
                    <span className="font-medium text-gray-700">回答日:</span>{" "}
                    {new Date(selectedDetail.submitted_at).toLocaleString(
                      "ja-JP"
                    )}
                  </div>
                  <div>
                    <span className="font-medium text-gray-700">最終更新:</span>{" "}
                    {new Date(selectedDetail.updated_at).toLocaleString(
                      "ja-JP"
                    )}
                  </div>
                </div>
              </div>

              <div className="space-y-4 max-h-80 overflow-y-auto pr-2">
                {selectedDetail.answers.map((answer) => (
                  <div
                    key={answer.question_id}
                    className="rounded-md border border-gray-200 p-3"
                  >
                    <div className="text-xs uppercase text-gray-500">
                      {answer.question_type}
                    </div>
                    <div className="text-sm font-medium text-gray-800">
                      {answer.question_title}
                    </div>
                    <div className="mt-2 text-sm text-gray-700 whitespace-pre-line">
                      {answer.display_value || "未回答"}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
