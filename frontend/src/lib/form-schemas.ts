import { z } from "zod"

/**
 * フォームの入力ルール。openapi.yaml の制約（minLength など）と揃える。
 * サーバーでも同じ検証が行われるため、ここでの検証は送信前に誤りを伝えるためのもの。
 */

const email = z.email("有効なメールアドレスを入力してください")

export const PASSWORD_MIN_LENGTH = 8
const newPassword = z
  .string()
  .min(PASSWORD_MIN_LENGTH, `パスワードは${PASSWORD_MIN_LENGTH}文字以上で入力してください`)

const displayName = z.string().trim().min(1, "表示名を入力してください")

export const loginSchema = z.object({
  email,
  password: z.string().min(1, "パスワードを入力してください"),
})

export const signupSchema = z.object({
  display_name: displayName,
  email,
  password: newPassword,
})

export const passwordResetRequestSchema = z.object({ email })

/** 新しいパスワードと確認用の入力が一致することを確かめる。エラーは確認欄に表示する。 */
const passwordsMatch = {
  check: (v: { new_password: string; confirm_password: string }) =>
    v.new_password === v.confirm_password,
  params: { message: "パスワードが一致しません", path: ["confirm_password"] },
}

export const passwordResetConfirmSchema = z
  .object({ new_password: newPassword, confirm_password: z.string() })
  .refine(passwordsMatch.check, passwordsMatch.params)

export const changePasswordSchema = z
  .object({
    current_password: z.string().min(1, "現在のパスワードを入力してください"),
    new_password: newPassword,
    confirm_password: z.string(),
  })
  .refine(passwordsMatch.check, passwordsMatch.params)

export const profileSchema = z.object({ display_name: displayName })

// サーバーは編集画面の URL のほかフォーム ID 単体も受け付けるため、形式の検証はサーバーに任せる。
export const registerFormSchema = z.object({
  url: z.string().trim().min(1, "GoogleフォームのURLを入力してください"),
})

export const inviteSchema = z.object({
  email,
  role: z.enum(["admin", "editor"]),
})
