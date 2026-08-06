import type React from "react";

import { useState, useEffect } from "react";
import { useNavigate, useLocation, Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
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
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/useAuth";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { apiClient } from "@/lib/api";
import type { FormSummary } from "@/types";
import { RegisterFormDialog } from "@/components/register-form-dialog";

type AppLayoutProps = {
  children: React.ReactNode;
};

export function AppLayout({ children }: AppLayoutProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isFormsOpen, setIsFormsOpen] = useState(true);
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [loadingForms, setLoadingForms] = useState(false);
  const { logout } = useAuth();

  const isFormsListPage = location.pathname === "/";
  const isSettingsPage = location.pathname === "/settings";

  // フォーム一覧を取得
  const loadForms = async () => {
    try {
      setLoadingForms(true);
      const response = await apiClient.getForms();
      setForms(response.forms);
    } catch (error) {
      console.error("Failed to load forms:", error);
    } finally {
      setLoadingForms(false);
    }
  };

  useEffect(() => {
    loadForms();
  }, []);

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <div className="h-screen bg-muted/30 flex overflow-hidden">
      <aside
        className={cn(
          "bg-blue-50 border-r transition-all duration-300 flex flex-col shrink-0 z-10",
          "[&_[data-slot=button]]:hover:bg-blue-100",
          isCollapsed ? "w-16" : "w-64"
        )}
      >
        <div className="p-4 border-b flex items-center justify-between">
          {!isCollapsed && (
            <Link
              to="/"
              className="flex items-center gap-2 -mx-2 -my-1 rounded-md px-2 py-1 transition-colors hover:bg-accent"
            >
              <img src="/favicon.svg" alt="forma Logo" className="w-6 h-6" />
              <span className="font-bold text-xl tracking-tight">forma</span>
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

        <nav className="flex-1 p-2 overflow-y-auto">
          <Button
            variant="ghost"
            className={cn(
              "w-full justify-start gap-3 mb-2",
              isCollapsed && "justify-center",
              isFormsListPage && "bg-blue-100"
            )}
            onClick={() => navigate("/")}
          >
            <Home className="h-5 w-5" />
            {!isCollapsed && <span>ホーム</span>}
          </Button>

          {!isCollapsed && (
            <Collapsible open={isFormsOpen} onOpenChange={setIsFormsOpen}>
              <CollapsibleTrigger asChild>
                <Button
                  variant="ghost"
                  className="w-full justify-between gap-2 mb-1"
                >
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
                  <div className="px-2 py-1 text-xs text-muted-foreground">
                    読み込み中...
                  </div>
                ) : forms.length === 0 ? (
                  <div className="px-2 py-1 text-xs text-muted-foreground">
                    フォームがありません
                  </div>
                ) : (
                  forms.map((form) => {
                    const isActive =
                      location.pathname === `/forms/${form.id}`;
                    return (
                      <div key={form.id} className="pl-6">
                        <Button
                          variant="ghost"
                          size="sm"
                          className={cn(
                            "w-full justify-start text-sm h-8 pr-2 min-w-0 rounded-2xl",
                            isActive && "bg-blue-100 font-medium"
                          )}
                          onClick={() => navigate(`/forms/${form.id}`)}
                        >
                          <span className="truncate">{form.title}</span>
                        </Button>
                      </div>
                    );
                  })
                )}
                <div className="pl-6 mt-4">
                  <RegisterFormDialog
                    onRegistered={loadForms}
                    trigger={
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start gap-2 text-sm h-8 pr-2 min-w-0 rounded-2xl border border-dashed border-muted-foreground/40 text-muted-foreground"
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

        <div className="p-2 border-t space-y-1">
          <Button
            variant="ghost"
            className={cn(
              "w-full justify-start gap-3",
              isCollapsed && "justify-center",
              isSettingsPage && "bg-blue-100"
            )}
            onClick={() => navigate("/settings")}
          >
            <Settings className="h-5 w-5" />
            {!isCollapsed && <span>設定</span>}
          </Button>
          <Button
            variant="ghost"
            className={cn(
              "w-full justify-start gap-3",
              isCollapsed && "justify-center"
            )}
            onClick={handleLogout}
          >
            <LogOut className="h-5 w-5" />
            {!isCollapsed && <span>ログアウト</span>}
          </Button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto bg-white">
        <div className="container mx-auto p-6">{children}</div>
      </main>
    </div>
  );
}
