import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { Layout } from "@/components/Layout";
import { Button } from "@/components/ui/Button";
import { Loader, EmptyState, Toast } from "@/components/ui/Common";
import { Card, CardContent, CardTitle } from "@/components/ui/Card";
import { Icon } from "@/components/ui/Icon";
import { apiClient, ApiError } from "@/lib/api";
import type { FormSummary, ToastMessage } from "@/types";
import { Users, RefreshCw, BarChart3 } from "lucide-react";

export function HomePage() {
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [syncingForms, setSyncingForms] = useState<Set<string>>(new Set());
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const addToast = (toast: Omit<ToastMessage, "id">) => {
    const id = Math.random().toString(36).substr(2, 9);
    setToasts((prev) => [...prev, { ...toast, id }]);
  };

  const removeToast = (id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  };

  const loadForms = async () => {
    try {
      setIsLoading(true);
      const response = await apiClient.getForms();
      setForms(response.forms);
    } catch (error) {
      if (error instanceof ApiError) {
        addToast({
          type: "error",
          title: "エラー",
          message: "フォーム一覧の取得に失敗しました",
        });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleSync = async (formId: string) => {
    try {
      setSyncingForms((prev) => new Set(prev).add(formId));
      const response = await apiClient.syncForm(formId);

      addToast({
        type: "success",
        title: "同期完了",
        message: `${response.synced}件の回答を同期し、${response.newTickets}件の新しいチケットを作成しました`,
      });
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.isForbidden) {
          addToast({
            type: "error",
            title: "アクセス権限がありません",
            message: "このフォームの同期権限がありません",
          });
        } else {
          addToast({
            type: "error",
            title: "同期エラー",
            message: "フォームの同期に失敗しました",
          });
        }
      }
    } finally {
      setSyncingForms((prev) => {
        const newSet = new Set(prev);
        newSet.delete(formId);
        return newSet;
      });
    }
  };

  useEffect(() => {
    loadForms();
  }, []);

  if (isLoading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-96">
          <Loader size="lg" />
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">フォーム一覧</h1>
            <p className="mt-1 text-sm text-gray-600">
              管理しているGoogleフォームの一覧です
            </p>
          </div>
          <Button
            onClick={loadForms}
            variant="secondary"
            aria-label="フォーム一覧を更新"
          >
            <Icon icon={RefreshCw} size="sm" className="mr-2" />
            更新
          </Button>
        </div>

        {forms.length === 0 ? (
          <EmptyState
            title="フォームがありません"
            description="まだフォームが登録されていません。管理者に連絡してフォームを追加してもらってください。"
          />
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {forms.map((form) => (
              <FormCard
                key={form.form_id}
                form={form}
                onSync={handleSync}
                isSyncing={syncingForms.has(form.form_id)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Toast messages */}
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          type={toast.type}
          title={toast.title}
          message={toast.message}
          onClose={() => removeToast(toast.id)}
          duration={toast.duration}
        />
      ))}
    </Layout>
  );
}

interface FormCardProps {
  form: FormSummary;
  onSync: (formId: string) => void;
  isSyncing: boolean;
}

function FormCard({ form, onSync, isSyncing }: FormCardProps) {
  return (
    <Card className="overflow-hidden hover:shadow-md transition-shadow">
      <CardContent>
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <CardTitle className="truncate">{form.title}</CardTitle>
            <p className="mt-1 text-sm text-gray-500 font-mono">
              ID: {form.form_id}
            </p>
          </div>
        </div>

        <div className="mt-6 flex flex-col gap-3">
          <div className="flex gap-2">
            <Link to={`/forms/${form.form_id}/kanban`} className="flex-1">
              <Button
                variant="primary"
                className="w-full"
                aria-label={`${form.title}の看板を表示`}
              >
                <Icon icon={BarChart3} size="sm" className="mr-2" />
                看板へ
              </Button>
            </Link>
          </div>

          <div className="flex gap-2">
            <Button
              variant="secondary"
              size="sm"
              className="flex-1"
              aria-label={`${form.title}のメンバー管理`}
            >
              <Icon icon={Users} size="sm" className="mr-2" />
              メンバー
            </Button>

            <Button
              variant="secondary"
              size="sm"
              onClick={() => onSync(form.form_id)}
              disabled={isSyncing}
              loading={isSyncing}
              aria-label={`${form.title}を同期`}
            >
              <Icon icon={RefreshCw} size="sm" className="mr-2" />
              同期
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
