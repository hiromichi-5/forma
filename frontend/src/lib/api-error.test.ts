import { describe, expect, it } from "vitest"
import { ApiError } from "@/lib/api"
import { getApiErrorMessage, hasApiErrorCode } from "@/lib/api-error"

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

  it("対応する文言が無ければ fallback を返す", () => {
    expect(getApiErrorMessage(notFound, {}, "失敗")).toBe("失敗")
    expect(getApiErrorMessage(new Error("x"), { FORM_NOT_FOUND: "x" }, "失敗")).toBe("失敗")
  })
})
