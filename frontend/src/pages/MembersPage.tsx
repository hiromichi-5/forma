import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import {
  Users,
  Search,
  Plus,
  Mail,
  Shield,
  Settings,
  RefreshCw,
  UserCheck,
  FileSpreadsheet,
} from "lucide-react";
import { Button } from "../components/ui/Button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/Card";
import { Badge } from "../components/ui/Badge";
import { Input } from "../components/ui/Input";
import { Select } from "../components/ui/Select";
import { apiClient, ApiError } from "../lib/api";
import type { Member, FormSummary } from "../types";

export default function MembersPage() {
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [selectedFormId, setSelectedFormId] = useState<string>("");
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    fetchForms();
  }, []);

  useEffect(() => {
    if (selectedFormId) {
      fetchMembers(selectedFormId);
    }
  }, [selectedFormId]);

  const fetchForms = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getForms();
      setForms(data.forms);
      if (data.forms.length > 0 && !selectedFormId) {
        setSelectedFormId(data.forms[0].form_id);
      }
    } catch (err) {
      console.error("Failed to fetch forms:", err);
      setError(
        err instanceof ApiError ? err.message : "フォームの取得に失敗しました"
      );
    } finally {
      setLoading(false);
    }
  };

  const fetchMembers = async (formId: string) => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getMembers(formId);
      setMembers(data.members);
    } catch (err) {
      console.error("Failed to fetch members:", err);
      setError(
        err instanceof ApiError ? err.message : "メンバーの取得に失敗しました"
      );
    } finally {
      setLoading(false);
    }
  };

  const getRoleBadge = (role: "admin" | "editor") => {
    switch (role) {
      case "admin":
        return (
          <Badge
            variant="destructive"
            className="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300"
          >
            <Shield className="w-3 h-3 mr-1" />
            管理者
          </Badge>
        );
      case "editor":
        return (
          <Badge
            variant="default"
            className="bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300"
          >
            <UserCheck className="w-3 h-3 mr-1" />
            編集者
          </Badge>
        );
      default:
        return <Badge variant="secondary">{role}</Badge>;
    }
  };

  const filteredMembers = members.filter((member) =>
    member.email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const selectedForm = forms.find((f) => f.form_id === selectedFormId);

  if (loading && forms.length === 0) {
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
            <Users className="h-8 w-8 text-blue-600" />
            メンバー管理
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            フォームごとのメンバー管理とアクセス権限の設定
          </p>
        </div>
        {selectedFormId && (
          <Button className="flex items-center gap-2">
            <Plus className="h-4 w-4" />
            メンバーを招待
          </Button>
        )}
      </div>

      {/* Form Selection */}
      {forms.length > 0 && (
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <FileSpreadsheet className="h-5 w-5 text-gray-500" />
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              フォーム:
            </span>
          </div>
          <div className="w-64">
            <Select value={selectedFormId} onValueChange={setSelectedFormId}>
              {forms.map((form) => (
                <option key={form.form_id} value={form.form_id}>
                  {form.title}
                </option>
              ))}
            </Select>
          </div>
        </div>
      )}

      {/* Search Bar */}
      {selectedFormId && (
        <div className="relative max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
          <Input
            type="text"
            placeholder="メンバーを検索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/50 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <div className="flex items-center gap-2 text-red-800 dark:text-red-200">
            <RefreshCw className="h-5 w-5" />
            {error}
          </div>
        </div>
      )}

      {/* Members Grid */}
      {selectedFormId && (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredMembers.map((member, index) => (
            <motion.div
              key={member.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3, delay: index * 0.1 }}
            >
              <Card className="h-full hover:shadow-lg transition-shadow">
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <CardTitle className="text-lg">
                        {member.email.split("@")[0]}
                      </CardTitle>
                      <div className="flex items-center gap-1 text-sm text-gray-600 dark:text-gray-400 mt-1">
                        <Mail className="h-3 w-3" />
                        {member.email}
                      </div>
                    </div>
                    {getRoleBadge(member.role)}
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="space-y-4">
                    {/* Member ID */}
                    <div className="text-sm">
                      <span className="text-gray-500 dark:text-gray-400">
                        メンバーID:{" "}
                      </span>
                      <span className="font-mono text-xs">{member.id}</span>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-2 pt-2">
                      <Button
                        variant="secondary"
                        size="sm"
                        className="flex items-center gap-1 flex-1"
                      >
                        <Mail className="h-3 w-3" />
                        メッセージ
                      </Button>
                      <Button variant="ghost" size="sm" className="px-2">
                        <Settings className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      )}

      {/* Empty State for no form selected */}
      {!selectedFormId && forms.length > 0 && (
        <div className="text-center py-12">
          <FileSpreadsheet className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900 dark:text-white">
            フォームを選択してください
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            上のドロップダウンからフォームを選択してメンバーを表示
          </p>
        </div>
      )}

      {/* Empty State for no members */}
      {selectedFormId && !loading && filteredMembers.length === 0 && (
        <div className="text-center py-12">
          <Users className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900 dark:text-white">
            {searchQuery ? "検索結果が見つかりません" : "メンバーがいません"}
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {searchQuery
              ? "検索条件を変更してもう一度お試しください"
              : selectedForm
              ? `「${selectedForm.title}」にメンバーを招待してください`
              : "このフォームにはまだメンバーがいません"}
          </p>
          {!searchQuery && (
            <div className="mt-6">
              <Button className="flex items-center gap-2">
                <Plus className="h-4 w-4" />
                メンバーを招待
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Stats Summary */}
      {selectedFormId && members.length > 0 && (
        <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                {members.length}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                総メンバー数
              </p>
            </div>
            <div>
              <p className="text-2xl font-bold text-red-600 dark:text-red-400">
                {members.filter((m) => m.role === "admin").length}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">管理者</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                {members.filter((m) => m.role === "editor").length}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">編集者</p>
            </div>
          </div>
        </div>
      )}

      {/* No Forms State */}
      {forms.length === 0 && !loading && (
        <div className="text-center py-12">
          <FileSpreadsheet className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900 dark:text-white">
            フォームがありません
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            メンバーを管理するには、まずフォームを作成してください
          </p>
        </div>
      )}
    </div>
  );
}
