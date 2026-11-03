import { Clock, X } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import type { FormInvite } from "@/types"
import { ROLE_LABELS } from "./role"

const formatExpiresAt = (expiresAt: string) => {
  const date = new Date(expiresAt)
  return `${date.getMonth() + 1}/${date.getDate()} まで`
}

type InviteListProps = {
  invites: FormInvite[]
  onRevoke: (invite: FormInvite) => void
}

/** 承諾待ちの招待の一覧。招待が無ければ何も表示しない。 */
export function InviteList({ invites, onRevoke }: InviteListProps) {
  if (invites.length === 0) return null

  return (
    <div className="space-y-2">
      <h3 className="text-sm font-semibold text-foreground">保留中の招待 ({invites.length})</h3>
      <div className="max-h-[200px] space-y-1.5 overflow-y-auto">
        {invites.map((invite) => (
          <div
            key={invite.id}
            className="flex items-center justify-between rounded-lg border border-dashed bg-card p-2"
          >
            <div className="flex flex-1 items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
                <Clock className="h-4 w-4 text-muted-foreground" />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium text-foreground">{invite.email}</p>
                <p className="text-xs text-muted-foreground">
                  {formatExpiresAt(invite.expires_at)}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="secondary">{ROLE_LABELS[invite.role]}</Badge>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => onRevoke(invite)}
                aria-label={`${invite.email} への招待をキャンセル`}
                className="h-8 w-8"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
