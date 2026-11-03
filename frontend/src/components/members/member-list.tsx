import { Shield, User, X } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { ROLE_LABELS, toRoleValue, type MemberRole, type MemberView } from "./role"

type MemberListProps = {
  members: MemberView[]
  isLoading: boolean
  onRoleChange: (memberId: string, role: MemberRole) => void
  onRemove: (member: MemberView) => void
}

export function MemberList({ members, isLoading, onRoleChange, onRemove }: MemberListProps) {
  return (
    <div className="space-y-2">
      <h3 className="text-sm font-semibold text-foreground">現在のメンバー ({members.length})</h3>
      <div className="max-h-[280px] space-y-1.5 overflow-y-auto">
        {isLoading && <div className="p-2 text-sm text-muted-foreground">読み込み中...</div>}
        {members.map((member) => (
          <div
            key={member.id}
            className="flex items-center justify-between rounded-lg border bg-card p-2"
          >
            <div className="flex flex-1 items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
                {member.role === "admin" ? (
                  <Shield className="h-4 w-4 text-primary" />
                ) : (
                  <User className="h-4 w-4 text-muted-foreground" />
                )}
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium text-foreground">{member.name}</p>
                <p className="text-xs text-muted-foreground">{member.email}</p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Select
                value={member.role}
                onValueChange={(v) => onRoleChange(member.id, toRoleValue(v))}
              >
                <SelectTrigger className="h-8 w-[110px]" aria-label={`${member.name} の権限`}>
                  <Badge variant={member.role === "admin" ? "default" : "secondary"}>
                    {ROLE_LABELS[member.role]}
                  </Badge>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">{ROLE_LABELS.admin}</SelectItem>
                  <SelectItem value="editor">{ROLE_LABELS.editor}</SelectItem>
                </SelectContent>
              </Select>
              <Tooltip>
                <TooltipTrigger asChild>
                  <span className="inline-flex">
                    <Button
                      variant="ghost"
                      size="icon"
                      disabled={member.role === "admin"}
                      onClick={() => onRemove(member)}
                      aria-label={`${member.name} をメンバーから削除`}
                      className="h-8 w-8"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </span>
                </TooltipTrigger>
                <TooltipContent>
                  {member.role === "admin" ? "管理者は削除できません" : "メンバーを削除"}
                </TooltipContent>
              </Tooltip>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
