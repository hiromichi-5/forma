export type MemberRole = "admin" | "editor"

export const ROLE_LABELS: Record<MemberRole, string> = {
  admin: "管理者",
  editor: "編集者",
}

/** Select の値（string）を権限に戻す。想定外の値は権限の弱い editor とする。 */
export const toRoleValue = (value: string): MemberRole => (value === "admin" ? "admin" : "editor")

export type MemberView = {
  id: string
  name: string
  email: string
  role: MemberRole
}
