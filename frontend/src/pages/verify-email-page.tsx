import { useEffect, useRef } from "react"
import { useMutation } from "@tanstack/react-query"
import { Link, useSearchParams } from "react-router"
import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import { CheckCircle, Loader2, XCircle } from "lucide-react"

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get("token")
  const verify = useMutation({
    mutationFn: (token: string) => apiClient.verifyEmail({ token }),
  })
  // トークンは一度使うと無効になる。開発時の StrictMode で effect が 2 回走っても、送信は 1 回にする。
  const requestedTokenRef = useRef<string | null>(null)
  const { mutate } = verify

  useEffect(() => {
    if (!token || requestedTokenRef.current === token) return
    requestedTokenRef.current = token
    mutate(token)
  }, [token, mutate])

  const status: "loading" | "success" | "error" = !token
    ? "error"
    : verify.isSuccess
      ? "success"
      : verify.isError
        ? "error"
        : "loading"
  const errorMessage = !token
    ? "認証トークンが見つかりません"
    : getApiErrorMessage(
        verify.error,
        {
          TOKEN_NOT_FOUND: "トークンが無効または期限切れです",
          VALIDATION_ERROR: "認証トークンが不正です",
        },
        "認証に失敗しました",
      )

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-6">
          <div className="space-y-6 text-center">
            {status === "loading" && (
              <>
                <Loader2 className="mx-auto h-12 w-12 animate-spin text-primary" />
                <CardTitle className="text-xl">メールアドレスを確認中...</CardTitle>
              </>
            )}

            {status === "success" && (
              <>
                <div className="flex justify-center">
                  <div className="rounded-full bg-green-100 p-4">
                    <CheckCircle className="h-12 w-12 text-green-600" />
                  </div>
                </div>
                <div>
                  <CardTitle className="text-xl">メール認証が完了しました</CardTitle>
                  <p className="mt-2 text-sm text-muted-foreground">
                    アカウントが有効化されました。ログインしてご利用ください。
                  </p>
                </div>
                <Link to="/login" className="inline-block font-medium text-primary hover:underline">
                  ログインページへ
                </Link>
              </>
            )}

            {status === "error" && (
              <>
                <div className="flex justify-center">
                  <div className="rounded-full bg-destructive/10 p-4">
                    <XCircle className="h-12 w-12 text-destructive" />
                  </div>
                </div>
                <div>
                  <CardTitle className="text-xl">認証に失敗しました</CardTitle>
                  <p className="mt-2 text-sm text-muted-foreground">{errorMessage}</p>
                </div>
                <Link to="/login" className="inline-block font-medium text-primary hover:underline">
                  ログインページへ
                </Link>
              </>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
