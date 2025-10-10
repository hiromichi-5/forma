import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import { Input } from "../components/ui/Input";
import { apiClient, ApiError } from "../lib/api";
import { RefreshCw, CheckCircle2, AlertTriangle } from "lucide-react";

export default function InviteAcceptPage() {
  const [code, setCode] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!code.trim()) {
      setError("招待コードを入力してください");
      return;
    }
    try {
      setSubmitting(true);
      setError(null);
      await apiClient.acceptInvite({ code: code.trim() });
      setSuccess(true);
    } catch (err) {
      console.error("Failed to accept invite:", err);
      if (err instanceof ApiError) {
        switch (err.status) {
          case 404:
            setError("招待コードが見つかりませんでした");
            break;
          case 409:
            setError("すでにメンバーとして参加済みです");
            break;
          case 410:
            setError("招待コードの有効期限が切れています");
            break;
          case 400:
            setError("招待コードが不正です");
            break;
          default:
            setError(err.message || "招待の受理に失敗しました");
        }
      } else {
        setError("招待の受理に失敗しました");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleGoMembers = () => {
    navigate("/members");
  };

  return (
    <div className="min-h-[70vh] flex items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-xl font-semibold text-center">
            招待コードの受理
          </CardTitle>
        </CardHeader>
        <CardContent>
          {success ? (
            <div className="text-center space-y-4">
              <CheckCircle2 className="mx-auto h-12 w-12 text-green-500" />
              <p className="text-gray-700">
                招待コードの受理が完了しました。メンバー一覧から参加状況を確認できます。
              </p>
              <Button className="w-full" onClick={handleGoMembers}>
                メンバー画面へ移動
              </Button>
            </div>
          ) : (
            <form className="space-y-6" onSubmit={handleSubmit}>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-700" htmlFor="invite-code">
                  招待コード
                </label>
                <Input
                  id="invite-code"
                  placeholder="例: abcd1234"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  autoComplete="off"
                />
              </div>

              {error && (
                <div className="flex items-center gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                  <AlertTriangle className="h-4 w-4" />
                  {error}
                </div>
              )}

              <Button type="submit" className="w-full" disabled={submitting}>
                {submitting ? (
                  <span className="flex items-center justify-center gap-2">
                    <RefreshCw className="h-4 w-4 animate-spin" />
                    送信中...
                  </span>
                ) : (
                  "参加する"
                )}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
