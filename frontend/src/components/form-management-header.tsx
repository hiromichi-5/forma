import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { LayoutList, LayoutGrid, Search, Users, ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router-dom";
import type { FormStatus, FormQuestion } from "@/types";

const hexToRgba = (hex: string | null | undefined, alpha: number): string => {
  if (!hex) return "transparent";
  const sanitized = hex.replace("#", "");
  const red = Number.parseInt(sanitized.slice(0, 2), 16);
  const green = Number.parseInt(sanitized.slice(2, 4), 16);
  const blue = Number.parseInt(sanitized.slice(4, 6), 16);
  return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
};

type FormManagementHeaderProps = {
  formTitle: string;
  viewMode: "list" | "kanban";
  onViewModeChange: (mode: "list" | "kanban") => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  statusFilter: "all" | string;
  onStatusFilterChange: (status: "all" | string) => void;
  statuses: FormStatus[];
  questions: FormQuestion[];
  titleQuestionId: string | null;
  onTitleQuestionChange: (questionId: string | null) => void;
  onMembersClick: () => void;
};

export function FormManagementHeader({
  formTitle,
  viewMode,
  onViewModeChange,
  searchQuery,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  statuses,
  questions,
  titleQuestionId,
  onTitleQuestionChange,
  onMembersClick,
}: FormManagementHeaderProps) {
  const navigate = useNavigate();
  const sortedStatuses = [...statuses].sort(
    (a, b) => a.display_order - b.display_order
  );
  const selectedStatus =
    statusFilter === "all"
      ? null
      : sortedStatuses.find((status) => status.id === statusFilter);

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => navigate("/")}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-foreground">{formTitle}</h1>
        </div>
        <Button
          onClick={onMembersClick}
          variant="outline"
          className="gap-2 bg-transparent"
        >
          <Users className="h-4 w-4" />
          メンバー管理
        </Button>
      </div>

      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="回答者のメールアドレスで検索..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-9"
          />
        </div>

        <Select value={statusFilter} onValueChange={onStatusFilterChange}>
          <SelectTrigger className="w-full sm:w-[180px] border-0 shadow-none">
            <div
              className="flex items-center gap-2 px-2 py-1 rounded"
              style={{
                backgroundColor: hexToRgba(selectedStatus?.color ?? null, 0.1),
              }}
            >
              {selectedStatus ? (
                <div
                  className="w-2 h-2 rounded-full shrink-0"
                  style={{
                    backgroundColor: selectedStatus.color ?? "#9CA3AF",
                  }}
                />
              ) : (
                <div className="w-2 h-2 rounded-full shrink-0 bg-muted-foreground/60" />
              )}
              <span className="text-sm">
                {selectedStatus?.name ?? "全てのステータス"}
              </span>
            </div>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全てのステータス</SelectItem>
            {sortedStatuses.map((status) => (
              <SelectItem key={status.id} value={status.id}>
                {status.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <div className="relative grid grid-cols-2 gap-0.5 bg-muted p-0.5 rounded-md">
          <div
            className="absolute top-0.5 bottom-0.5 bg-background rounded shadow-sm transition-transform duration-200 ease-in-out"
            style={{
              width: "calc(50% - 0.125rem)",
              transform:
                viewMode === "kanban"
                  ? "translateX(calc(100% + 0.125rem))"
                  : "translateX(0)",
            }}
          />

          <button
            type="button"
            onClick={() => onViewModeChange("list")}
            className={`relative z-10 flex items-center justify-center gap-1 px-3 py-1.5 rounded text-sm font-medium transition-colors whitespace-nowrap ${
              viewMode === "list"
                ? "text-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            <LayoutList className="h-4 w-4" />
            リスト
          </button>

          <button
            type="button"
            onClick={() => onViewModeChange("kanban")}
            className={`relative z-10 flex items-center justify-center gap-1 px-3 py-1.5 rounded text-sm font-medium transition-colors whitespace-nowrap ${
              viewMode === "kanban"
                ? "text-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            <LayoutGrid className="h-4 w-4" />
            カンバン
          </button>
        </div>
      </div>

      <div className="flex justify-end">
        <div className="space-y-2">
          <p className="text-xs text-muted-foreground text-right">
            リストやカンバンに表示する質問を選択できます
          </p>
          <Select
            value={titleQuestionId ?? "none"}
            onValueChange={(value) =>
              onTitleQuestionChange(value === "none" ? null : value)
            }
          >
            <SelectTrigger className="w-full sm:w-[300px]">
              <SelectValue placeholder="タイトル質問を選択" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">タイトル質問なし</SelectItem>
              {questions.map((question) => (
                <SelectItem
                  key={question.question_id}
                  value={question.question_id}
                >
                  {question.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </div>
  );
}
