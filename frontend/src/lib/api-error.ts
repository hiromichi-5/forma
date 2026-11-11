import { toast } from "sonner"
import { ApiError, type ErrorCode } from "@/lib/api"

export type ApiErrorMessageMap = Partial<Record<ErrorCode, string>>

/** どの画面でも意味が変わらないエラーの文言。画面ごとの文言で上書きできる。 */
const DEFAULT_MESSAGES: ApiErrorMessageMap = {
  NETWORK_ERROR: "ネットワークエラーが発生しました。接続を確認してください",
  UNEXPECTED_RESPONSE: "サーバーで問題が発生しました。しばらくしてから再度お試しください",
  INTERNAL: "サーバーで問題が発生しました。しばらくしてから再度お試しください",
  INVALID_SESSION: "セッションの有効期限が切れました。ログインし直してください",
  RATE_LIMITED: "操作の回数が上限に達しました。しばらくしてから再度お試しください",
  FORBIDDEN: "この操作を行う権限がありません",
}

export function hasApiErrorCode(error: unknown, code: ErrorCode): error is ApiError {
  return error instanceof ApiError && error.error.code === code
}

/** エラーに対応する文言を、画面ごとの文言 → 共通の文言 → fallback の順で選ぶ。 */
export function getApiErrorMessage(
  error: unknown,
  messages: ApiErrorMessageMap,
  fallback: string,
): string {
  if (!(error instanceof ApiError)) {
    return fallback
  }
  const code = error.error.code
  return messages[code] ?? DEFAULT_MESSAGES[code] ?? fallback
}

/**
 * エラーをトーストで知らせる。
 * セッション失効は AuthProvider がまとめて通知してログイン画面へ戻すため、ここでは重ねて出さない。
 */
export function toastApiError(error: unknown, messages: ApiErrorMessageMap, fallback: string) {
  if (hasApiErrorCode(error, "INVALID_SESSION")) return
  toast.error(getApiErrorMessage(error, messages, fallback))
}
