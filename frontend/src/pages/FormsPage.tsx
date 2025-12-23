import { useState, useEffect, type FormEvent } from "react";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../components/ui/Dialog";
import { Label } from "../components/ui/label";
import { apiClient, ApiError } from "../lib/api";
import type { FormSummary } from "../types";

export default function FormsPage() {
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newFormUrl, setNewFormUrl] = useState("");
  const [newFormUrlError, setNewFormUrlError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    fetchForms();
  }, []);

  const fetchForms = async (withSpinner: boolean = true) => {
    try {
      if (withSpinner) {
        setLoading(true);
      }
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

  const resetDialogState = () => {
    setNewFormUrl("");
    setNewFormUrlError(null);
    setSubmitError(null);
    setIsSubmitting(false);
  };

  const handleDialogOpenChange = (open: boolean) => {
    setIsDialogOpen(open);
    if (!open) {
      resetDialogState();
    }
  };

  const handleRegisterForm = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedUrl = newFormUrl.trim();
    if (!trimmedUrl) {
      setNewFormUrlError("フォームのURLまたはIDを入力してください");
      return;
    }

    setNewFormUrlError(null);
    setSubmitError(null);
    setIsSubmitting(true);

    try {
      await apiClient.registerForm({ url: trimmedUrl });
      await fetchForms(false);
      handleDialogOpenChange(false);
    } catch (err) {
      console.error("Failed to register form:", err);
      setSubmitError(
        err instanceof ApiError
          ? err.message
          : "フォームの登録に失敗しました。しばらくしてから再度お試しください。"
      );
      setIsSubmitting(false);
    }
  };

  const handleUrlChange = (value: string) => {
    setNewFormUrl(value);
    if (newFormUrlError) {
      setNewFormUrlError(null);
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
        <Dialog open={isDialogOpen} onOpenChange={handleDialogOpenChange}>
          <DialogTrigger asChild>
            <Button className="flex items-center gap-2">
              <Plus className="h-4 w-4" />
              新規フォーム連携
            </Button>
          </DialogTrigger>
          <DialogContent className="bg-white">
            <form onSubmit={handleRegisterForm} className="space-y-6">
              <DialogHeader>
                <DialogTitle>新規フォームを登録</DialogTitle>
                <DialogDescription>
                  GoogleフォームのURLまたはフォームIDを入力してください。
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-2">
                <Label htmlFor="new-form-url" required>
                  フォームURLまたはID
                </Label>
                <Input
                  id="new-form-url"
                  type="text"
                  placeholder="https://docs.google.com/forms/d/..."
                  value={newFormUrl}
                  onChange={(event) => handleUrlChange(event.target.value)}
                  aria-describedby={
                    newFormUrlError ? "new-form-url-error" : undefined
                  }
                  aria-invalid={!!newFormUrlError}
                  disabled={isSubmitting}
                />
                {newFormUrlError && (
                  <p
                    id="new-form-url-error"
                    className="text-sm text-red-600"
                    role="alert"
                  >
                    {newFormUrlError}
                  </p>
                )}
              </div>

              {submitError && (
                <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                  {submitError}
                </div>
              )}

              <DialogFooter>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => handleDialogOpenChange(false)}
                  disabled={isSubmitting}
                >
                  キャンセル
                </Button>
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "登録中..." : "登録する"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
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

      {forms.length > 0 && (
        <div className="bg-gray-50 rounded-lg p-6">
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
                <p className="text-sm text-gray-600 break-words">
                  フォームID: {form.form_id}
                </p>
              </CardHeader>
              <CardContent>
                <div className="flex justify-end gap-2 pt-2">
                  <a
                    href={`https://docs.google.com/forms/d/${form.form_id}/viewform`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center justify-center gap-1 rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:border-gray-400 hover:bg-gray-50 active:bg-gray-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
                  >
                    <ExternalLink className="h-3 w-3" />
                    Googleフォーム
                  </a>
                  <Link
                    to={`/forms/${form.form_id}/dashboard`}
                    className="inline-flex items-center justify-center gap-1 rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:border-gray-400 hover:bg-gray-50 active:bg-gray-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
                  >
                    <LayoutDashboard className="h-3 w-3" />
                    看板
                  </Link>
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
              <Button
                className="flex items-center gap-2"
                onClick={() => handleDialogOpenChange(true)}
              >
                <Plus className="h-4 w-4" />
                新規フォーム連携
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
