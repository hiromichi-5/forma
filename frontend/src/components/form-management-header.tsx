import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { LayoutList, LayoutGrid, Search, Users, Settings, Trash2, ChevronDown, ExternalLink, RefreshCw, Bell } from "lucide-react";
import { fallbackStatusColor, sortStatuses } from "@/lib/ticket-display";
import type { FormStatus, FormQuestion } from "@/types";

type FormManagementHeaderProps = {
  googleFormId: string | null;
  formTitle: string;
  viewMode: "list" | "kanban";
  onViewModeChange: (mode: "list" | "kanban") => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  statusFilters: string[];
  onStatusFiltersChange: (statusIds: string[]) => void;
  statuses: FormStatus[];
  questions: FormQuestion[];
  titleQuestionId: string | null;
  onTitleQuestionChange: (questionId: string | null) => void;
  onStatusManageClick?: () => void;
  onMembersClick: () => void;
  onNotificationsClick: () => void;
  onUnregisterClick: () => void;
  isSyncing: boolean;
  onSyncClick: () => void;
};

export function FormManagementHeader({
  googleFormId,
  formTitle,
  viewMode,
  onViewModeChange,
  searchQuery,
  onSearchChange,
  statusFilters,
  onStatusFiltersChange,
  statuses,
  questions,
  titleQuestionId,
  onTitleQuestionChange,
  onStatusManageClick,
  onMembersClick,
  onNotificationsClick,
  onUnregisterClick,
  isSyncing,
  onSyncClick,
}: FormManagementHeaderProps) {
  const sortedStatuses = sortStatuses(statuses);
  const selectedStatuses = sortedStatuses.filter((status) =>
    statusFilters.includes(status.id)
  );
  // 1件だけ選んでいる場合は、そのステータスの色と名前をそのまま見せる。
  const singleSelectedStatus =
    selectedStatuses.length === 1 ? selectedStatuses[0] : null;
  const statusFilterLabel =
    selectedStatuses.length === 0
      ? "全てのステータス"
      : singleSelectedStatus
        ? singleSelectedStatus.name
        : `${selectedStatuses.length}件のステータス`;

  const toggleStatusFilter = (statusId: string, checked: boolean) => {
    onStatusFiltersChange(
      checked
        ? [...statusFilters, statusId]
        : statusFilters.filter((id) => id !== statusId)
    );
  };

  return (
    <div className="space-y-4">
      {/* 行1: タイトルと管理系ボタン */}
      <div className="flex items-center gap-3">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              className="flex items-center gap-1.5 rounded-md px-2 py-1 -mx-2 text-left transition-colors hover:bg-accent"
            >
              <h1 className="text-2xl font-bold text-foreground">{formTitle}</h1>
              <ChevronDown className="h-5 w-5 text-muted-foreground shrink-0" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="start"
            className="p-1.5 border-gray-300 shadow-sm rounded-xl"
          >
            {googleFormId && (
              <>
                <DropdownMenuItem asChild className="text-base py-2 rounded-lg">
                  <a
                    href={`https://docs.google.com/forms/d/${googleFormId}/edit`}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <ExternalLink className="h-4 w-4" />
                    Googleフォームを編集
                  </a>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
              </>
            )}
            {onStatusManageClick && (
              <DropdownMenuItem
                onClick={onStatusManageClick}
                className="text-base py-2 rounded-lg"
              >
                <Settings className="h-4 w-4" />
                ステータス管理
              </DropdownMenuItem>
            )}
            <DropdownMenuItem onClick={onMembersClick} className="text-base py-2 rounded-lg">
              <Users className="h-4 w-4" />
              メンバー管理
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={onNotificationsClick}
              className="text-base py-2 rounded-lg"
            >
              <Bell className="h-4 w-4" />
              通知設定
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={onUnregisterClick}
              className="text-base py-2 rounded-lg text-destructive focus:text-destructive"
            >
              <Trash2 className="h-4 w-4" />
              登録解除
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <Button
          type="button"
          variant="default"
          size="sm"
          onClick={onSyncClick}
          disabled={isSyncing}
          className="gap-1.5"
        >
          <RefreshCw className={`h-4 w-4 ${isSyncing ? "animate-spin" : ""}`} />
          {isSyncing ? "同期中..." : "Googleフォームと同期する"}
        </Button>
      </div>

      {/* 行2: 検索・フィルター・表示設定 */}
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

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              className="border-input hover:bg-accent/50 focus-visible:border-ring focus-visible:ring-ring/50 flex h-9 w-full items-center justify-between gap-2 rounded-xl border bg-transparent px-3 py-2 text-sm whitespace-nowrap transition-[color,box-shadow] outline-none focus-visible:ring-[3px] sm:w-[200px]"
            >
              <div className="flex min-w-0 items-center gap-2">
                {singleSelectedStatus && (
                  <div
                    className="w-2 h-2 rounded-full shrink-0"
                    style={{
                      backgroundColor:
                        singleSelectedStatus.color ?? fallbackStatusColor,
                    }}
                  />
                )}
                <span className="truncate">{statusFilterLabel}</span>
              </div>
              <ChevronDown className="size-4 shrink-0 opacity-50" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-[220px]">
            <DropdownMenuCheckboxItem
              checked={statusFilters.length === 0}
              onCheckedChange={() => onStatusFiltersChange([])}
            >
              全てのステータス
            </DropdownMenuCheckboxItem>
            <DropdownMenuSeparator />
            {sortedStatuses.map((status) => (
              <DropdownMenuCheckboxItem
                key={status.id}
                checked={statusFilters.includes(status.id)}
                onCheckedChange={(checked) =>
                  toggleStatusFilter(status.id, checked)
                }
                // 続けて複数選べるよう、選択してもメニューを閉じない。
                onSelect={(event) => event.preventDefault()}
              >
                <div className="flex min-w-0 items-center gap-2">
                  <div
                    className="w-2 h-2 rounded-full shrink-0"
                    style={{
                      backgroundColor: status.color ?? fallbackStatusColor,
                    }}
                  />
                  <span className="truncate">{status.name}</span>
                </div>
              </DropdownMenuCheckboxItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        <Select
          value={titleQuestionId ?? "none"}
          onValueChange={(value) =>
            onTitleQuestionChange(value === "none" ? null : value)
          }
        >
          <SelectTrigger className="w-full sm:w-[200px]">
            <SelectValue placeholder="表示する質問を選択" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">タイトル質問なし</SelectItem>
            {questions.map((question) => (
              <SelectItem key={question.question_id} value={question.question_id}>
                {question.title}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <div className="relative grid grid-cols-2 gap-0.5 bg-muted p-0.5 rounded-md">
          <div
            className="absolute top-0.5 bottom-0.5 bg-white rounded shadow-sm transition-transform duration-200 ease-in-out"
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
    </div>
  );
}
