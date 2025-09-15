import { useState } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/Button";
import { Input, Field } from "@/components/ui/Form";
import { ApiError } from "@/lib/api";

export function LoginPage() {
  const { user, login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
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
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h1 className="mt-6 text-center text-3xl font-bold text-gray-900">
            Forma
          </h1>
          <p className="mt-2 text-center text-sm text-gray-600">ログイン</p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            <Field
              label="メールアドレス"
              id="email"
              required
              error={error && email === "" ? "必須項目です" : undefined}
            >
              <Input
                id="email"
                type="email"
                placeholder="user@example.com"
                value={email}
                onChange={setEmail}
                required
                disabled={isLoading}
                aria-describedby={error ? "email-error" : undefined}
              />
            </Field>

            <Field
              label="パスワード"
              id="password"
              required
              error={error && password === "" ? "必須項目です" : undefined}
            >
              <Input
                id="password"
                type="password"
                placeholder="パスワードを入力"
                value={password}
                onChange={setPassword}
                onBlur={() => {}}
                required
                disabled={isLoading}
                aria-describedby={error ? "password-error" : undefined}
              />
            </Field>
          </div>

          {error && (
            <div
              className="rounded-md bg-red-50 p-4"
              role="alert"
              aria-live="polite"
            >
              <div className="text-sm text-red-700">{error}</div>
            </div>
          )}

          <div>
            <Button
              type="submit"
              variant="primary"
              size="lg"
              disabled={isLoading || !email || !password}
              loading={isLoading}
              className="w-full"
              aria-label={isLoading ? "ログイン中..." : "ログイン"}
            >
              {isLoading ? "ログイン中..." : "ログイン"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
