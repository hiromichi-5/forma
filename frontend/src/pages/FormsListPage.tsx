import { useEffect, useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { Input } from "@/components/ui/input"
import { Card, CardContent } from "@/components/ui/card"
import { Search } from "lucide-react"
import { AppLayout } from "@/components/app-layout"
import { apiClient } from "@/lib/api"
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
  const [forms, setForms] = useState<FormListItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState("")

  const loadForms = async () => {
    setIsLoading(true)
    setErrorMessage("")
    try {
      const formsResponse = await apiClient.getForms()
      const formItems = await buildFormList(formsResponse.forms)
      setForms(formItems)
    } catch (error) {
      console.error("Failed to load forms:", error)
      setErrorMessage("フォーム一覧の取得に失敗しました")
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    let isActive = true

    const load = async () => {
      setIsLoading(true)
      setErrorMessage("")
      try {
        const formsResponse = await apiClient.getForms()
        if (!isActive) return
        const formItems = await buildFormList(formsResponse.forms)
        if (!isActive) return
        setForms(formItems)
      } catch (error) {
        if (!isActive) return
        console.error("Failed to load forms:", error)
        setErrorMessage("フォーム一覧の取得に失敗しました")
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }

    load()

    return () => {
      isActive = false
    }
  }, [])

  const filteredForms = useMemo(
    () => forms.filter((form) => form.title.toLowerCase().includes(searchQuery.toLowerCase())),
    [forms, searchQuery]
  )

  return (
    <AppLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-foreground">フォーム一覧</h1>
            <p className="text-sm text-muted-foreground mt-1">管理しているGoogleフォームの一覧</p>
          </div>
          <RegisterFormDialog onRegistered={loadForms} />
        </div>

        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="フォーム名で検索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        {isLoading && (
          <Card>
            <CardContent className="p-6 text-sm text-muted-foreground">フォーム一覧を読み込み中...</CardContent>
          </Card>
        )}

        {!isLoading && errorMessage && (
          <Card className="border-destructive/40">
            <CardContent className="p-6 text-sm text-destructive">{errorMessage}</CardContent>
          </Card>
        )}

        {!isLoading && !errorMessage && filteredForms.length === 0 && (
          <Card>
            <CardContent className="p-6 text-sm text-muted-foreground">
              条件に一致するフォームがありません
            </CardContent>
          </Card>
        )}

        {!isLoading && !errorMessage && filteredForms.length > 0 && (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
            {filteredForms.map((form) => (
              <Card
                asChild
                key={form.id}
                className="border-border/60 bg-card/95 shadow-none transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <Link to={`/forms/${form.id}`}>
                  <CardContent className="p-4">
                    <h3 className="line-clamp-2 text-sm font-semibold leading-5 text-foreground">
                      {form.title}
                    </h3>
                    <p className="mt-2 text-xs text-muted-foreground">
                      {form.responseCount}件
                      <span className="mx-1.5">•</span>
                      {form.latestSubmittedAt
                        ? `最終提出 ${formatCompactDate(form.latestSubmittedAt)}`
                        : "回答なし"}
                    </p>
                  </CardContent>
                </Link>
              </Card>
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  )
}

async function buildFormList(forms: FormSummary[]): Promise<FormListItem[]> {
  const results = await Promise.all(
    forms.map(async (form) => {
      try {
        const ticketsRes = await apiClient.getTickets(form.id)
        const latestSubmittedAt = ticketsRes.tickets.reduce<Date | null>((latest, ticket) => {
          const submittedAt = new Date(ticket.submitted_at)
          if (!latest || submittedAt > latest) return submittedAt
          return latest
        }, null)

        return {
          id: form.id,
          title: form.title,
          responseCount: ticketsRes.tickets.length,
          latestSubmittedAt,
        }
      } catch {
        return {
          id: form.id,
          title: form.title,
          responseCount: 0,
          latestSubmittedAt: null,
        }
      }
    })
  )
  return results
}

function formatCompactDate(date: Date): string {
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(date)
}
