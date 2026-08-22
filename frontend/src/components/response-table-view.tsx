"use client";

import { Fragment, useMemo, useState } from "react";
import type { User } from "@/types/form-response";
import type {
  FormStatus,
  TicketDetail,
  TicketPriority,
  TicketSummary,
} from "@/types";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ChevronDown, ChevronRight, MessageSquare } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { ja } from "date-fns/locale";
import {
  fallbackStatusColor,
  hexToRgba,
  priorityConfig,
  respondentEmailLabel,
  sortStatuses,
  statusById,
  toPriorityValue,
} from "@/lib/ticket-display";

type ResponseTableViewProps = {
  responses: TicketSummary[];
  details: Record<string, TicketDetail>;
  users: User[];
  statuses: FormStatus[];
  onExpandRow: (id: string) => void;
  onStatusChange: (id: string, statusId: string) => void;
  onAssignChange: (id: string, userId: string | null) => void;
  onPriorityChange: (id: string, priority: TicketPriority) => void;
  onOpenDetail: (response: TicketSummary) => void;
};

export function ResponseTableView({
  responses,
  details,
  users,
  statuses,
  onExpandRow,
  onStatusChange,
  onAssignChange,
  onPriorityChange,
  onOpenDetail,
}: ResponseTableViewProps) {
  const [expandedRow, setExpandedRow] = useState<string | null>(null);

  const sortedStatuses = sortStatuses(statuses);
  const statusMap = useMemo(() => statusById(statuses), [statuses]);

  const toggleRow = (id: string) => {
    if (expandedRow === id) {
      setExpandedRow(null);
      return;
    }
    setExpandedRow(id);
    onExpandRow(id);
  };

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
          {responses.map((response) => {
            const isExpanded = expandedRow === response.id;
            const detailId = `response-details-${response.id}`;
            const email = respondentEmailLabel(response.respondent_email);
            const status = statusMap.get(response.status.id) ?? response.status;

            return (
              <Fragment key={response.id}>
                <TableRow className="hover:bg-muted/50 transition-colors">
                  <TableCell>
                    <button
                      type="button"
                      className="flex w-full items-start justify-between gap-3 rounded-sm text-left outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                      onClick={() => toggleRow(response.id)}
                      aria-expanded={isExpanded}
                      aria-controls={detailId}
                    >
                      <div>
                        <p className="font-medium text-foreground">{email}</p>
                        <p className="mt-1 text-sm text-muted-foreground">
                          {response.title}
                        </p>
                      </div>
                      {isExpanded ? (
                        <ChevronDown
                          className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground"
                          aria-hidden="true"
                        />
                      ) : (
                        <ChevronRight
                          className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground"
                          aria-hidden="true"
                        />
                      )}
                    </button>
                  </TableCell>
                  <TableCell>
                    <Select
                      value={response.status.id}
                      onValueChange={(value) =>
                        onStatusChange(response.id, value)
                      }
                    >
                      <SelectTrigger className="w-[130px] h-8 border-0 shadow-none">
                        <div
                          className="flex items-center gap-2 px-2 py-1 rounded"
                          style={{
                            backgroundColor: hexToRgba(status.color, 0.1),
                          }}
                        >
                          <div
                            className="w-2 h-2 rounded-full shrink-0"
                            style={{
                              backgroundColor: status.color || fallbackStatusColor,
                            }}
                          />
                          <span className="text-sm">{status.name}</span>
                        </div>
                      </SelectTrigger>
                      <SelectContent>
                        {sortedStatuses.map((status) => (
                          <SelectItem key={status.id} value={status.id}>
                            {status.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </TableCell>
                  <TableCell>
                    <Select
                      value={response.assignee?.id ?? "unassigned"}
                      onValueChange={(value) =>
                        onAssignChange(
                          response.id,
                          value === "unassigned" ? null : value
                        )
                      }
                    >
                      <SelectTrigger className="w-[130px] h-8 border-0 shadow-none">
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
                  <TableCell>
                    <Select
                      value={response.priority}
                      onValueChange={(value) =>
                        onPriorityChange(response.id, toPriorityValue(value))
                      }
                    >
                      <SelectTrigger className="w-[90px] h-8 border-0 shadow-none">
                        <div
                          className="flex items-center gap-2 px-2 py-1 rounded"
                          style={{
                            backgroundColor: hexToRgba(
                              priorityConfig[response.priority].hex,
                              0.1
                            ),
                          }}
                        >
                          <span
                            className="text-sm font-medium"
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
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-muted-foreground">
                      {formatDistanceToNow(new Date(response.submitted_at), {
                        addSuffix: true,
                        locale: ja,
                      })}
                    </span>
                  </TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => onOpenDetail(response)}
                      className="gap-2 h-8"
                      aria-label={`回答詳細を開く: ${email}`}
                    >
                      <MessageSquare className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
                {isExpanded && (
                  <TableRow className="hover:bg-muted/20">
                    <TableCell colSpan={6} className="bg-muted/20 p-4">
                      <div id={detailId} className="space-y-3">
                        <h4 className="font-semibold text-sm text-foreground">
                          回答内容
                        </h4>
                        {details[response.id] ? (
                          details[response.id].answers.map((answer, index) => (
                            <div
                              key={`${answer.question_id}-${index}`}
                              className="text-sm"
                            >
                              <span className="text-muted-foreground">
                                {answer.question_title}{" "}
                              </span>
                              <span className="text-foreground">
                                {answer.display_value}
                              </span>
                            </div>
                          ))
                        ) : (
                          <p className="text-sm text-muted-foreground">
                            読み込み中...
                          </p>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                )}
              </Fragment>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
