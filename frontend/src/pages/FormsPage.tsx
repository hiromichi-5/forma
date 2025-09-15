import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import {
  FileSpreadsheet,
  Search,
  Plus,
  RefreshCw,
  ExternalLink,
  Settings,
} from "lucide-react";
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
  const [syncing, setSyncing] = useState<Set<string>>(new Set());

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

  const handleSync = async (formId: string) => {
    try {
      setSyncing((prev) => new Set(prev).add(formId));
      await apiClient.syncForm(formId);
      // Note: Actual sync status would need to be checked via a separate endpoint
      // For now, just show success message
    } catch (err) {
      console.error("Failed to sync form:", err);
      setError(
        err instanceof ApiError ? err.message : "フォームの同期に失敗しました"
      );
    } finally {
      setSyncing((prev) => {
        const newSet = new Set(prev);
        newSet.delete(formId);
        return newSet;
      });
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
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <FileSpreadsheet className="h-8 w-8 text-blue-600" />
            フォーム管理
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Googleフォームの連携と同期管理
          </p>
        </div>
        <Button className="flex items-center gap-2">
          <Plus className="h-4 w-4" />
          新規フォーム連携
        </Button>
      </div>

      {/* Search Bar */}
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

      {/* Error State */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/50 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <div className="flex items-center gap-2 text-red-800 dark:text-red-200">
            <RefreshCw className="h-5 w-5" />
            {error}
          </div>
        </div>
      )}

      {/* Forms Grid */}
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
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  フォームID: {form.form_id}
                </p>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-2 pt-2">
                  <Button
                    size="sm"
                    onClick={() => handleSync(form.form_id)}
                    disabled={syncing.has(form.form_id)}
                    className="flex items-center gap-1 flex-1"
                  >
                    <RefreshCw
                      className={`h-3 w-3 ${
                        syncing.has(form.form_id) ? "animate-spin" : ""
                      }`}
                    />
                    {syncing.has(form.form_id) ? "同期中..." : "同期"}
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    className="flex items-center gap-1"
                  >
                    <ExternalLink className="h-3 w-3" />
                    開く
                  </Button>
                  <Button variant="ghost" size="sm" className="px-2">
                    <Settings className="h-3 w-3" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>

      {/* Empty State */}
      {!loading && filteredForms.length === 0 && (
        <div className="text-center py-12">
          <FileSpreadsheet className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900 dark:text-white">
            {searchQuery ? "検索結果が見つかりません" : "フォームがありません"}
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
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

      {/* Stats Summary */}
      {forms.length > 0 && (
        <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                {forms.length}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                総フォーム数
              </p>
            </div>
            <div>
              <p className="text-2xl font-bold text-green-600 dark:text-green-400">
                {filteredForms.length}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                表示中のフォーム
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
