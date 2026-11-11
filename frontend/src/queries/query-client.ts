import { QueryClient } from "@tanstack/react-query"
import { ApiError } from "@/lib/api"

/** 4xx はリクエスト自体の誤りや権限の問題なので、再試行しても結果は変わらない。 */
const isClientError = (error: unknown) =>
  error instanceof ApiError && error.status >= 400 && error.status < 500

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // 同じ画面を行き来するたびに取得し直さないよう、短時間は取得済みのデータを使う。
        staleTime: 30_000,
        retry: (failureCount, error) => !isClientError(error) && failureCount < 2,
      },
    },
  })
}
