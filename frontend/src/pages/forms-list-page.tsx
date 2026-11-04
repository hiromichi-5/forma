import { useMemo, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { Link } from "react-router"
import { Input } from "@/components/ui/input"
import { Card, CardContent } from "@/components/ui/card"
import { Plus, Search } from "lucide-react"
import { formQueries } from "@/queries/forms"
import type { FormSummary } from "@/types"
import { RegisterFormDialog } from "@/components/register-form-dialog"
type FormListItem = {
  id: string
  title: string
  responseCount: number
  latestSubmittedAt: Date | null
}

export default function FormsListPage() {
  const [searchQuery, setSearchQuery] = useState("")
  const {
    data: forms = [],
    isPending: isLoading,
    isError,
  } = useQuery({
    ...formQueries.list(),
    select: (res) => res.forms.map(toFormListItem),
  })
  const errorMessage = isError ? "フォーム一覧の取得に失敗しました" : ""

  const filteredForms = useMemo(
    () => forms.filter((form) => form.title.toLowerCase().includes(searchQuery.toLowerCase())),
    [forms, searchQuery],
  )

  return (
    <>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">フォーム一覧</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理しているGoogleフォームの一覧</p>
        </div>

        <div className="relative">
          <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="フォーム名で検索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        {isLoading && (
          <Card>
            <CardContent className="p-6 text-sm text-muted-foreground">
              フォーム一覧を読み込み中...
            </CardContent>
          </Card>
        )}

        {!isLoading && errorMessage && (
          <Card className="border-destructive/40">
            <CardContent className="p-6 text-sm text-destructive">{errorMessage}</CardContent>
          </Card>
        )}

        {!isLoading && !errorMessage && filteredForms.length === 0 && forms.length > 0 && (
          <Card>
            <CardContent className="p-6 text-sm text-muted-foreground">
              条件に一致するフォームがありません
            </CardContent>
          </Card>
        )}

        {!isLoading && !errorMessage && (filteredForms.length > 0 || forms.length === 0) && (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
            {filteredForms.map((form) => (
              <Card
                asChild
                key={form.id}
                className="border-border/60 bg-card/95 shadow-none transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none"
              >
                <Link to={`/forms/${form.id}`}>
                  <CardContent className="p-4">
                    <h3 className="line-clamp-2 text-sm leading-5 font-semibold text-foreground">
                      {form.title}
                    </h3>
                    <p className="mt-2 text-xs text-muted-foreground">
                      {form.responseCount}件<span className="mx-1.5">•</span>
                      {form.latestSubmittedAt
                        ? `最終提出 ${formatCompactDate(form.latestSubmittedAt)}`
                        : "回答なし"}
                    </p>
                  </CardContent>
                </Link>
              </Card>
            ))}
            <RegisterFormDialog
              trigger={
                <button
                  type="button"
                  className="flex min-h-[84px] flex-col items-center justify-center gap-1.5 rounded-xl border border-dashed border-muted-foreground/40 bg-transparent p-4 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none"
                >
                  <Plus className="h-5 w-5" />
                  <span className="text-sm font-medium">フォームを追加する</span>
                </button>
              }
            />
          </div>
        )}
      </div>
    </>
  )
}

function toFormListItem(form: FormSummary): FormListItem {
  return {
    id: form.id,
    title: form.title,
    responseCount: form.ticket_count,
    latestSubmittedAt: form.latest_submitted_at ? new Date(form.latest_submitted_at) : null,
  }
}

function formatCompactDate(date: Date): string {
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(date)
}
