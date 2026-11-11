import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { renderNotificationEmailPreview } from "@/lib/notification-email-preview"
import type { NotificationEmailSample } from "@/lib/notification-email-preview"
import type { NotificationMode, NotificationSetting, NotificationType } from "@/types"

const TYPE_LABELS: Record<NotificationType, string> = {
  status_change: "対応状況が変わったとき",
  assignee_assigned: "担当者が割り当てられたとき",
}

const DETAIL_LABELS: Record<NotificationType, string> = {
  status_change: "変更後の対応状況の名前をメールに含める",
  assignee_assigned: "担当者の名前をメールに含める",
}

const MODE_OPTIONS: { value: NotificationMode; label: string; description: string }[] = [
  {
    value: "always",
    label: "常時通知",
    description: "変更のたびに自動でメールを送信します",
  },
  {
    value: "confirm",
    label: "毎回確認",
    description: "変更のたびに送信するかどうかを確認します",
  },
  {
    value: "off",
    label: "通知しない",
    description: "自動送信も手動送信も行いません",
  },
]

type NotificationSettingsFieldsProps = {
  settings: NotificationSetting[]
  onChange: (notificationType: NotificationType, patch: Partial<NotificationSetting>) => void
  disabled: boolean
  emailSample: NotificationEmailSample
}

/** 通知の種類ごとの設定欄と、送信されるメールのプレビュー。 */
export function NotificationSettingsFields({
  settings,
  onChange,
  disabled,
  emailSample,
}: NotificationSettingsFieldsProps) {
  return (
    <>
      {settings.map((setting) => (
        <div key={setting.notification_type} className="space-y-3 rounded-lg border p-4">
          <h3 className="font-medium">{TYPE_LABELS[setting.notification_type]}</h3>

          <RadioGroup
            value={setting.mode}
            onValueChange={(value) =>
              onChange(setting.notification_type, { mode: value as NotificationMode })
            }
            disabled={disabled}
            className="space-y-2"
          >
            {MODE_OPTIONS.map((option) => {
              const id = `${setting.notification_type}-${option.value}`
              return (
                <div key={option.value} className="flex items-start gap-3">
                  <RadioGroupItem value={option.value} id={id} className="mt-1" />
                  <Label htmlFor={id} className="cursor-pointer font-normal">
                    <span className="block">{option.label}</span>
                    <span className="block text-xs text-muted-foreground">
                      {option.description}
                    </span>
                  </Label>
                </div>
              )
            })}
          </RadioGroup>

          <div className="flex items-center justify-between border-t pt-3">
            <Label
              htmlFor={`${setting.notification_type}-detail`}
              className="cursor-pointer font-normal"
            >
              {DETAIL_LABELS[setting.notification_type]}
            </Label>
            <Switch
              id={`${setting.notification_type}-detail`}
              checked={setting.include_detail}
              onCheckedChange={(checked) =>
                onChange(setting.notification_type, { include_detail: checked })
              }
              disabled={disabled || setting.mode === "off"}
            />
          </div>

          <EmailPreview
            notificationType={setting.notification_type}
            includeDetail={setting.include_detail}
            sample={emailSample}
            disabled={setting.mode === "off"}
          />
        </div>
      ))}
    </>
  )
}

type EmailPreviewProps = {
  notificationType: NotificationType
  includeDetail: boolean
  sample: NotificationEmailSample
  disabled: boolean
}

function EmailPreview({ notificationType, includeDetail, sample, disabled }: EmailPreviewProps) {
  const preview = renderNotificationEmailPreview(notificationType, includeDetail, sample)

  return (
    <div className="space-y-2 rounded-md border border-dashed bg-muted/30 p-3">
      <p className="text-xs font-medium">
        {disabled ? "通知を有効にすると送信されるメール" : "送信されるメール"}
      </p>
      <p className="text-xs break-all">
        <span className="text-muted-foreground">件名: </span>
        {preview.subject}
      </p>

      <div className="h-48 overflow-hidden rounded-sm bg-white">
        <iframe
          title="メール本文のプレビュー"
          srcDoc={preview.html}
          sandbox=""
          className="h-[133.33%] w-[133.33%] origin-top-left scale-75"
        />
      </div>

      <p className="text-xs text-muted-foreground">
        実際のメールは回答者のメールアドレスに送信されます。
      </p>
    </div>
  )
}
