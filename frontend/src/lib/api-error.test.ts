import { describe, expect, it, vi } from "vitest"
import { toast } from "sonner"
import { ApiError } from "@/lib/api"
import { getApiErrorMessage, hasApiErrorCode, toastApiError } from "@/lib/api-error"

const notFound = new ApiError(404, { code: "FORM_NOT_FOUND", message: "not found" })

describe("hasApiErrorCode", () => {
  it("指定したコードの ApiError なら true", () => {
    expect(hasApiErrorCode(notFound, "FORM_NOT_FOUND")).toBe(true)
    expect(hasApiErrorCode(notFound, "FORBIDDEN")).toBe(false)
  })

  it("ApiError 以外は false", () => {
    expect(hasApiErrorCode(new Error("x"), "FORM_NOT_FOUND")).toBe(false)
  })
})

describe("getApiErrorMessage", () => {
  it("コードに対応する文言を返す", () => {
    expect(getApiErrorMessage(notFound, { FORM_NOT_FOUND: "見つかりません" }, "失敗")).toBe(
      "見つかりません",
    )
  })

  it("画面ごとの文言が無ければ、共通の文言を使う", () => {
    const network = new ApiError(0, { code: "NETWORK_ERROR" })
    expect(getApiErrorMessage(network, {}, "失敗")).toMatch(/ネットワークエラー/)
  })

  it("対応する文言が無ければ fallback を返す", () => {
    expect(getApiErrorMessage(notFound, {}, "失敗")).toBe("失敗")
    expect(getApiErrorMessage(new Error("x"), { FORM_NOT_FOUND: "x" }, "失敗")).toBe("失敗")
  })
})

describe("toastApiError", () => {
  it("エラーの文言をトーストで表示する", () => {
    const spy = vi.spyOn(toast, "error")
    toastApiError(notFound, { FORM_NOT_FOUND: "見つかりません" }, "失敗")
    expect(spy).toHaveBeenCalledWith("見つかりません")
  })

  it("セッション失効は AuthProvider が通知するため、重ねて表示しない", () => {
    const spy = vi.spyOn(toast, "error")
    toastApiError(new ApiError(401, { code: "INVALID_SESSION" }), {}, "失敗")
    expect(spy).not.toHaveBeenCalled()
  })
})
