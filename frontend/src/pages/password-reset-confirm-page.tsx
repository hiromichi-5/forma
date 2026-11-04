import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import type { z } from "zod"
import { Link, useSearchParams } from "react-router"
import { Button } from "@/components/ui/button"
import { TextField } from "@/components/text-field"
import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import { PASSWORD_MIN_LENGTH, passwordResetConfirmSchema } from "@/lib/form-schemas"
import { CheckCircle, Loader2 } from "lucide-react"

export default function PasswordResetConfirmPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get("token")
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting: isLoading },
  } = useForm<z.infer<typeof passwordResetConfirmSchema>>({
    resolver: zodResolver(passwordResetConfirmSchema),
    defaultValues: { new_password: "", confirm_password: "" },
  })
  const [error, setError] = useState("")
  const [isComplete, setIsComplete] = useState(false)

  if (!token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted p-4">
        <Card className="w-full max-w-md shadow-lg">
          <CardContent className="space-y-4 p-6 text-center">
            <CardTitle className="text-xl">無効なリンク</CardTitle>
            <p className="text-sm text-muted-foreground">
              パスワードリセット用のトークンが見つかりません。
            </p>
            <Link
              to="/password-reset"
              className="inline-block text-sm font-medium text-primary hover:underline"
            >
              パスワードリセットを再度リクエスト
            </Link>
          </CardContent>
        </Card>
      </div>
    )
  }

  const onSubmit = async (values: z.infer<typeof passwordResetConfirmSchema>) => {
    setError("")

    try {
      await apiClient.passwordResetConfirm({ token, new_password: values.new_password })
      setIsComplete(true)
    } catch (err) {
      setError(
        getApiErrorMessage(
          err,
          {
            TOKEN_NOT_FOUND:
              "トークンが無効または期限切れです。再度リセットをリクエストしてください。",
            VALIDATION_ERROR: "パスワードは8文字以上で入力してください",
          },
          "パスワードリセットに失敗しました",
        ),
      )
    }
  }

  if (isComplete) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted p-4">
        <Card className="w-full max-w-md shadow-lg">
          <CardContent className="p-6">
            <div className="space-y-6 text-center">
              <div className="flex justify-center">
                <div className="rounded-full bg-green-100 p-4">
                  <CheckCircle className="h-12 w-12 text-green-600" />
                </div>
              </div>
              <div>
                <CardTitle className="text-xl">パスワードを変更しました</CardTitle>
                <p className="mt-2 text-sm text-muted-foreground">
                  新しいパスワードでログインしてください。
                </p>
              </div>
              <Link to="/login" className="inline-block font-medium text-primary hover:underline">
                ログインページへ
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-6">
          <div className="space-y-6">
            <div className="text-center">
              <CardTitle className="text-2xl font-bold">新しいパスワード設定</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                新しいパスワードを入力してください。
              </p>
            </div>

            <form className="space-y-4" noValidate onSubmit={handleSubmit(onSubmit)}>
              <TextField
                id="newPassword"
                label="新しいパスワード"
                type="password"
                autoComplete="new-password"
                description={`${PASSWORD_MIN_LENGTH}文字以上`}
                disabled={isLoading}
                error={errors.new_password}
                {...register("new_password")}
              />

              <TextField
                id="confirmPassword"
                label="パスワード確認"
                type="password"
                autoComplete="new-password"
                disabled={isLoading}
                error={errors.confirm_password}
                {...register("confirm_password")}
              />

              {error && (
                <div
                  className="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
                  role="alert"
                >
                  {error}
                </div>
              )}

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    変更中...
                  </>
                ) : (
                  "パスワードを変更"
                )}
              </Button>
            </form>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
