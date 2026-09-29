import { useEffect, useRef } from "react"

type InfiniteScrollTriggerProps = {
  hasMore: boolean
  loading: boolean
  onLoadMore: () => void
}

export function InfiniteScrollTrigger({ hasMore, loading, onLoadMore }: InfiniteScrollTriggerProps) {
  const ref = useRef<HTMLDivElement>(null)
  const onLoadMoreRef = useRef(onLoadMore)
  useEffect(() => {
    onLoadMoreRef.current = onLoadMore
  })

  // 読み込み完了ごとに監視し直すことで、1ページ分では画面が埋まらない場合も続けて読み込む。
  useEffect(() => {
    const element = ref.current
    if (!element || !hasMore || loading) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) onLoadMoreRef.current()
      },
      { rootMargin: "200px" }
    )
    observer.observe(element)
    return () => observer.disconnect()
  }, [hasMore, loading])

  return (
    <div ref={ref} className="py-2 text-center text-sm text-muted-foreground">
      {loading ? "読み込み中..." : null}
    </div>
  )
}
