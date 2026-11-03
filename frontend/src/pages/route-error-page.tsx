import { isRouteErrorResponse, Link, useRouteError } from "react-router"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardTitle } from "@/components/ui/card"

/**
 * 画面の描画中に予期しないエラーが起きたときの表示。
 * これが無いと、どこか 1 か所の例外でアプリ全体が真っ白になる。
 */
export default function RouteErrorPage() {
  const error = useRouteError()
  console.error(error)
  const detail = isRouteErrorResponse(error)
    ? `${error.status} ${error.statusText}`
    : error instanceof Error
      ? error.message
      : null

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="space-y-4 p-6 text-center">
          <CardTitle className="text-xl">問題が発生しました</CardTitle>
          <p className="text-sm text-muted-foreground">
            画面の表示中にエラーが発生しました。再読み込みしても解決しない場合は、時間を置いてお試しください。
          </p>
          {import.meta.env.DEV && detail && (
            <pre className="overflow-auto rounded bg-muted p-2 text-left text-xs">{detail}</pre>
          )}
          <div className="flex justify-center gap-2">
            <Button onClick={() => window.location.reload()}>再読み込み</Button>
            <Button variant="outline" asChild>
              <Link to="/">ホームへ</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
