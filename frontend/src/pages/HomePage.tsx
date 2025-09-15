import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import {
  LayoutDashboard,
  FileSpreadsheet,
  ListChecks,
  Clock,
  TrendingUp,
  Activity,
  RefreshCw,
  BarChart3,
} from "lucide-react";
import { Button } from "../components/ui/Button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/Card";
import { Badge } from "../components/ui/Badge";
import { apiClient, ApiError } from "../lib/api";
import type { FormSummary, Ticket } from "../types";

interface DashboardStats {
  totalForms: number;
  totalTickets: number;
  pendingTickets: number;
  recentResponses: number;
}

export function HomePage() {
  const [stats, setStats] = useState<DashboardStats>({
    totalForms: 0,
    totalTickets: 0,
    pendingTickets: 0,
    recentResponses: 0,
  });
  const [recentTickets, setRecentTickets] = useState<Ticket[]>([]);
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [syncingForms, setSyncingForms] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string | null>(null);

  const loadDashboardData = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const formsResponse = await apiClient.getForms();
      const forms = formsResponse.forms;

      const ticketPromises = forms.map((form) =>
        apiClient.getTickets(form.form_id).catch((err) => {
          console.warn(`Failed to get tickets for form ${form.form_id}:`, err);
          return { tickets: [] };
        })
      );

      const responsePromises = forms.map((form) =>
        apiClient.getResponses(form.form_id).catch((err) => {
          console.warn(
            `Failed to get responses for form ${form.form_id}:`,
            err
          );
          return { responses: [] };
        })
      );

      const [ticketResults, responseResults] = await Promise.all([
        Promise.all(ticketPromises),
        Promise.all(responsePromises),
      ]);

      const allTickets = ticketResults.flatMap(
        (result) => result.tickets || []
      );
      const allResponses = responseResults.flatMap(
        (result) => result.responses || []
      );

      const now = new Date();
      const last24Hours = new Date(now.getTime() - 24 * 60 * 60 * 1000);

      const recentResponses = allResponses.filter(
        (r) => r && r.submitted_at && new Date(r.submitted_at) > last24Hours
      ).length;

      const pendingTickets = allTickets.filter(
        (t) => t && (t.status === "new" || t.status === "in_progress")
      ).length;

      setStats({
        totalForms: forms.length,
        totalTickets: allTickets.length,
        pendingTickets,
        recentResponses,
      });

      setForms(forms);
      setRecentTickets(allTickets.filter((t) => t).slice(0, 5));
    } catch (err) {
      if (err instanceof ApiError) {
        setError("データの読み込みに失敗しました");
      } else {
        setError("ネットワークエラーが発生しました");
      }
      console.error("Failed to load dashboard data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadDashboardData();
  }, []);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">読み込み中...</div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-2">
            <LayoutDashboard className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">ダッシュボード</h1>
          </div>
        </div>

        <Button onClick={loadDashboardData} variant="secondary">
          <RefreshCw className="h-4 w-4 mr-2" />
          更新
        </Button>
      </div>

      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.1 }}
        >
          <Card className="shadow-sm">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">フォーム数</CardTitle>
              <FileSpreadsheet className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalForms}</div>
              <p className="text-xs text-muted-foreground">
                連携中のGoogleフォーム
              </p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.3 }}
        >
          <Card className="shadow-sm">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                未対応チケット
              </CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-amber-600">
                {stats.pendingTickets}
              </div>
              <p className="text-xs text-muted-foreground">
                新規・対応中のチケット
              </p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.4 }}
        >
          <Card className="shadow-sm">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">直近の回答</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">
                {stats.recentResponses}
              </div>
              <p className="text-xs text-muted-foreground">過去24時間</p>
            </CardContent>
          </Card>
        </motion.div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.5 }}
        >
          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileSpreadsheet className="h-5 w-5" />
                フォーム一覧
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {forms.length > 0 ? (
                  forms.slice(0, 5).map((form) => (
                    <div
                      key={form.form_id}
                      className="flex items-center justify-between p-3 rounded-lg border hover:bg-gray-50 transition-colors"
                    >
                      <div className="space-y-1 flex-1 min-w-0">
                        <div className="text-sm font-medium truncate">
                          {form.title}
                        </div>
                        <div className="text-xs text-muted-foreground font-mono">
                          {form.form_id}
                        </div>
                      </div>
                      <div className="flex gap-2 ml-2">
                        <Link to={`/kanban/${form.form_id}`}>
                          <Button variant="ghost" size="sm">
                            <BarChart3 className="h-4 w-4 mr-1" />
                            看板
                          </Button>
                        </Link>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <FileSpreadsheet className="h-8 w-8 mx-auto mb-2 opacity-50" />
                    <p>フォームがありません</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.6 }}
        >
          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                最近のチケット
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {recentTickets.length > 0 ? (
                  recentTickets.map((ticket) => (
                    <div
                      key={ticket.id}
                      className="flex items-center justify-between p-3 rounded-lg border"
                    >
                      <div className="space-y-1">
                        <div className="text-sm font-medium">
                          #{ticket.id.slice(-6)}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {ticket.form_id}
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge
                          className={
                            ticket.status === "new"
                              ? "bg-blue-100 text-blue-800"
                              : ticket.status === "in_progress"
                              ? "bg-amber-100 text-amber-800"
                              : "bg-green-100 text-green-800"
                          }
                        >
                          {ticket.status === "new"
                            ? "新規"
                            : ticket.status === "in_progress"
                            ? "対応中"
                            : "完了"}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {new Date(ticket.created_at).toLocaleDateString(
                            "ja-JP"
                          )}
                        </span>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <ListChecks className="h-8 w-8 mx-auto mb-2 opacity-50" />
                    <p>チケットがありません</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </div>
  );
}
