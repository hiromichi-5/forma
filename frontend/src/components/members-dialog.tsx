"use client"

import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { UserPlus, X, Shield, User } from "lucide-react"

interface Member {
  id: string
  name: string
  email: string
  role: "admin" | "editor"
}

const mockMembers: Member[] = [
  { id: "1", name: "田中 太郎", email: "tanaka@example.com", role: "admin" },
  { id: "2", name: "佐藤 花子", email: "sato@example.com", role: "editor" },
  { id: "3", name: "鈴木 一郎", email: "suzuki@example.com", role: "editor" },
]

interface MembersDialogProps {
  formId: string
  onClose: () => void
}

export function MembersDialog({ formId, onClose }: MembersDialogProps) {
  const [members, setMembers] = useState<Member[]>(mockMembers)
  const [newMemberEmail, setNewMemberEmail] = useState("")
  const [newMemberRole, setNewMemberRole] = useState<"admin" | "editor">("editor")

  const handleAddMember = () => {
    if (newMemberEmail) {
      const newMember: Member = {
        id: `member-${Date.now()}`,
        name: newMemberEmail.split("@")[0],
        email: newMemberEmail,
        role: newMemberRole,
      }
      setMembers([...members, newMember])
      setNewMemberEmail("")
      setNewMemberRole("editor")
    }
  }

  const handleRemoveMember = (id: string) => {
    setMembers(members.filter((m) => m.id !== id))
  }

  const handleRoleChange = (id: string, role: "admin" | "editor") => {
    setMembers(members.map((m) => (m.id === id ? { ...m, role } : m)))
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
                <Select value={newMemberRole} onValueChange={(v) => setNewMemberRole(v as "admin" | "editor")}>
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
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
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
                      onValueChange={(v) => handleRoleChange(member.id, v as "admin" | "editor")}
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
