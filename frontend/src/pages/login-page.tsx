import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import type { z } from "zod"
import { Navigate, Link, useLocation } from "react-router"
import { useAuth } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { TextField } from "@/components/text-field"
import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage, hasApiErrorCode } from "@/lib/api-error"
import { loginSchema } from "@/lib/form-schemas"
import { Loader2 } from "lucide-react"

export default function LoginPage() {
  const { user, login } = useAuth()
  const location = useLocation()
  const from = (location.state as { from?: string })?.from ?? "/"
  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors, isSubmitting: isLoading },
  } = useForm<z.infer<typeof loginSchema>>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  })
  const [error, setError] = useState("")
  const [showResendVerification, setShowResendVerification] = useState(false)
  const [resendStatus, setResendStatus] = useState<"idle" | "sending" | "sent">("idle")

  if (user) {
    return <Navigate to={from} replace />
  }

  const onSubmit = async (values: z.infer<typeof loginSchema>) => {
    setError("")
    setShowResendVerification(false)

    try {
      await login(values)
    } catch (err) {
      if (hasApiErrorCode(err, "EMAIL_NOT_VERIFIED")) {
        setError("メールアドレスが未認証です。確認メールをご確認ください。")
        setShowResendVerification(true)
      } else {
        setError(
          getApiErrorMessage(
            err,
            {
              INVALID_CREDENTIALS: "メールアドレスまたはパスワードが間違っています",
              RATE_LIMITED:
                "ログインの試行回数が上限に達しました。しばらくしてから再度お試しください",
              VALIDATION_ERROR: "入力内容を確認してください",
            },
            "ログインに失敗しました",
          ),
        )
      }
    }
  }

  const handleResendVerification = async () => {
    setResendStatus("sending")
    try {
      await apiClient.resendVerification({ email: getValues("email") })
      setResendStatus("sent")
    } catch {
      setResendStatus("idle")
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-6">
          <div className="space-y-6">
            <div className="text-center">
              <div className="mb-4 flex justify-center">
                <div className="rounded-full bg-primary p-3">
                  <img src="/favicon.svg" alt="Logo" className="h-20 w-20" />
                </div>
              </div>
              <CardTitle className="text-2xl font-bold">forma</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                チームでGoogleフォームを取り込んで、
                <br />
                回答をチケット化して管理
              </p>
            </div>

            <form className="space-y-4" noValidate onSubmit={handleSubmit(onSubmit)}>
              <TextField
                id="email"
                label="メールアドレス"
                type="email"
                autoComplete="email"
                disabled={isLoading}
                error={errors.email}
                {...register("email")}
              />

              <TextField
                id="password"
                label="パスワード"
                type="password"
                autoComplete="current-password"
                disabled={isLoading}
                error={errors.password}
                labelAction={
                  <Link to="/password-reset" className="text-xs text-primary hover:underline">
                    パスワードを忘れた方
                  </Link>
                }
                {...register("password")}
              />

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

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
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
                <Link to="/signup" className="font-medium text-primary hover:underline">
                  サインアップ
                </Link>
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
