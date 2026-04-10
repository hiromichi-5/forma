import { useState } from "react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Card, CardContent, CardTitle } from "../components/ui/card";
import { apiClient } from "../lib/api";
import { getApiErrorMessage } from "../lib/api-error";
import { Loader2, MailCheck } from "lucide-react";

export default function PasswordResetPage() {
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSent, setIsSent] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await apiClient.passwordReset({ email });
      setIsSent(true);
    } catch (err) {
      setError(
        getApiErrorMessage(
          err,
          {
            VALIDATION_ERROR: "有効なメールアドレスを入力してください",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "リセットメールの送信に失敗しました"
        )
      );
    } finally {
      setIsLoading(false);
    }
  };

  if (isSent) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
        <Card className="w-full max-w-md shadow-lg">
          <CardContent className="p-6">
            <div className="space-y-6 text-center">
              <div className="flex justify-center">
                <div className="rounded-full bg-primary/10 p-4">
                  <MailCheck className="h-12 w-12 text-primary" />
                </div>
              </div>
              <div>
                <CardTitle className="text-xl">メールを送信しました</CardTitle>
                <p className="mt-3 text-sm text-muted-foreground">
                  アカウントが存在する場合、パスワードリセット用のメールを送信しました。
                  メールをご確認ください。
                </p>
              </div>
              <Link
                to="/login"
                className="inline-block font-medium text-primary hover:underline text-sm"
              >
                ログインページへ戻る
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-6">
          <div className="space-y-6">
            <div className="text-center">
              <CardTitle className="text-2xl font-bold">パスワードリセット</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                登録したメールアドレスを入力してください。
                <br />
                パスワードリセット用のリンクを送信します。
              </p>
            </div>

            <form className="space-y-4" onSubmit={handleSubmit}>
              <div className="space-y-2">
                <Label htmlFor="email">メールアドレス</Label>
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={isLoading}
                />
              </div>

              {error && (
                <div
                  className="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
                  role="alert"
                >
                  {error}
                </div>
              )}

              <Button
                type="submit"
                className="w-full"
                disabled={isLoading || !email}
              >
                {isLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    送信中...
                  </>
                ) : (
                  "リセットメールを送信"
                )}
              </Button>
            </form>

            <div className="text-center">
              <Link
                to="/login"
                className="text-sm text-muted-foreground hover:underline"
              >
                ログインページへ戻る
              </Link>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
