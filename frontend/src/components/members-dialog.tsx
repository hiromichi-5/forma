"use client"

import { useEffect, useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Mail, X, Shield, User, Clock } from "lucide-react"
import { apiClient, ApiError } from "@/lib/api"
import type { FormInvite } from "@/types"

type MemberView = {
  id: string
  name: string
  email: string
  role: "admin" | "editor"
}

type MembersDialogProps = {
  formId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

const ERROR_MESSAGES: Record<string, string> = {
  ALREADY_MEMBER: "このメールアドレスは既にメンバーです",
  ACTIVE_INVITE_ALREADY_EXISTS: "このメールアドレスには既に有効な招待があります",
  FORBIDDEN: "この操作を行う権限がありません",
}

export function MembersDialog({ formId, open, onOpenChange }: MembersDialogProps) {
  const [members, setMembers] = useState<MemberView[]>([])
  const [invites, setInvites] = useState<FormInvite[]>([])
  const [inviteEmail, setInviteEmail] = useState("")
  const [inviteRole, setInviteRole] = useState<"admin" | "editor">("editor")
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState("")

  const toRoleValue = (value: string): "admin" | "editor" => (value === "admin" ? "admin" : "editor")

  const loadData = async () => {
    setIsLoading(true)
    setErrorMessage("")
    try {
      const [membersRes, invitesRes] = await Promise.all([
        apiClient.getMembers(formId),
        apiClient.listInvites(formId),
      ])
      setMembers(
        membersRes.members.map((member) => ({
          id: member.id,
          name: member.display_name,
          email: member.email,
          role: member.role,
        }))
      )
      setInvites(invitesRes.invites)
    } catch (error) {
      console.error("Failed to load data:", error)
      setErrorMessage("データの取得に失敗しました")
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    void loadData()
  }, [formId])

  const handleInvite = async () => {
    if (!inviteEmail) return
    setErrorMessage("")

    try {
      await apiClient.createInvite(formId, {
        email: inviteEmail,
        role: inviteRole,
      })
      setInviteEmail("")
      setInviteRole("editor")
      await loadData()
    } catch (error) {
      console.error("Failed to create invite:", error)
      if (error instanceof ApiError) {
        setErrorMessage(ERROR_MESSAGES[error.error.code] ?? "招待の送信に失敗しました")
      } else {
        setErrorMessage("招待の送信に失敗しました")
      }
    }
  }

  const handleRevokeInvite = async (inviteId: string) => {
    setErrorMessage("")
    try {
      await apiClient.revokeInvite(formId, inviteId)
      await loadData()
    } catch (error) {
      console.error("Failed to revoke invite:", error)
      setErrorMessage("招待の取消に失敗しました")
    }
  }

  const handleRemoveMember = async (id: string) => {
    setErrorMessage("")
    try {
      await apiClient.removeMember(formId, id)
      await loadData()
    } catch (error) {
      console.error("Failed to remove member:", error)
      setErrorMessage("メンバーの削除に失敗しました")
    }
  }

  const handleRoleChange = async (id: string, role: "admin" | "editor") => {
    setErrorMessage("")
    try {
      await apiClient.changeMemberRole(formId, id, { role })
      await loadData()
    } catch (error) {
      console.error("Failed to change member role:", error)
      setErrorMessage("権限変更に失敗しました")
    }
  }

  const formatExpiresAt = (expiresAt: string) => {
    const date = new Date(expiresAt)
    return `${date.getMonth() + 1}/${date.getDate()} まで`
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>メンバー管理</DialogTitle>
          <DialogDescription>このフォームにアクセスできるメンバーを管理します</DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {errorMessage && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">{errorMessage}</div>
          )}

          <div className="space-y-3 p-4 bg-muted/30 rounded-lg">
            <h3 className="text-sm font-semibold text-foreground">メンバーを招待</h3>
            <div className="flex gap-2">
              <div className="flex-1 space-y-2">
                <Label htmlFor="email" className="text-xs">
                  メールアドレス
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="example@company.com"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                />
              </div>
              <div className="w-[140px] space-y-2">
                <Label className="text-xs">権限</Label>
                <Select value={inviteRole} onValueChange={(v) => setInviteRole(toRoleValue(v))}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="admin">管理者</SelectItem>
                    <SelectItem value="editor">編集者</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-end">
                <Button onClick={handleInvite} className="gap-2">
                  <Mail className="h-4 w-4" />
                  招待
                </Button>
              </div>
            </div>
          </div>

          {invites.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-semibold text-foreground">保留中の招待 ({invites.length})</h3>
              <div className="space-y-2">
                {invites.map((invite) => (
                  <div key={invite.id} className="flex items-center justify-between p-3 bg-card border border-dashed rounded-lg">
                    <div className="flex items-center gap-3 flex-1">
                      <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center">
                        <Clock className="h-5 w-5 text-muted-foreground" />
                      </div>
                      <div className="flex-1">
                        <p className="font-medium text-sm text-foreground">{invite.email}</p>
                        <p className="text-xs text-muted-foreground">{formatExpiresAt(invite.expires_at)}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary">
                        {invite.role === "admin" ? "管理者" : "編集者"}
                      </Badge>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRevokeInvite(invite.id)}
                        className="h-8 w-8"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-foreground">現在のメンバー ({members.length})</h3>
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
              {isLoading && (
                <div className="p-3 text-sm text-muted-foreground">読み込み中...</div>
              )}
              {members.map((member) => (
                <div key={member.id} className="flex items-center justify-between p-3 bg-card border rounded-lg">
                  <div className="flex items-center gap-3 flex-1">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                      {member.role === "admin" ? (
                        <Shield className="h-5 w-5 text-primary" />
                      ) : (
                        <User className="h-5 w-5 text-muted-foreground" />
                      )}
                    </div>
                    <div className="flex-1">
                      <p className="font-medium text-sm text-foreground">{member.name}</p>
                      <p className="text-xs text-muted-foreground">{member.email}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Select
                      value={member.role}
                      onValueChange={(v) => handleRoleChange(member.id, toRoleValue(v))}
                    >
                      <SelectTrigger className="w-[110px] h-8">
                        <Badge variant={member.role === "admin" ? "default" : "secondary"}>
                          {member.role === "admin" ? "管理者" : "編集者"}
                        </Badge>
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="admin">管理者</SelectItem>
                        <SelectItem value="editor">編集者</SelectItem>
                      </SelectContent>
                    </Select>
                    {member.role !== "admin" && (
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRemoveMember(member.id)}
                        className="h-8 w-8"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
