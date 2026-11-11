import { useState } from "react"
import { Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { InviteForm } from "@/components/members/invite-form"
import { InviteList } from "@/components/members/invite-list"
import { MemberList } from "@/components/members/member-list"
import { ColorInput } from "@/components/statuses/color-input"
import { createColorState } from "@/components/statuses/color-state"
import { StatusRow } from "@/components/statuses/status-row"
import { NotificationSettingsFields } from "@/components/notification-settings-fields"
import { DEMO_MEMBERS } from "@/components/welcome/demo-data"
import { sortStatuses } from "@/lib/ticket-display"
import type { NotificationEmailSample } from "@/lib/notification-email-preview"
import type { FormInvite, FormStatus, NotificationSetting, NotificationType } from "@/types"

// 設定の画面は見た目だけを再現する。保存の操作はアカウント作成へ案内する。
type DemoDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onBlockedAction: () => void
}

export function DemoStatusesDialog({
  statuses,
  ...props
}: DemoDialogProps & { statuses: FormStatus[] }) {
  const { open, onOpenChange, onBlockedAction } = props
  const sortedStatuses = sortStatuses(statuses)
  const [newColor, setNewColor] = useState(() => createColorState(null))

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>ステータス管理</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <Input
              placeholder="新しいステータス名"
              aria-label="新しいステータス名"
              className="flex-1"
            />
            <ColorInput value={newColor} onChange={setNewColor} />
            <Button onClick={onBlockedAction} size="sm" className="gap-1.5">
              <Plus className="h-4 w-4" />
              追加する
            </Button>
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium text-muted-foreground">ステータス一覧</h3>
              <span className="text-xs text-muted-foreground">{sortedStatuses.length}件</span>
            </div>
            <div className="space-y-1">
              {sortedStatuses.map((status, index) => (
                <div key={status.id} className="rounded-md border bg-card">
                  <StatusRow
                    status={status}
                    isFirst={index === 0}
                    isLast={index === sortedStatuses.length - 1}
                    disabled={false}
                    onEdit={onBlockedAction}
                    onSetDefault={onBlockedAction}
                    onMove={onBlockedAction}
                    onDelete={onBlockedAction}
                  />
                </div>
              ))}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

const DEMO_INVITES: FormInvite[] = [
  {
    id: "demo-invite-1",
    email: "new-member@example.com",
    role: "editor",
    invited_by: DEMO_MEMBERS[0].id,
    expires_at: new Date(Date.now() + 6 * 24 * 60 * 60_000).toISOString(),
    created_at: new Date().toISOString(),
  },
]

export function DemoMembersDialog({ open, onOpenChange, onBlockedAction }: DemoDialogProps) {
  const members = DEMO_MEMBERS.map((member) => ({
    id: member.id,
    name: member.display_name,
    email: member.email,
    role: member.role,
  }))

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>メンバー管理</DialogTitle>
          <DialogDescription>このフォームにアクセスできるメンバーを管理します</DialogDescription>
        </DialogHeader>
        <div className="space-y-6">
          <div className="space-y-3 rounded-lg border border-dashed bg-muted/60 p-4">
            <h3 className="text-sm font-semibold text-foreground">メンバーを招待する</h3>
            <InviteForm
              onInvite={() => {
                onBlockedAction()
                return Promise.resolve(false)
              }}
            />
          </div>
          <InviteList invites={DEMO_INVITES} onRevoke={onBlockedAction} />
          <MemberList
            members={members}
            isLoading={false}
            onRoleChange={onBlockedAction}
            onRemove={onBlockedAction}
          />
        </div>
      </DialogContent>
    </Dialog>
  )
}

const DEMO_NOTIFICATION_SETTINGS: NotificationSetting[] = [
  { notification_type: "status_change", mode: "confirm", include_detail: true },
  { notification_type: "assignee_assigned", mode: "off", include_detail: false },
]

export function DemoNotificationsDialog({
  emailSample,
  ...props
}: DemoDialogProps & { emailSample: NotificationEmailSample }) {
  const { open, onOpenChange, onBlockedAction } = props
  const [settings, setSettings] = useState(DEMO_NOTIFICATION_SETTINGS)

  const handleChange = (type: NotificationType, patch: Partial<NotificationSetting>) =>
    setSettings((prev) => prev.map((s) => (s.notification_type === type ? { ...s, ...patch } : s)))

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>通知設定</DialogTitle>
          <DialogDescription>
            チケットが更新されたときに、回答者へメールで通知します。
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-6">
          <NotificationSettingsFields
            settings={settings}
            onChange={handleChange}
            disabled={false}
            emailSample={emailSample}
          />
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              キャンセル
            </Button>
            <Button onClick={onBlockedAction}>保存</Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
