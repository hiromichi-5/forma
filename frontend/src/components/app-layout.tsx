import { Suspense, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { useNavigate, useLocation, Link, Outlet } from "react-router"
import { PageLoading } from "@/components/page-loading"
import { Button } from "@/components/ui/button"
import {
  LogOut,
  Settings,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  Home,
  FileText,
  Plus,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { useAuth } from "@/hooks/use-auth"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { formQueries } from "@/queries/forms"
import { RegisterFormDialog } from "@/components/register-form-dialog"

/** ログイン後の画面の共通レイアウト。ページを移動してもサイドバーの状態は保たれる。 */
export function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [isCollapsed, setIsCollapsed] = useState(false)
  const [isFormsOpen, setIsFormsOpen] = useState(true)
  const { data: forms = [], isPending: loadingForms } = useQuery(formQueries.list())
  const { logout } = useAuth()

  const isFormsListPage = location.pathname === "/"
  const isSettingsPage = location.pathname === "/settings"

  const handleLogout = async () => {
    await logout()
    void navigate("/login")
  }

  return (
    <div className="flex h-screen overflow-hidden bg-muted/30">
      <aside
        className={cn(
          "z-10 flex shrink-0 flex-col border-r bg-blue-50 transition-all duration-300",
          "[&_[data-slot=button]]:hover:bg-blue-100",
          isCollapsed ? "w-16" : "w-64",
        )}
      >
        <div className="flex items-center justify-between border-b p-4">
          {!isCollapsed && (
            <Link
              to="/"
              className="-mx-2 -my-1 flex items-center gap-2 rounded-md px-2 py-1 transition-colors hover:bg-accent"
            >
              <img src="/favicon.svg" alt="forma Logo" className="h-6 w-6" />
              <span className="text-xl font-bold tracking-tight">forma</span>
            </Link>
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setIsCollapsed(!isCollapsed)}
            className="ml-auto"
          >
            {isCollapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>

        <nav className="flex-1 overflow-y-auto p-2">
          <Button
            variant="ghost"
            className={cn(
              "mb-2 w-full justify-start gap-3",
              isCollapsed && "justify-center",
              isFormsListPage && "bg-blue-100",
            )}
            onClick={() => navigate("/")}
          >
            <Home className="h-5 w-5" />
            {!isCollapsed && <span>ホーム</span>}
          </Button>

          {!isCollapsed && (
            <Collapsible open={isFormsOpen} onOpenChange={setIsFormsOpen}>
              <CollapsibleTrigger asChild>
                <Button variant="ghost" className="mb-1 w-full justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <FileText className="h-4 w-4" />
                    <span className="text-sm font-medium">フォーム</span>
                  </div>
                  {isFormsOpen ? (
                    <ChevronUp className="h-4 w-4" />
                  ) : (
                    <ChevronDown className="h-4 w-4" />
                  )}
                </Button>
              </CollapsibleTrigger>
              <CollapsibleContent className="space-y-1">
                {loadingForms ? (
                  <div className="px-2 py-1 text-xs text-muted-foreground">読み込み中...</div>
                ) : forms.length === 0 ? (
                  <div className="px-2 py-1 text-xs text-muted-foreground">
                    フォームがありません
                  </div>
                ) : (
                  forms.map((form) => {
                    const isActive = location.pathname === `/forms/${form.id}`
                    return (
                      <div key={form.id} className="pl-6">
                        <Button
                          variant="ghost"
                          size="sm"
                          className={cn(
                            "h-8 w-full min-w-0 justify-start rounded-2xl pr-2 text-sm",
                            isActive && "bg-blue-100 font-medium",
                          )}
                          onClick={() => navigate(`/forms/${form.id}`)}
                        >
                          <span className="truncate">{form.title}</span>
                        </Button>
                      </div>
                    )
                  })
                )}
                <div className="mt-4 pl-6">
                  <RegisterFormDialog
                    trigger={
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 w-full min-w-0 justify-start gap-2 rounded-2xl border border-dashed border-muted-foreground/40 pr-2 text-sm text-muted-foreground"
                      >
                        <Plus className="h-3.5 w-3.5 shrink-0" />
                        <span className="truncate">フォームを追加する</span>
                      </Button>
                    }
                  />
                </div>
              </CollapsibleContent>
            </Collapsible>
          )}
        </nav>

        <div className="space-y-1 border-t p-2">
          <Button
            variant="ghost"
            className={cn(
              "w-full justify-start gap-3",
              isCollapsed && "justify-center",
              isSettingsPage && "bg-blue-100",
            )}
            onClick={() => navigate("/settings")}
          >
            <Settings className="h-5 w-5" />
            {!isCollapsed && <span>設定</span>}
          </Button>
          <Button
            variant="ghost"
            className={cn("w-full justify-start gap-3", isCollapsed && "justify-center")}
            onClick={handleLogout}
          >
            <LogOut className="h-5 w-5" />
            {!isCollapsed && <span>ログアウト</span>}
          </Button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto bg-white">
        <div className="container mx-auto p-6">
          <Suspense fallback={<PageLoading />}>
            <Outlet />
          </Suspense>
        </div>
      </main>
    </div>
  )
}
