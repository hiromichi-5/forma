import { describe, expect, it } from "vitest"
import {
  changePasswordSchema,
  loginSchema,
  registerFormSchema,
  signupSchema,
} from "@/lib/form-schemas"

const firstError = (result: { success: boolean; error?: { issues: { message: string }[] } }) =>
  result.error?.issues[0]?.message

describe("form schemas", () => {
  it("新しく登録するメールアドレスは形式を検証する", () => {
    expect(
      firstError(signupSchema.safeParse({ display_name: "a", email: "foo", password: "abcdefgh" })),
    ).toBe("有効なメールアドレスを入力してください")
  })

  it("ログインでは登録済みのアドレスを形式で弾かず、未入力だけを検出する", () => {
    expect(loginSchema.safeParse({ email: "admin@localhost", password: "x" }).success).toBe(true)
    expect(firstError(loginSchema.safeParse({ email: "", password: "x" }))).toBe(
      "メールアドレスを入力してください",
    )
  })

  it("新しいパスワードは 8 文字以上（API の minLength と同じ）", () => {
    const result = signupSchema.safeParse({
      display_name: "a",
      email: "a@example.com",
      password: "1234567",
    })
    expect(firstError(result)).toMatch(/8文字以上/)
  })

  it("表示名は前後の空白を除いて 1 文字以上", () => {
    const result = signupSchema.safeParse({
      display_name: "  ",
      email: "a@example.com",
      password: "12345678",
    })
    expect(firstError(result)).toBe("表示名を入力してください")
  })

  it("確認用のパスワードが一致しない場合は、確認欄のエラーにする", () => {
    const result = changePasswordSchema.safeParse({
      current_password: "old",
      new_password: "12345678",
      confirm_password: "12345679",
    })
    expect(result.error?.issues[0]).toMatchObject({
      path: ["confirm_password"],
      message: "パスワードが一致しません",
    })
  })

  it("フォームの URL は前後の空白を除いて必須", () => {
    expect(registerFormSchema.safeParse({ url: "  " }).success).toBe(false)
    expect(registerFormSchema.parse({ url: " https://docs.google.com/forms/d/abc/edit " })).toEqual(
      {
        url: "https://docs.google.com/forms/d/abc/edit",
      },
    )
  })
})
