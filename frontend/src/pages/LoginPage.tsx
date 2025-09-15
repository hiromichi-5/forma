import React, { useState } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";
import { Label } from "../components/ui/Label";
import { Card, CardContent, CardTitle } from "../components/ui/Card";
import { ApiError } from "../lib/api";
import { ShieldCheck, Loader2 } from "lucide-react";

export function LoginPage() {
  const { user, login } = useAuth();
  const [email, setEmail] = useState("admin@example.com");
  const [password, setPassword] = useState("password");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  if (user) {
    return <Navigate to="/" replace />;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await login({ email, password });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.isUnauthorized) {
          setError("メールアドレスまたはパスワードが間違っています");
        } else if (err.isValidationError) {
          setError("入力内容を確認してください");
        } else {
          setError(
            "ログインに失敗しました。しばらく時間をおいて再度お試しください"
          );
        }
      } else {
        setError("予期しないエラーが発生しました");
      }
    } finally {
      setIsLoading(false);
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
                  <ShieldCheck className="h-8 w-8 text-primary-foreground" />
                </div>
              </div>
              <CardTitle className="text-2xl font-bold">Forma Admin</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                チームでGoogleフォームを取り込み、回答をチケット化して管理
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
                <Label htmlFor="password">パスワード</Label>
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
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
