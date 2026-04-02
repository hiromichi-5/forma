import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Card, CardContent, CardTitle } from "../components/ui/card";
import { ApiError, apiClient } from "../lib/api";
import { useRequireAuth } from "../hooks/useAuth";
import { CheckCircle, Loader2, XCircle } from "lucide-react";

export default function AcceptInvitePage() {
  const { inviteId } = useParams<{ inviteId: string }>();
  const { isLoading: authLoading, user } = useRequireAuth();
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    if (authLoading || !user || !inviteId) return;
    if (status !== "idle") return;

    const accept = async () => {
      setStatus("loading");
      try {
        await apiClient.acceptInvite(inviteId);
        setStatus("success");
      } catch (err) {
        setStatus("error");
        if (err instanceof ApiError) {
          switch (err.error.code) {
            case "INVITE_NOT_FOUND":
              setErrorMessage("この招待は存在しないか、既に使用されています");
              break;
            case "INVITE_EXPIRED":
              setErrorMessage("この招待は期限切れです");
              break;
            case "RESOURCE_HIDDEN":
              setErrorMessage("この招待はあなたのメールアドレス宛ではありません");
              break;
            case "ALREADY_MEMBER":
              setErrorMessage("既にこのフォームのメンバーです");
              break;
            default:
              setErrorMessage("招待の受諾に失敗しました");
          }
        } else {
          setErrorMessage("予期しないエラーが発生しました");
        }
      }
    };

    accept();
  }, [authLoading, user, inviteId, status]);

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-6">
          <div className="space-y-6 text-center">
            {status === "loading" || status === "idle" ? (
              <>
                <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
                <CardTitle className="text-xl">招待を受諾中...</CardTitle>
              </>
            ) : status === "success" ? (
              <>
                <div className="flex justify-center">
                  <div className="rounded-full bg-green-100 p-4">
                    <CheckCircle className="h-12 w-12 text-green-600" />
                  </div>
                </div>
                <div>
                  <CardTitle className="text-xl">招待を受諾しました</CardTitle>
                  <p className="mt-2 text-sm text-muted-foreground">
                    フォームにアクセスできるようになりました。
                  </p>
                </div>
                <Link
                  to="/"
                  className="inline-block font-medium text-primary hover:underline"
                >
                  フォーム一覧へ
                </Link>
              </>
            ) : (
              <>
                <div className="flex justify-center">
                  <div className="rounded-full bg-destructive/10 p-4">
                    <XCircle className="h-12 w-12 text-destructive" />
                  </div>
                </div>
                <div>
                  <CardTitle className="text-xl">受諾に失敗しました</CardTitle>
                  <p className="mt-2 text-sm text-muted-foreground">{errorMessage}</p>
                </div>
                <Link
                  to="/"
                  className="inline-block font-medium text-primary hover:underline"
                >
                  ホームへ戻る
                </Link>
              </>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
