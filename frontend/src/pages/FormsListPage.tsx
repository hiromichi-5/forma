import { useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Search, Plus, FileText } from "lucide-react"
import { AppLayout } from "@/components/app-layout"
import { apiClient } from "@/lib/api"
import type { FormSummary, TicketSummary } from "@/types"

type FormListItem = {
  id: string
  title: string
  responseCount: number
  newCount: number
  lastUpdated: Date | null
}

const buildFormList = (forms: FormSummary[], tickets: TicketSummary[]): FormListItem[] => {
  const ticketMap = new Map<string, TicketSummary[]>()

  tickets.forEach((ticket) => {
    const bucket = ticketMap.get(ticket.form_id)
    if (bucket) {
      bucket.push(ticket)
    } else {
      ticketMap.set(ticket.form_id, [ticket])
    }
  })

  return forms.map((form) => {
    const formTickets = ticketMap.get(form.form_id) ?? []
    const responseCount = formTickets.length
    const newCount = formTickets.filter((ticket) => ticket.status === "new").length
    const lastUpdated = formTickets.reduce<Date | null>((latest, ticket) => {
      const updatedAt = new Date(ticket.updated_at)
      if (!latest || updatedAt > latest) {
        return updatedAt
      }
      return latest
    }, null)

    return {
      id: form.form_id,
      title: form.title,
      responseCount,
      newCount,
      lastUpdated,
    }
  })
}

export default function FormsListPage() {
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = useState("")
  const [forms, setForms] = useState<FormListItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState("")

  useEffect(() => {
    let isActive = true

    const loadForms = async () => {
      setIsLoading(true)
      setErrorMessage("")
      try {
        // 外部API(バックエンド)との同期のための処理
        const [formsResponse, ticketsResponse] = await Promise.all([
          apiClient.getForms(),
          apiClient.getTickets(),
        ])
        if (!isActive) return
        setForms(buildFormList(formsResponse.forms, ticketsResponse.tickets))
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

    // 外部API(バックエンド)との同期のための処理
    loadForms()

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
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            フォーム連携
          </Button>
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

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {isLoading && (
            <Card className="md:col-span-2 lg:col-span-3">
              <CardContent className="p-6 text-sm text-muted-foreground">フォーム一覧を読み込み中...</CardContent>
            </Card>
          )}

          {!isLoading && errorMessage && (
            <Card className="md:col-span-2 lg:col-span-3 border-destructive/40">
              <CardContent className="p-6 text-sm text-destructive">{errorMessage}</CardContent>
            </Card>
          )}

          {!isLoading && !errorMessage && filteredForms.length === 0 && (
            <Card className="md:col-span-2 lg:col-span-3">
              <CardContent className="p-6 text-sm text-muted-foreground">
                条件に一致するフォームがありません
              </CardContent>
            </Card>
          )}

          {!isLoading &&
            !errorMessage &&
            filteredForms.map((form) => (
            <Card
              key={form.id}
              className="hover:shadow-md transition-shadow cursor-pointer"
              onClick={() => navigate(`/forms/${form.id}`)}
            >
              <CardContent className="p-6 space-y-4">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="font-semibold text-lg text-foreground mb-1">{form.title}</h3>
                    <p className="text-xs text-muted-foreground">
                      {form.lastUpdated ? form.lastUpdated.toLocaleDateString("ja-JP") : "更新情報なし"}
                    </p>
                  </div>
                  {form.newCount > 0 && <Badge className="bg-blue-500 text-white">{form.newCount}件の新規</Badge>}
                </div>

                <div className="flex items-center gap-4 text-sm">
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <FileText className="h-4 w-4" />
                    <span>{form.responseCount}件の回答</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </AppLayout>
  )
}
