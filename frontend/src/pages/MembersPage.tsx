import { useState, useEffect, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import {
  Users,
  Search,
  Plus,
  Mail,
  Shield,
  RefreshCw,
  UserCheck,
  FileSpreadsheet,
  ArrowLeft,
  Clipboard,
  ClipboardCheck,
  Trash2,
  Clock,
  Trash,
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../components/ui/Select";
import { Label } from "../components/ui/Label";
import { apiClient, ApiError } from "../lib/api";
import type { Member, FormSummary, FormInvite } from "../types";

export default function MembersPage() {
  const { form_id } = useParams<{ form_id: string }>();
  const navigate = useNavigate();
  const [forms, setForms] = useState<FormSummary[]>([]);
  const [selectedFormId, setSelectedFormId] = useState<string>("");
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [invites, setInvites] = useState<FormInvite[]>([]);
  const [inviteLoading, setInviteLoading] = useState(false);
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [issuing, setIssuing] = useState(false);
  const [copiedCode, setCopiedCode] = useState<string | null>(null);
  const [infoMessage, setInfoMessage] = useState<string | null>(null);
  const [memberRoleUpdating, setMemberRoleUpdating] = useState<
    Record<string, boolean>
  >({});
  const [memberRemovingId, setMemberRemovingId] = useState<string | null>(null);

  const fetchInvites = useCallback(async (formId: string) => {
    try {
      setInviteLoading(true);
      setInviteError(null);
      const data = await apiClient.listInvites(formId);
      setInvites(data.invites);
    } catch (err) {
      console.error("Failed to fetch invites:", err);
      setInviteError(
        err instanceof ApiError ? err.message : "招待の取得に失敗しました"
      );
    } finally {
      setInviteLoading(false);
    }
  }, []);

  const fetchForms = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getForms();
      setForms(data.forms);
      if (data.forms.length > 0) {
        setSelectedFormId((prev) => {
          if (prev) {
            return prev;
          }
          if (form_id) {
            return form_id;
          }
          return data.forms[0].form_id;
        });
      }
    } catch (err) {
      console.error("Failed to fetch forms:", err);
      setError(
        err instanceof ApiError ? err.message : "フォームの取得に失敗しました"
      );
    } finally {
      setLoading(false);
    }
  }, [form_id]);

  useEffect(() => {
    fetchForms();
  }, [fetchForms]);

  useEffect(() => {
    if (form_id) {
      setSelectedFormId(form_id);
    }
  }, [form_id]);

  useEffect(() => {
    if (selectedFormId) {
      setInfoMessage(null);
      setInvites([]);
      fetchMembers(selectedFormId);
      fetchInvites(selectedFormId);
    }
  }, [selectedFormId, fetchInvites]);

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
          <Badge variant="destructive" className="bg-red-100 text-red-800">
            <Shield className="w-3 h-3 mr-1" />
            管理者
          </Badge>
        );
      case "editor":
        return (
          <Badge variant="default" className="bg-blue-100 text-blue-800">
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

  const handleFormSelection = (value: string) => {
    if (value === "all") {
      navigate("/members");
    } else {
      navigate(`/members/${value}`);
    }
  };

  const copyToClipboard = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      setCopiedCode(code);
      setTimeout(() => {
        setCopiedCode((prev) => (prev === code ? null : prev));
      }, 2000);
    } catch (err) {
      console.error("Clipboard copy failed:", err);
      setInviteError("クリップボードへのコピーに失敗しました");
    }
  };

  const handleIssueInvite = async () => {
    if (!selectedFormId) return;
    try {
      setIssuing(true);
      setInviteError(null);
      const invite = await apiClient.createInvite(selectedFormId);
      await fetchInvites(selectedFormId);
      setInfoMessage(
        "新しい招待コードを発行しました。コピーして共有してください。"
      );
      await copyToClipboard(invite.code);
    } catch (err) {
      console.error("Failed to issue invite:", err);
      setInviteError(
        err instanceof ApiError ? err.message : "招待コードの発行に失敗しました"
      );
    } finally {
      setIssuing(false);
    }
  };

  const handleCopy = async (code: string) => {
    setInviteError(null);
    await copyToClipboard(code);
  };

  const handleRevoke = async (code: string) => {
    if (!selectedFormId) return;
    try {
      setInviteError(null);
      await apiClient.revokeInvite(selectedFormId, code);
      setInvites((prev) => prev.filter((inv) => inv.code !== code));
    } catch (err) {
      console.error("Failed to revoke invite:", err);
      setInviteError(
        err instanceof ApiError ? err.message : "招待コードの失効に失敗しました"
      );
    }
  };

  const handleChangeRole = async (member: Member, nextRole: Member["role"]) => {
    if (!selectedFormId || member.role === nextRole) {
      return;
    }
    setInfoMessage(null);
    setError(null);
    setMemberRoleUpdating((prev) => ({ ...prev, [member.id]: true }));
    const previousRole = member.role;
    setMembers((prev) =>
      prev.map((m) => (m.id === member.id ? { ...m, role: nextRole } : m))
    );
    try {
      await apiClient.changeMemberRole(selectedFormId, member.id, {
        role: nextRole,
      });
      setInfoMessage(
        `「${member.email}」の権限を「${
          nextRole === "admin" ? "管理者" : "編集者"
        }」に更新しました。`
      );
    } catch (err) {
      console.error("Failed to change member role:", err);
      setError(
        err instanceof ApiError
          ? err.message
          : "メンバーの権限変更に失敗しました"
      );
      setMembers((prev) =>
        prev.map((m) => (m.id === member.id ? { ...m, role: previousRole } : m))
      );
    } finally {
      setMemberRoleUpdating((prev) => ({ ...prev, [member.id]: false }));
    }
  };

  const handleRemoveMember = async (member: Member) => {
    if (!selectedFormId) return;
    const confirmed = window.confirm(
      `メンバー「${member.email}」を削除してもよろしいですか？`
    );
    if (!confirmed) return;

    setInfoMessage(null);
    setError(null);
    setMemberRemovingId(member.id);
    try {
      await apiClient.removeMember(selectedFormId, member.id);
      setMembers((prev) => prev.filter((m) => m.id !== member.id));
      setInfoMessage(`「${member.email}」をメンバーから削除しました。`);
    } catch (err) {
      console.error("Failed to remove member:", err);
      setError(
        err instanceof ApiError ? err.message : "メンバーの削除に失敗しました"
      );
    } finally {
      setMemberRemovingId(null);
    }
  };

  if (loading && forms.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {form_id && (
            <Button
              variant="ghost"
              onClick={() => window.history.back()}
              aria-label="戻る"
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>
          )}
          <div className="rounded-lg bg-primary/10 p-2">
            <Users className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">メンバー管理</h1>
            <p className="text-muted-foreground">
              フォームごとのメンバー管理とアクセス権限の設定
              {form_id && ` - ${selectedForm?.title || form_id}`}
            </p>
          </div>
        </div>
        {selectedFormId && (
          <Button
            className="flex items-center gap-2"
            onClick={handleIssueInvite}
            disabled={issuing}
          >
            <Plus className="h-4 w-4" />
            {issuing ? "発行中..." : "招待コードを発行"}
          </Button>
        )}
      </div>

      {forms.length > 0 && (
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <FileSpreadsheet className="h-5 w-5 text-gray-500" />
            <span className="text-sm font-medium text-gray-700">フォーム:</span>
          </div>
          <div className="w-64">
            <Select
              value={selectedFormId || "all"}
              onValueChange={handleFormSelection}
            >
              <SelectTrigger className="w-48">
                <SelectValue placeholder="フォームを選択" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">すべてのフォーム</SelectItem>
                {forms.map((form) => (
                  <SelectItem key={form.form_id} value={form.form_id}>
                    {form.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      )}

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

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center gap-2 text-red-800">
            <RefreshCw className="h-5 w-5" />
            {error}
          </div>
        </div>
      )}

      {inviteError && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center gap-2 text-red-800">
            <RefreshCw className="h-5 w-5" />
            {inviteError}
          </div>
        </div>
      )}

      {infoMessage && (
        <div className="bg-blue-50 border border-blue-200 text-blue-800 rounded-lg p-4">
          {infoMessage}
        </div>
      )}

      {selectedFormId && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">招待コード</h2>
            {inviteLoading && (
              <div className="flex items-center gap-2 text-sm text-gray-500">
                <RefreshCw className="h-4 w-4 animate-spin" />
                読み込み中...
              </div>
            )}
          </div>

          {invites.length === 0 && !inviteLoading ? (
            <div className="rounded-md border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500">
              招待コードがありません。上の「招待コードを発行」ボタンから新しいコードを作成してください。
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {invites.map((invite, index) => (
                <motion.div
                  key={invite.code}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3, delay: index * 0.05 }}
                >
                  <Card className="h-full border-blue-100">
                    <CardHeader className="pb-3">
                      <CardTitle className="flex items-center justify-between text-base">
                        <span className="font-mono text-sm">{invite.code}</span>
                        <Badge
                          variant="outline"
                          className="flex items-center gap-1 text-blue-700 border-blue-200"
                        >
                          <Clock className="h-3 w-3" />
                          期限 {formatRelativeTime(invite.expires_at)}
                        </Badge>
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3 text-sm">
                      <div className="flex justify-between text-gray-500">
                        <span>発行日</span>
                        <span>{formatDate(invite.created_at)}</span>
                      </div>
                      <div className="flex justify-between text-gray-500">
                        <span>ロール</span>
                        <span className="font-medium text-gray-700">
                          編集者
                        </span>
                      </div>
                      <div className="flex gap-2 pt-2">
                        <Button
                          variant="secondary"
                          className="flex-1 flex items-center gap-2"
                          onClick={() => handleCopy(invite.code)}
                        >
                          {copiedCode === invite.code ? (
                            <ClipboardCheck className="h-4 w-4" />
                          ) : (
                            <Clipboard className="h-4 w-4" />
                          )}
                          コピー
                        </Button>
                        <Button
                          variant="ghost"
                          className="text-red-600 hover:bg-red-50"
                          onClick={() => handleRevoke(invite.code)}
                          aria-label={`${invite.code} を失効させる`}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                </motion.div>
              ))}
            </div>
          )}
        </div>
      )}

      {selectedFormId && members.length > 0 && (
        <div className="rounded-lg p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold text-blue-600">
                {members.length}
              </p>
              <p className="text-sm text-gray-600">総メンバー数</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-red-600">
                {members.filter((m) => m.role === "admin").length}
              </p>
              <p className="text-sm text-gray-600">管理者</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-blue-600">
                {members.filter((m) => m.role === "editor").length}
              </p>
              <p className="text-sm text-gray-600">編集者</p>
            </div>
          </div>
        </div>
      )}
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
                      {getRoleBadge(member.role)}
                      <CardTitle className="text-lg">
                        {member.email.split("@")[0]}
                      </CardTitle>
                      <div className="flex items-center gap-1 text-sm text-gray-600 mt-1">
                        <Mail className="h-3 w-3" />
                        {member.email}
                      </div>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="space-y-4">
                    <div className="text-sm">
                      <span className="text-gray-500">メンバーID: </span>
                      <span className="font-mono text-xs">{member.id}</span>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor={`member-role-${member.id}`}>権限</Label>
                      <Select
                        value={member.role}
                        onValueChange={(value) =>
                          handleChangeRole(member, value as Member["role"])
                        }
                        disabled={!!memberRoleUpdating[member.id]}
                      >
                        <SelectTrigger
                          id={`member-role-${member.id}`}
                          className="w-full"
                        >
                          <SelectValue placeholder="権限を選択" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="admin">管理者</SelectItem>
                          <SelectItem value="editor">編集者</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Button
                        variant="danger"
                        className="w-full flex items-center gap-2"
                        onClick={() => handleRemoveMember(member)}
                        loading={memberRemovingId === member.id}
                      >
                        <Trash className="h-4 w-4 text-white" />
                        <span className="text-white">メンバーから削除</span>
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      )}

      {!selectedFormId && forms.length > 0 && (
        <div className="text-center py-12">
          <FileSpreadsheet className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900">
            フォームを選択してください
          </h3>
          <p className="mt-1 text-sm text-gray-500">
            上のドロップダウンからフォームを選択してメンバーを表示
          </p>
        </div>
      )}

      {selectedFormId && !loading && filteredMembers.length === 0 && (
        <div className="text-center py-12">
          <Users className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900">
            {searchQuery ? "検索結果が見つかりません" : "メンバーがいません"}
          </h3>
          <p className="mt-1 text-sm text-gray-500">
            {searchQuery
              ? "検索条件を変更してもう一度お試しください"
              : selectedForm
              ? `「${selectedForm.title}」にメンバーを招待してください`
              : "このフォームにはまだメンバーがいません"}
          </p>
          {!searchQuery && (
            <div className="mt-6">
              <Button
                className="flex items-center gap-2"
                onClick={handleIssueInvite}
              >
                <Plus className="h-4 w-4" />
                招待コードを発行
              </Button>
            </div>
          )}
        </div>
      )}

      {forms.length === 0 && !loading && (
        <div className="text-center py-12">
          <FileSpreadsheet className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900">
            フォームがありません
          </h3>
          <p className="mt-1 text-sm text-gray-500">
            メンバーを管理するには、まずフォームを作成してください
          </p>
        </div>
      )}
    </div>
  );
}

function formatDate(iso: string) {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return iso;
  }
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  const hh = String(date.getHours()).padStart(2, "0");
  const mm = String(date.getMinutes()).padStart(2, "0");
  return `${y}-${m}-${d} ${hh}:${mm}`;
}

function formatRelativeTime(iso: string) {
  const target = new Date(iso);
  const diffMs = target.getTime() - Date.now();
  if (Number.isNaN(diffMs)) {
    return iso;
  }
  if (diffMs <= 0) {
    return "期限切れ";
  }
  const minutes = Math.floor(diffMs / (1000 * 60));
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  if (days > 0) {
    return `${days}日${hours % 24}時間後`;
  }
  if (hours > 0) {
    return `${hours}時間${minutes % 60}分後`;
  }
  return `${Math.max(minutes, 1)}分後`;
}
