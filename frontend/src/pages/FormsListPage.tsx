import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Search, Plus, Users, FileText } from "lucide-react"
import { AppLayout } from "@/components/app-layout"

interface Form {
  id: string
  title: string
  responseCount: number
  newCount: number
  lastUpdated: Date
  members: number
}

const mockForms: Form[] = [
  {
    id: "form-001",
    title: "新規お問い合わせフォーム",
    responseCount: 24,
    newCount: 3,
    lastUpdated: new Date("2024-01-16T11:00:00"),
    members: 3,
  },
  {
    id: "form-002",
    title: "サポートリクエスト",
    responseCount: 18,
    newCount: 2,
    lastUpdated: new Date("2024-01-15T14:30:00"),
    members: 2,
  },
  {
    id: "form-003",
    title: "イベント参加申込",
    responseCount: 45,
    newCount: 8,
    lastUpdated: new Date("2024-01-14T09:20:00"),
    members: 4,
  },
]

export default function FormsListPage() {
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = useState("")

  const filteredForms = mockForms.filter((form) => form.title.toLowerCase().includes(searchQuery.toLowerCase()))

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
          {filteredForms.map((form) => (
            <Card
              key={form.id}
              className="hover:shadow-md transition-shadow cursor-pointer"
              onClick={() => navigate(`/forms/${form.id}`)}
            >
              <CardContent className="p-6 space-y-4">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="font-semibold text-lg text-foreground mb-1">{form.title}</h3>
                    <p className="text-xs text-muted-foreground">{form.lastUpdated.toLocaleDateString("ja-JP")}</p>
                  </div>
                  {form.newCount > 0 && <Badge className="bg-blue-500 text-white">{form.newCount}件の新規</Badge>}
                </div>

                <div className="flex items-center gap-4 text-sm">
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <FileText className="h-4 w-4" />
                    <span>{form.responseCount}件の回答</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <Users className="h-4 w-4" />
                    <span>{form.members}人</span>
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
