import { Badge } from "@/components/ui/badge"
import { Textarea } from "@/components/ui/textarea"

export function UpcomingBadge() {
  return (
    <span className="shrink-0 rounded-full bg-blue-50 px-2 py-0.5 text-[11px] leading-4 font-medium text-blue-600">
      現在開発中
    </span>
  )
}

type DemoMemoProps = {
  defaultValue: string
  onChange: (memo: string) => void
}

export function DemoMemo({ defaultValue, onChange }: DemoMemoProps) {
  return (
    <>
      <div className="mb-2 flex items-center gap-2">
        <Badge variant="secondary">メモ</Badge>
        <span className="text-xs text-muted-foreground">
          回答者には表示されません。チーム内で共有できます
        </span>
        <UpcomingBadge />
      </div>
      <Textarea
        placeholder="メモを入力..."
        aria-label="メモ"
        defaultValue={defaultValue}
        onChange={(e) => onChange(e.target.value)}
        className="flex-1 resize-none"
      />
    </>
  )
}
