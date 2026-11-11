import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import type { z } from "zod"
import { Link, Navigate } from "react-router"
import { useAuth } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { TextField } from "@/components/text-field"
import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { getApiErrorMessage } from "@/lib/api-error"
import { PASSWORD_MIN_LENGTH, signupSchema } from "@/lib/form-schemas"
import { Loader2, MailCheck } from "lucide-react"

export default function SignupPage() {
  const { user } = useAuth()
  const { signup } = useAuth()
  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors, isSubmitting: isLoading },
  } = useForm<z.infer<typeof signupSchema>>({
    resolver: zodResolver(signupSchema),
    defaultValues: { display_name: "", email: "", password: "" },
  })
  const [error, setError] = useState("")
  const [isSignupComplete, setIsSignupComplete] = useState(false)

  if (user) {
    return <Navigate to="/" replace />
  }

  const onSubmit = async (values: z.infer<typeof signupSchema>) => {
    setError("")

    try {
      await signup(values)
      setIsSignupComplete(true)
    } catch (err) {
      setError(
        getApiErrorMessage(
          err,
          {
            CONFLICT: "このメールアドレスは既に登録されています",
            VALIDATION_ERROR: "入力内容を確認してください（パスワードは8文字以上）",
          },
          "サインアップに失敗しました",
        ),
      )
    }
  }

  if (isSignupComplete) {
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
                <CardTitle className="text-2xl font-bold">確認メールを送信しました</CardTitle>
                <p className="mt-3 text-sm text-muted-foreground">
                  <span className="font-medium text-foreground">{getValues("email")}</span>{" "}
                  に確認メールを送信しました。
                  <br />
                  メール内のリンクをクリックしてアカウントを有効化してください。
                </p>
              </div>
              <div className="pt-2">
                <Link to="/login" className="text-sm font-medium text-primary hover:underline">
                  ログインページへ
                </Link>
              </div>
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
              <CardTitle className="text-2xl font-bold">forma</CardTitle>
              <p className="mt-2 text-sm text-muted-foreground">
                新しいアカウントを作成して
                <br />
                フォーム管理を始めましょう
              </p>
            </div>

            <form className="space-y-4" noValidate onSubmit={handleSubmit(onSubmit)}>
              <TextField
                id="displayName"
                label="表示名"
                autoComplete="name"
                disabled={isLoading}
                error={errors.display_name}
                {...register("display_name")}
              />

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
                autoComplete="new-password"
                description={`${PASSWORD_MIN_LENGTH}文字以上`}
                disabled={isLoading}
                error={errors.password}
                {...register("password")}
              />

              {error && (
                <div
                  className="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
                  role="alert"
                  aria-live="polite"
                >
                  {error}
                </div>
              )}

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
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
                <Link to="/login" className="font-medium text-primary hover:underline">
                  ログイン
                </Link>
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
