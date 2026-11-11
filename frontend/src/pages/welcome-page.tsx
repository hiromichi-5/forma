import { useRef, useState } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { Link, Navigate } from "react-router"
import { FileText, Home, Plus, Search } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Annotation } from "@/components/welcome/annotation"
import { DemoFormSection } from "@/components/welcome/demo-form-section"
import {
  createDemoQueryClient,
  createDemoWorkspace,
  type DemoFormListItem,
} from "@/components/welcome/demo-workspace"
import { useAuth } from "@/hooks/use-auth"
import { useSignupPrompt } from "@/hooks/use-signup-prompt"
import { cn } from "@/lib/utils"

export default function WelcomePage() {
  const { user } = useAuth()
  const promptSignup = useSignupPrompt()
  const [queryClient] = useState(createDemoQueryClient)
  const [workspace] = useState(() => createDemoWorkspace(queryClient))
  const [forms, setForms] = useState(() => workspace.formList())
  const [selectedFormId, setSelectedFormId] = useState(forms[0].id)
  const [searchQuery, setSearchQuery] = useState("")
  const formSectionRef = useRef<HTMLElement>(null)

  if (user) {
    return <Navigate to="/" replace />
  }

  const openForm = (formId: string) => {
    setSelectedFormId(formId)
    formSectionRef.current?.scrollIntoView({ behavior: "smooth", block: "start" })
  }

  const filteredForms = forms.filter((form) =>
    form.title.toLowerCase().includes(searchQuery.toLowerCase()),
  )

  return (
    <QueryClientProvider client={queryClient}>
      <div className="flex h-screen flex-col overflow-hidden md:flex-row">
        <DemoSidebar forms={forms} selectedFormId={selectedFormId} onSelectForm={openForm} />

        <div className="flex items-center justify-between border-b bg-blue-50 px-4 py-2 md:hidden">
          <Brand />
          <div className="flex gap-2">
            <Button asChild size="sm" variant="outline">
              <Link to="/login">ログイン</Link>
            </Button>
            <Button asChild size="sm">
              <Link to="/signup">無料ではじめる</Link>
            </Button>
          </div>
        </div>

        <main className="flex-1 overflow-auto bg-white">
          <div className="container mx-auto space-y-6 p-6 pt-10">
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

            <div>
              <Annotation arrow="down" align="start" className="ml-6">
                クリックすると回答の管理画面を開けます
              </Annotation>
              <div className="grid gap-3 pb-24 sm:grid-cols-2 xl:grid-cols-3">
                {filteredForms.map((form) => (
                  <FormCard
                    key={form.id}
                    form={form}
                    selected={form.id === selectedFormId}
                    onClick={() => openForm(form.id)}
                  />
                ))}
                <div className="relative">
                  <button
                    type="button"
                    onClick={promptSignup}
                    className="flex h-full min-h-[84px] w-full flex-col items-center justify-center gap-1.5 rounded-xl border border-dashed border-muted-foreground/40 bg-transparent p-4 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none"
                  >
                    <Plus className="h-5 w-5" />
                    <span className="text-sm font-medium">フォームを追加する</span>
                  </button>
                  <Annotation arrow="up" align="end" className="absolute top-full right-6 mt-1">
                    Googleフォームを登録することができます
                  </Annotation>
                </div>
              </div>
            </div>

            <section
              ref={formSectionRef}
              className="min-h-screen scroll-mt-6 border-t-2 border-dashed border-gray-300 pt-10 pb-24"
            >
              <DemoFormSection
                key={selectedFormId}
                workspace={workspace}
                formId={selectedFormId}
                onSynced={() => setForms(workspace.formList())}
              />
            </section>
          </div>
        </main>
      </div>
    </QueryClientProvider>
  )
}

function Brand() {
  return (
    <div className="flex items-center gap-2">
      <img src="/favicon.svg" alt="forma Logo" className="h-6 w-6" />
      <span className="text-xl font-bold tracking-tight">forma</span>
    </div>
  )
}

type DemoSidebarProps = {
  forms: DemoFormListItem[]
  selectedFormId: string
  onSelectForm: (formId: string) => void
}

function DemoSidebar({ forms, selectedFormId, onSelectForm }: DemoSidebarProps) {
  return (
    <aside className="hidden w-64 shrink-0 flex-col border-r bg-blue-50 md:flex">
      <div className="border-b p-4">
        <Brand />
      </div>

      <nav className="flex-1 overflow-y-auto p-2 [&_[data-slot=button]]:hover:bg-blue-100">
        <Button variant="ghost" className="mb-2 w-full justify-start gap-3 bg-blue-100">
          <Home className="h-5 w-5" />
          <span>ホーム</span>
        </Button>
        <div className="mb-1 flex items-center gap-2 px-4 py-2">
          <FileText className="h-4 w-4" />
          <span className="text-sm font-medium">フォーム</span>
        </div>
        <div className="space-y-1">
          {forms.map((form) => (
            <div key={form.id} className="pl-6">
              <Button
                variant="ghost"
                size="sm"
                className={cn(
                  "h-8 w-full min-w-0 justify-start rounded-2xl pr-2 text-sm",
                  form.id === selectedFormId && "bg-blue-100 font-medium",
                )}
                onClick={() => onSelectForm(form.id)}
              >
                <span className="truncate">{form.title}</span>
              </Button>
            </div>
          ))}
        </div>
      </nav>

      <div className="space-y-2 p-4">
        <div className="flex justify-center">
          <Annotation arrow="down" align="center">
            アカウント作成はこちらから！
          </Annotation>
        </div>
        <Button asChild size="lg" className="h-12 w-full rounded-full text-base font-bold">
          <Link to="/signup">無料ではじめる</Link>
        </Button>
        <Button asChild variant="outline" size="lg" className="w-full rounded-full">
          <Link to="/login">ログイン</Link>
        </Button>
      </div>
    </aside>
  )
}

function FormCard({
  form,
  selected,
  onClick,
}: {
  form: DemoFormListItem
  selected: boolean
  onClick: () => void
}) {
  return (
    <Card
      asChild
      className={cn(
        "border-border/60 bg-card/95 text-left shadow-none transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none",
        selected && "border-primary/40 bg-blue-50/60",
      )}
    >
      <button type="button" onClick={onClick}>
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
      </button>
    </Card>
  )
}

function formatCompactDate(date: Date): string {
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(date)
}
