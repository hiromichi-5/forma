import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import type { z } from "zod"
import { Link } from "react-router"
import { Button } from "@/components/ui/button"
import { TextField } from "@/components/text-field"
import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import { passwordResetRequestSchema } from "@/lib/form-schemas"
import { Loader2, MailCheck } from "lucide-react"

export default function PasswordResetPage() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting: isLoading },
  } = useForm<z.infer<typeof passwordResetRequestSchema>>({
    resolver: zodResolver(passwordResetRequestSchema),
    defaultValues: { email: "" },
  })
  const [error, setError] = useState("")
  const [isSent, setIsSent] = useState(false)

  const onSubmit = async (values: z.infer<typeof passwordResetRequestSchema>) => {
    setError("")

    try {
      await apiClient.passwordReset(values)
      setIsSent(true)
    } catch (err) {
      setError(
        getApiErrorMessage(
          err,
          {
            VALIDATION_ERROR: "有効なメールアドレスを入力してください",
          },
          "リセットメールの送信に失敗しました",
        ),
      )
    }
  }

  if (isSent) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted p-4">
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
                className="inline-block text-sm font-medium text-primary hover:underline"
              >
                ログインページへ戻る
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
              <CardTitle className="text-2xl font-bold">パスワードリセット</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                登録したメールアドレスを入力してください。
                <br />
                パスワードリセット用のリンクを送信します。
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
                    送信中...
                  </>
                ) : (
                  "リセットメールを送信"
                )}
              </Button>
            </form>

            <div className="text-center">
              <Link to="/login" className="text-sm text-muted-foreground hover:underline">
                ログインページへ戻る
              </Link>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
