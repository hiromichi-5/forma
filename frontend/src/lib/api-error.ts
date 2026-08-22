import { ApiError, type ErrorCode } from "@/lib/api";

export type ApiErrorMessageMap = Partial<Record<ErrorCode, string>>;

export function hasApiErrorCode(
  error: unknown,
  code: ErrorCode
): error is ApiError {
  return error instanceof ApiError && error.error.code === code;
}

export function getApiErrorMessage(
  error: unknown,
  messages: ApiErrorMessageMap,
  fallback: string
): string {
  if (!(error instanceof ApiError)) {
    return fallback;
  }

  return messages[error.error.code] ?? fallback;
}
