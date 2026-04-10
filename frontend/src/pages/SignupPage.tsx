import { useState } from "react";
import { Link, Navigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Card, CardContent, CardTitle } from "../components/ui/card";
import { getApiErrorMessage } from "../lib/api-error";
import { Loader2, MailCheck } from "lucide-react";

export default function SignupPage() {
  const { user } = useAuth();
  const { signup } = useAuth();
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSignupComplete, setIsSignupComplete] = useState(false);

  if (user) {
    return <Navigate to="/" replace />;
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await signup({ email, password, display_name: displayName });
      setIsSignupComplete(true);
    } catch (err) {
      setError(
        getApiErrorMessage(
          err,
          {
            CONFLICT: "このメールアドレスは既に登録されています",
            VALIDATION_ERROR: "入力内容を確認してください（パスワードは8文字以上）",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "サインアップに失敗しました"
        )
      );
    } finally {
      setIsLoading(false);
    }
  };

  if (isSignupComplete) {
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
                <CardTitle className="text-2xl font-bold">確認メールを送信しました</CardTitle>
                <p className="mt-3 text-sm text-muted-foreground">
                  <span className="font-medium text-foreground">{email}</span>
                  {" "}に確認メールを送信しました。
                  <br />
                  メール内のリンクをクリックしてアカウントを有効化してください。
                </p>
              </div>
              <div className="pt-2">
                <Link
                  to="/login"
                  className="font-medium text-primary hover:underline text-sm"
                >
                  ログインページへ
                </Link>
              </div>
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
              <CardTitle className="text-2xl font-bold">Forma</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                新しいアカウントを作成して
                <br />
                フォーム管理を始めましょう
              </p>
            </div>

            <form className="space-y-4" onSubmit={handleSubmit}>
              <div className="space-y-2">
                <Label htmlFor="displayName">表示名</Label>
                <Input
                  id="displayName"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  required
                  disabled={isLoading}
                />
              </div>

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

              <div className="space-y-2">
                <Label htmlFor="password">パスワード</Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={8}
                  disabled={isLoading}
                />
                <p className="text-xs text-muted-foreground">8文字以上</p>
              </div>

              {error && (
                <div
                  className="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
                  role="alert"
                  aria-live="polite"
                >
                  {error}
                </div>
              )}

              <Button
                type="submit"
                className="w-full"
                disabled={isLoading || !email || !password || !displayName}
              >
                {isLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    作成中...
                  </>
                ) : (
                  "サインアップ"
                )}
              </Button>
            </form>

            <div className="text-center">
              <p className="text-sm text-muted-foreground">
                すでにアカウントをお持ちの方は{" "}
                <Link
                  to="/login"
                  className="font-medium text-primary hover:underline"
                >
                  ログイン
                </Link>
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
