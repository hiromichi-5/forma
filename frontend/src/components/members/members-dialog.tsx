import { useMemo, useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import type { FormInvite } from "@/types"
import { formQueries } from "@/queries/forms"
import { queryKeys } from "@/queries/keys"
import { InviteForm, type InviteValues } from "./invite-form"
import { InviteList } from "./invite-list"
import { MemberList } from "./member-list"
import type { MemberRole, MemberView } from "./role"

const EMPTY_INVITES: FormInvite[] = []

type MembersDialogProps = {
  formId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

const inviteErrorMessages = {
  ALREADY_MEMBER: "このメールアドレスは既にメンバーです",
  ACTIVE_INVITE_ALREADY_EXISTS: "このメールアドレスには既に有効な招待があります",
  RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
} as const

export function MembersDialog({ formId, open, onOpenChange }: MembersDialogProps) {
  const queryClient = useQueryClient()
  const membersQuery = useQuery(formQueries.members(formId))
  // 招待はこのダイアログでしか使わないため、開いている間だけ取得する。
  const invitesQuery = useQuery({ ...formQueries.invites(formId), enabled: open })
  const members = useMemo<MemberView[]>(
    () =>
      (membersQuery.data ?? []).map((member) => ({
        id: member.id,
        name: member.display_name,
        email: member.email,
        role: member.role,
      })),
    [membersQuery.data],
  )
  const invites = invitesQuery.data ?? EMPTY_INVITES
  const isLoading = membersQuery.isPending
  const loadError = membersQuery.error ?? invitesQuery.error
  const [errorMessage, setErrorMessage] = useState("")
  const displayedError =
    errorMessage ||
    (loadError
      ? getApiErrorMessage(
          loadError,
          { RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません" },
          "データの取得に失敗しました",
        )
      : "")
  const [isWorking, setIsWorking] = useState(false)
  const [revokeTarget, setRevokeTarget] = useState<FormInvite | null>(null)
  const [revokeError, setRevokeError] = useState("")
  const [removeTarget, setRemoveTarget] = useState<MemberView | null>(null)
  const [removeError, setRemoveError] = useState("")

  // メンバーの変更はフォーム管理画面（管理者判定・担当者の選択肢）にも影響するため、共有のキャッシュを取り直す。
  const loadData = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.formMembers(formId) }),
      queryClient.invalidateQueries({ queryKey: queryKeys.formInvites(formId) }),
    ])

  const handleInvite = async (values: InviteValues) => {
    setErrorMessage("")

    try {
      await apiClient.createInvite(formId, values)
      await loadData()
      return true
    } catch (error) {
      console.error("Failed to create invite:", error)
      setErrorMessage(
        getApiErrorMessage(
          error,
          {
            ...inviteErrorMessages,
            VALIDATION_ERROR: "招待する権限またはメールアドレスを確認してください",
          },
          "招待の送信に失敗しました",
        ),
      )
      return false
    }
  }

  const handleRevokeInvite = async () => {
    if (!revokeTarget) return
    setIsWorking(true)
    setRevokeError("")
    try {
      await apiClient.revokeInvite(formId, revokeTarget.id)
      await loadData()
      setRevokeTarget(null)
    } catch (error) {
      console.error("Failed to revoke invite:", error)
      setRevokeError(
        getApiErrorMessage(
          error,
          {
            RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
            INVITE_NOT_FOUND: "招待は既に存在しないか、別フォームのものです",
          },
          "招待の取消に失敗しました",
        ),
      )
    } finally {
      setIsWorking(false)
    }
  }

  const handleRemoveMember = async () => {
    if (!removeTarget) return
    setIsWorking(true)
    setRemoveError("")
    try {
      await apiClient.removeMember(formId, removeTarget.id)
      await loadData()
      setRemoveTarget(null)
    } catch (error) {
      console.error("Failed to remove member:", error)
      setRemoveError(
        getApiErrorMessage(
          error,
          {
            RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
            LAST_ADMIN: "最後の管理者は削除できません",
          },
          "メンバーの削除に失敗しました",
        ),
      )
    } finally {
      setIsWorking(false)
    }
  }

  const handleRoleChange = async (id: string, role: MemberRole) => {
    setErrorMessage("")
    try {
      await apiClient.changeMemberRole(formId, id, { role })
      await loadData()
    } catch (error) {
      console.error("Failed to change member role:", error)
      setErrorMessage(
        getApiErrorMessage(
          error,
          {
            RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
            LAST_ADMIN: "最後の管理者を編集者に変更することはできません",
            VALIDATION_ERROR: "変更する権限を確認してください",
          },
          "権限変更に失敗しました",
        ),
      )
    }
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>メンバー管理</DialogTitle>
            <DialogDescription>このフォームにアクセスできるメンバーを管理します</DialogDescription>
          </DialogHeader>

          <div className="space-y-6">
            {displayedError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {displayedError}
              </div>
            )}

            <div className="space-y-3 rounded-lg border border-dashed bg-muted/60 p-4">
              <h3 className="text-sm font-semibold text-foreground">メンバーを招待する</h3>
              <InviteForm onInvite={handleInvite} />
            </div>

            <InviteList invites={invites} onRevoke={setRevokeTarget} />

            <MemberList
              members={members}
              isLoading={isLoading}
              onRoleChange={handleRoleChange}
              onRemove={setRemoveTarget}
            />
          </div>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={revokeTarget !== null}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) {
            setRevokeTarget(null)
            setRevokeError("")
          }
        }}
      >
        <AlertDialogContent className="sm:max-w-xl">
          <AlertDialogHeader>
            <AlertDialogTitle>招待をキャンセルしますか？</AlertDialogTitle>
            <AlertDialogDescription>
              「{revokeTarget?.email}」への招待をキャンセルします。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {revokeError && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {revokeError}
            </div>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isWorking}>キャンセル</AlertDialogCancel>
            <Button onClick={handleRevokeInvite} disabled={isWorking} variant="destructive">
              招待をキャンセルする
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog
        open={removeTarget !== null}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) {
            setRemoveTarget(null)
            setRemoveError("")
          }
        }}
      >
        <AlertDialogContent className="sm:max-w-xl">
          <AlertDialogHeader>
            <AlertDialogTitle>メンバーを削除しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              「{removeTarget?.name}
              」をこのフォームのメンバーから削除します。一度削除したメンバーを追加する場合は、再度招待する必要があります。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {removeError && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {removeError}
            </div>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isWorking}>キャンセル</AlertDialogCancel>
            <Button onClick={handleRemoveMember} disabled={isWorking} variant="destructive">
              削除する
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
