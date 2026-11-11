import { useState } from "react"
import { Check } from "lucide-react"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Switch } from "@/components/ui/switch"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import type { ColorInputState } from "./color-state"

const STATUS_COLOR_PRESETS = [
  { name: "レッド", value: "#EF4444" },
  { name: "オレンジ", value: "#F97316" },
  { name: "アンバー", value: "#F59E0B" },
  { name: "イエロー", value: "#EAB308" },
  { name: "ライム", value: "#84CC16" },
  { name: "グリーン", value: "#22C55E" },
  { name: "エメラルド", value: "#10B981" },
  { name: "ティール", value: "#14B8A6" },
  { name: "シアン", value: "#06B6D4" },
  { name: "スカイ", value: "#0EA5E9" },
  { name: "ブルー", value: "#3B82F6" },
  { name: "インディゴ", value: "#6366F1" },
  { name: "バイオレット", value: "#8B5CF6" },
  { name: "パープル", value: "#A855F7" },
  { name: "ピンク", value: "#EC4899" },
  { name: "グレー", value: "#6B7280" },
]

type ColorInputProps = {
  value: ColorInputState
  onChange: (value: ColorInputState) => void
  /** 色見本の大きさ。 */
  swatchClassName?: string
}

/** 色を使うかどうかのスイッチと、プリセットから色を選ぶボタン。 */
export function ColorInput({ value, onChange, swatchClassName = "h-7 w-7" }: ColorInputProps) {
  const [open, setOpen] = useState(false)

  return (
    <div className="flex items-center gap-1">
      <Tooltip>
        <TooltipTrigger asChild>
          <span className="inline-flex">
            <Switch
              aria-label="色を設定する"
              checked={value.enabled}
              onCheckedChange={(enabled) => onChange({ ...value, enabled })}
            />
          </span>
        </TooltipTrigger>
        <TooltipContent>ONにするとステータスに色を設定できます</TooltipContent>
      </Tooltip>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <button
            type="button"
            disabled={!value.enabled}
            aria-label="色を選択"
            className={cn(
              "rounded-md border border-border disabled:cursor-not-allowed disabled:opacity-40",
              swatchClassName,
            )}
            style={{ backgroundColor: value.value }}
          />
        </PopoverTrigger>
        <PopoverContent className="w-48 p-2" align="start">
          <div className="grid grid-cols-6 gap-1">
            {STATUS_COLOR_PRESETS.map((preset) => (
              <button
                key={preset.value}
                type="button"
                title={preset.name}
                aria-label={preset.name}
                onClick={() => {
                  onChange({ ...value, value: preset.value })
                  setOpen(false)
                }}
                className="relative h-5 w-5 rounded-full border border-border/50"
                style={{ backgroundColor: preset.value }}
              >
                {value.value === preset.value && (
                  <Check className="absolute inset-0 m-auto h-3 w-3 text-white drop-shadow" />
                )}
              </button>
            ))}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
