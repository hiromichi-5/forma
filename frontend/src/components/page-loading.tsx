import { Loader2 } from "lucide-react"

/** ページの読み込み中に表示する。 */
export function PageLoading() {
  return (
    <div className="flex min-h-[400px] items-center justify-center" role="status">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      <span className="sr-only">読み込み中</span>
    </div>
  )
}
