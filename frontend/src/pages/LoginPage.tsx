import React, { useState } from "react";
import { Navigate, Link, useLocation } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Card, CardContent, CardTitle } from "../components/ui/card";
import { ApiError, apiClient } from "../lib/api";
import { Loader2 } from "lucide-react";

export default function LoginPage() {
  const { user, login } = useAuth();
  const location = useLocation();
  const from = (location.state as { from?: string })?.from ?? "/";
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [showResendVerification, setShowResendVerification] = useState(false);
  const [resendStatus, setResendStatus] = useState<"idle" | "sending" | "sent">("idle");

  if (user) {
    return <Navigate to={from} replace />;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setShowResendVerification(false);
    setIsLoading(true);

    try {
      await login({ email, password });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.error.code === "EMAIL_NOT_VERIFIED") {
          setError("メールアドレスが未認証です。確認メールをご確認ください。");
          setShowResendVerification(true);
        } else if (err.isUnauthorized) {
          setError("メールアドレスまたはパスワードが間違っています");
        } else if (err.isValidationError) {
          setError("入力内容を確認してください");
        } else {
          setError("ログインに失敗しました");
        }
      } else {
        setError("予期しないエラーが発生しました");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendVerification = async () => {
    setResendStatus("sending");
    try {
      await apiClient.resendVerification({ email });
      setResendStatus("sent");
    } catch {
      setResendStatus("idle");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-6">
          <div className="space-y-6">
            <div className="text-center">
              <div className="flex justify-center mb-4">
                <div className="rounded-full bg-primary p-3">
                  <img src="/favicon.svg" alt="Logo" className="h-20 w-20" />
                </div>
              </div>
              <CardTitle className="text-2xl font-bold">Forma</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                チームでGoogleフォームを取り込んで、
                <br />
                回答をチケット化して管理
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

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="password">パスワード</Label>
                  <Link
                    to="/password-reset"
                    className="text-xs text-primary hover:underline"
                  >
                    パスワードを忘れた方
                  </Link>
                </div>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={isLoading}
                />
              </div>

              {error && (
                <div
                  className="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
                  role="alert"
                  aria-live="polite"
                >
                  {error}
                  {showResendVerification && (
                    <div className="mt-2">
                      {resendStatus === "sent" ? (
                        <p className="text-muted-foreground">確認メールを再送しました</p>
                      ) : (
                        <Button
                          type="button"
                          variant="link"
                          size="sm"
                          className="h-auto p-0 text-primary"
                          disabled={resendStatus === "sending"}
                          onClick={handleResendVerification}
                        >
                          確認メールを再送する
                        </Button>
                      )}
                    </div>
                  )}
                </div>
              )}

              <Button
                type="submit"
                className="w-full"
                disabled={isLoading || !email || !password}
              >
                {isLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ログイン中...
                  </>
                ) : (
                  "ログイン"
                )}
              </Button>
            </form>

            <div className="text-center">
              <p className="text-sm text-muted-foreground">
                アカウントをお持ちでない方は{" "}
                <Link
                  to="/signup"
                  className="font-medium text-primary hover:underline"
                >
                  サインアップ
                </Link>
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
