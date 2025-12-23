import type React from "react";

import { useState, useEffect } from "react";
import { useNavigate, useLocation, Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  LayoutGrid,
  LogOut,
  Settings,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  Home,
  FileText,
  Users,
  RefreshCw,
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
import { toast } from "sonner";
import { MembersDialog } from "./members-dialog";

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
  const [membersDialogOpen, setMembersDialogOpen] = useState(false);
  const [selectedFormId, setSelectedFormId] = useState<string | null>(null);
  const { logout } = useAuth();

  const isFormsListPage = location.pathname === "/";

  // フォーム一覧を取得
  useEffect(() => {
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
    loadForms();
  }, []);

  const getBreadcrumbs = () => {
    const pathSegments = location.pathname.split("/").filter(Boolean);
    const breadcrumbs = [{ label: "ホーム", path: "/" }];

    if (pathSegments.length > 0) {
      if (pathSegments[0] === "forms" && pathSegments[1]) {
        const formId = pathSegments[1];
        const form = forms.find((f) => f.form_id === formId);
        breadcrumbs.push({
          label: form?.title || "フォーム管理",
          path: `/forms/${formId}`,
        });
      }
    }

    return breadcrumbs;
  };

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  const breadcrumbs = getBreadcrumbs();

  return (
    <div className="min-h-screen bg-muted/30 flex">
      <aside
        className={cn(
          "bg-card border-r transition-all duration-300 flex flex-col",
          isCollapsed ? "w-16" : "w-64"
        )}
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
          {!isCollapsed && breadcrumbs.length > 1 && (
            <div className="mb-4 px-2 py-1 text-xs text-muted-foreground">
              <div className="flex items-center gap-1 flex-wrap">
                {breadcrumbs.map((crumb, index) => (
                  <div key={crumb.path} className="flex items-center gap-1">
                    {index > 0 && <span>/</span>}
                    <Link
                      to={crumb.path}
                      className={cn(
                        "hover:text-foreground transition-colors",
                        index === breadcrumbs.length - 1 &&
                          "text-foreground font-medium"
                      )}
                    >
                      {crumb.label}
                    </Link>
                  </div>
                ))}
              </div>
            </div>
          )}

          <Button
            variant={isFormsListPage ? "secondary" : "ghost"}
            className={cn(
              "w-full justify-start gap-3 mb-2",
              isCollapsed && "justify-center"
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
                      location.pathname === `/forms/${form.form_id}`;
                    return (
                      <div
                        key={form.form_id}
                        className={cn(
                          "group flex items-center gap-1 rounded-md transition-colors",
                          isActive && "bg-secondary"
                        )}
                      >
                        <Button
                          variant="ghost"
                          size="sm"
                          className={cn(
                            "flex-1 justify-start text-sm h-8 pl-6 pr-2 min-w-0",
                            isActive && "font-medium"
                          )}
                          onClick={() => navigate(`/forms/${form.form_id}`)}
                        >
                          <span className="truncate">{form.title}</span>
                        </Button>

                        <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={async (e) => {
                              e.stopPropagation();
                              try {
                                await apiClient.syncForm(form.form_id);
                                toast.success("フォームを同期しました");
                              } catch (error) {
                                toast.error("同期に失敗しました");
                                console.error("Failed to sync form:", error);
                              }
                            }}
                            title="同期"
                          >
                            <RefreshCw className="h-3.5 w-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedFormId(form.form_id);
                              setMembersDialogOpen(true);
                            }}
                            title="メンバー管理"
                          >
                            <Users className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </div>
                    );
                  })
                )}
              </CollapsibleContent>
            </Collapsible>
          )}
        </nav>

        <div className="p-2 border-t space-y-1">
          <Button
            variant="ghost"
            className={cn(
              "w-full justify-start gap-3",
              isCollapsed && "justify-center"
            )}
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

      <main className="flex-1 overflow-auto">
        <div className="container mx-auto p-6">{children}</div>
      </main>

      {selectedFormId && (
        <MembersDialog
          formId={selectedFormId}
          open={membersDialogOpen}
          onOpenChange={setMembersDialogOpen}
        />
      )}
    </div>
  );
}
