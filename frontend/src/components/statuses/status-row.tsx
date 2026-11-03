import { useState } from "react"
import { ArrowDown, ArrowUp, Check, Flag, Pencil, Trash2, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { fallbackStatusColor } from "@/lib/ticket-display"
import { cn } from "@/lib/utils"
import type { FormStatus } from "@/types"
import { ColorInput } from "./color-input"
import { colorValueOrNull, createColorState } from "./color-state"

const DEFAULT_STATUS_DESCRIPTION = "新しい回答はこのステータスになります"

type StatusRowProps = {
  status: FormStatus
  isFirst: boolean
  isLast: boolean
  /** 他の操作の実行中は、順序が崩れないよう操作を受け付けない。 */
  disabled: boolean
  onEdit: () => void
  onSetDefault: () => void
  onMove: (direction: "up" | "down") => void
  onDelete: () => void
}

/** ステータス一覧の 1 行（表示中）。 */
export function StatusRow({
  status,
  isFirst,
  isLast,
  disabled,
  onEdit,
  onSetDefault,
  onMove,
  onDelete,
}: StatusRowProps) {
  return (
    <div className="flex items-center gap-2 p-2">
      <button
        type="button"
        className="flex min-w-0 flex-1 cursor-pointer items-center gap-2 text-left select-none"
        onClick={onEdit}
      >
        <div
          className="h-3 w-3 flex-shrink-0 rounded-full"
          style={{ backgroundColor: status.color ?? fallbackStatusColor }}
        />
        <span className="min-w-0 flex-1 truncate text-sm">{status.name}</span>
      </button>
      <Tooltip>
        <TooltipTrigger asChild>
          <span className="inline-flex">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 flex-shrink-0"
              aria-label={status.is_default ? "デフォルトのステータス" : "デフォルトに設定"}
              disabled={disabled}
              onClick={() => (status.is_default ? undefined : onSetDefault())}
            >
              <Flag
                className={cn(
                  "h-3.5 w-3.5",
                  status.is_default ? "fill-primary text-primary" : "text-muted-foreground",
                )}
              />
            </Button>
          </span>
        </TooltipTrigger>
        <TooltipContent>
          {status.is_default
            ? DEFAULT_STATUS_DESCRIPTION
            : `クリックすると${DEFAULT_STATUS_DESCRIPTION}`}
        </TooltipContent>
      </Tooltip>
      <div className="flex flex-shrink-0 items-center gap-0.5">
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          aria-label="上へ移動"
          disabled={disabled || isFirst}
          onClick={() => onMove("up")}
        >
          <ArrowUp className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          aria-label="下へ移動"
          disabled={disabled || isLast}
          onClick={() => onMove("down")}
        >
          <ArrowDown className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          aria-label="編集"
          disabled={disabled}
          onClick={onEdit}
        >
          <Pencil className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          aria-label="削除"
          disabled={disabled || status.is_default}
          onClick={onDelete}
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  )
}

type StatusEditRowProps = {
  status: FormStatus
  disabled: boolean
  onSave: (values: { name: string; color: string | null }) => void
  onCancel: () => void
}

/** ステータス一覧の 1 行（編集中）。編集中の値はこの行の中だけで持つ。 */
export function StatusEditRow({ status, disabled, onSave, onCancel }: StatusEditRowProps) {
  const [name, setName] = useState(status.name)
  const [color, setColor] = useState(() => createColorState(status.color))

  return (
    <div className="space-y-2 p-2">
      <div className="flex items-center gap-2">
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="h-8 flex-1"
          placeholder="ステータス名"
          aria-label="ステータス名"
        />
        <ColorInput value={color} onChange={setColor} swatchClassName="h-6 w-6" />
      </div>
      <div className="flex gap-1.5">
        <Button
          size="sm"
          onClick={() => onSave({ name: name.trim(), color: colorValueOrNull(color) })}
          disabled={disabled || !name.trim()}
          className="h-7 gap-1"
        >
          <Check className="h-3.5 w-3.5" />
          保存
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={onCancel}
          disabled={disabled}
          className="h-7 gap-1"
        >
          <X className="h-3.5 w-3.5" />
          キャンセル
        </Button>
      </div>
    </div>
  )
}
