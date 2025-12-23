import type React from "react"

import { useState } from "react"
import { useNavigate, useLocation } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { LayoutGrid, LogOut, Settings, ChevronLeft, ChevronRight } from "lucide-react"
import { cn } from "@/lib/utils"

interface AppLayoutProps {
  children: React.ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const [isCollapsed, setIsCollapsed] = useState(false)

  const isFormsListPage = location.pathname === "/forms-list"

  return (
    <div className="min-h-screen bg-muted/30 flex">
      <aside
        className={cn("bg-card border-r transition-all duration-300 flex flex-col", isCollapsed ? "w-16" : "w-64")}
      >
        <div className="p-4 border-b flex items-center justify-between">
          {!isCollapsed && (
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-primary rounded flex items-center justify-center">
                <LayoutGrid className="h-5 w-5 text-primary-foreground" />
              </div>
              <span className="font-bold text-lg">フォーム管理</span>
            </div>
          )}
          <Button variant="ghost" size="icon" onClick={() => setIsCollapsed(!isCollapsed)} className="ml-auto">
            {isCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </Button>
        </div>

        <nav className="flex-1 p-2">
          <Button
            variant={isFormsListPage ? "secondary" : "ghost"}
            className={cn("w-full justify-start gap-3", isCollapsed && "justify-center")}
            onClick={() => navigate("/forms-list")}
          >
            <LayoutGrid className="h-5 w-5" />
            {!isCollapsed && <span>フォーム一覧</span>}
          </Button>
        </nav>

        <div className="p-2 border-t space-y-1">
          <Button variant="ghost" className={cn("w-full justify-start gap-3", isCollapsed && "justify-center")}>
            <Settings className="h-5 w-5" />
            {!isCollapsed && <span>設定</span>}
          </Button>
          <Button
            variant="ghost"
            className={cn("w-full justify-start gap-3", isCollapsed && "justify-center")}
            onClick={() => navigate("/login-new")}
          >
            <LogOut className="h-5 w-5" />
            {!isCollapsed && <span>ログアウト</span>}
          </Button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto">
        <div className="container mx-auto p-6">{children}</div>
      </main>
    </div>
  )
}
