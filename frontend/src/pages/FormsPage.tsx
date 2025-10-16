import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import {
  FileSpreadsheet,
  Search,
  Plus,
  RefreshCw,
  ExternalLink,
  LayoutDashboard,
} from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/Card";
import { Input } from "../components/ui/Input";
import { apiClient, ApiError } from "../lib/api";
import type { FormSummary } from "../types";

export default function FormsPage() {
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    fetchForms();
  }, []);

  const fetchForms = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getForms();
      setForms(data.forms);
    } catch (err) {
      console.error("Failed to fetch forms:", err);
      setError(
        err instanceof ApiError ? err.message : "フォームの取得に失敗しました"
      );
    } finally {
      setLoading(false);
    }
  };

  const filteredForms = forms.filter((form) =>
    form.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
            <FileSpreadsheet className="h-8 w-8 text-blue-600" />
            フォーム管理
          </h1>
          <p className="text-gray-600 mt-2">Googleフォームの連携と同期管理</p>
        </div>
        <Button className="flex items-center gap-2">
          <Plus className="h-4 w-4" />
          新規フォーム連携
        </Button>
      </div>

      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
        <Input
          type="text"
          placeholder="フォームを検索..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10"
        />
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center gap-2 text-red-800">
            <RefreshCw className="h-5 w-5" />
            {error}
          </div>
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {filteredForms.map((form, index) => (
          <motion.div
            key={form.form_id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: index * 0.1 }}
          >
            <Card className="h-full hover:shadow-lg transition-shadow">
              <CardHeader>
                <CardTitle className="text-lg line-clamp-2">
                  {form.title}
                </CardTitle>
                <p className="text-sm text-gray-600">
                  フォームID: {form.form_id}
                </p>
              </CardHeader>
              <CardContent>
                <div className="flex justify-end gap-2 pt-2">
                  <Link to={`/kanban/${form.form_id}`} className="inline-flex">
                    <Button
                      variant="secondary"
                      size="sm"
                      className="flex items-center gap-1"
                    >
                      <LayoutDashboard className="h-3 w-3" />
                      看板
                    </Button>
                  </Link>
                  <a
                    href={`https://docs.google.com/forms/d/${form.form_id}/viewform`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center justify-center gap-1 rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:border-gray-400 hover:bg-gray-50 active:bg-gray-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
                  >
                    <ExternalLink className="h-3 w-3" />
                    開く
                  </a>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>

      {!loading && filteredForms.length === 0 && (
        <div className="text-center py-12">
          <FileSpreadsheet className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900">
            {searchQuery ? "検索結果が見つかりません" : "フォームがありません"}
          </h3>
          <p className="mt-1 text-sm text-gray-500">
            {searchQuery
              ? "検索条件を変更してもう一度お試しください"
              : "最初のGoogleフォームを連携して開始しましょう"}
          </p>
          {!searchQuery && (
            <div className="mt-6">
              <Button className="flex items-center gap-2">
                <Plus className="h-4 w-4" />
                新規フォーム連携
              </Button>
            </div>
          )}
        </div>
      )}

      {forms.length > 0 && (
        <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold text-blue-600">{forms.length}</p>
              <p className="text-sm text-gray-600">総フォーム数</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-green-600">
                {filteredForms.length}
              </p>
              <p className="text-sm text-gray-600">表示中のフォーム</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
