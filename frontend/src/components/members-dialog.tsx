"use client"

import { useEffect, useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { UserPlus, X, Shield, User } from "lucide-react"
import { apiClient } from "@/lib/api"

type MemberView = {
  id: string
  name: string
  email: string
  role: "admin" | "editor"
}

type MembersDialogProps = {
  formId: string
  onClose: () => void
}

export function MembersDialog({ formId, onClose }: MembersDialogProps) {
  const [members, setMembers] = useState<MemberView[]>([])
  const [newMemberEmail, setNewMemberEmail] = useState("")
  const [newMemberRole, setNewMemberRole] = useState<"admin" | "editor">("editor")
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState("")

  const toRoleValue = (value: string): "admin" | "editor" => (value === "admin" ? "admin" : "editor")

  const loadMembers = async () => {
    setIsLoading(true)
    setErrorMessage("")
    try {
      // 外部API(バックエンド)との同期のための処理
      const response = await apiClient.getMembers(formId)
      setMembers(
        response.members.map((member) => ({
          id: member.id,
          name: member.display_name,
          email: member.email,
          role: member.role,
        }))
      )
    } catch (error) {
      console.error("Failed to load members:", error)
      setErrorMessage("メンバーの取得に失敗しました")
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    // 外部API(バックエンド)との同期のための処理
    void loadMembers()
  }, [formId])

  const handleAddMember = async () => {
    if (!newMemberEmail) return

    try {
      await apiClient.addMember(formId, {
        email: newMemberEmail,
        role: newMemberRole,
      })
      setNewMemberEmail("")
      setNewMemberRole("editor")
      await loadMembers()
    } catch (error) {
      console.error("Failed to add member:", error)
      setErrorMessage("メンバーの追加に失敗しました")
    }
  }

  const handleRemoveMember = async (id: string) => {
    try {
      await apiClient.removeMember(formId, id)
      await loadMembers()
    } catch (error) {
      console.error("Failed to remove member:", error)
      setErrorMessage("メンバーの削除に失敗しました")
    }
  }

  const handleRoleChange = async (id: string, role: "admin" | "editor") => {
    try {
      await apiClient.changeMemberRole(formId, id, { role })
      await loadMembers()
    } catch (error) {
      console.error("Failed to change member role:", error)
      setErrorMessage("権限変更に失敗しました")
    }
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>メンバー管理</DialogTitle>
          <DialogDescription>このフォームにアクセスできるメンバーを管理します</DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          <div className="space-y-3 p-4 bg-muted/30 rounded-lg">
            <h3 className="text-sm font-semibold text-foreground">新しいメンバーを追加</h3>
            <div className="flex gap-2">
              <div className="flex-1 space-y-2">
                <Label htmlFor="email" className="text-xs">
                  メールアドレス
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="example@company.com"
                  value={newMemberEmail}
                  onChange={(e) => setNewMemberEmail(e.target.value)}
                />
              </div>
              <div className="w-[140px] space-y-2">
                <Label className="text-xs">権限</Label>
                <Select value={newMemberRole} onValueChange={(v) => setNewMemberRole(toRoleValue(v))}>
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
                <Button onClick={handleAddMember} className="gap-2">
                  <UserPlus className="h-4 w-4" />
                  追加
                </Button>
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-foreground">現在のメンバー ({members.length})</h3>
            {errorMessage && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">{errorMessage}</div>
            )}
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
              {isLoading && (
                <div className="p-3 text-sm text-muted-foreground">メンバーを読み込み中...</div>
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
