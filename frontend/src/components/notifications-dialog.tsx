import { useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { AlertTriangle } from "lucide-react"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import { renderNotificationEmailPreview } from "@/lib/notification-email-preview"
import type { NotificationEmailSample } from "@/lib/notification-email-preview"
import { isEmailCollectionDisabled } from "@/hooks/use-notification-settings"
import type { NotificationMode, NotificationSetting, NotificationType } from "@/types"
import { formQueries } from "@/queries/forms"
import { queryKeys } from "@/queries/keys"

const EMPTY_SETTINGS: NotificationSetting[] = []

type NotificationsDialogProps = {
  formId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  canEdit: boolean
  emailSample: NotificationEmailSample
}

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

const errorMessages = {
  RESOURCE_HIDDEN: "フォームが見つからないか、アクセス権がありません",
} as const

export function NotificationsDialog({
  formId,
  open,
  onOpenChange,
  canEdit,
  emailSample,
}: NotificationsDialogProps) {
  const queryClient = useQueryClient()
  const {
    data,
    isPending: isLoading,
    error: loadError,
  } = useQuery({
    ...formQueries.notificationSettings(formId),
    enabled: open,
  })
  // 保存するまでの編集中の値。未編集なら取得した値をそのまま表示する。
  const [draft, setDraft] = useState<NotificationSetting[] | null>(null)
  const settings = draft ?? data?.settings ?? EMPTY_SETTINGS
  const emailCollectionType = data?.email_collection_type ?? null
  const [isWorking, setIsWorking] = useState(false)
  const [saveError, setSaveError] = useState("")
  const errorMessage =
    saveError ||
    (loadError ? getApiErrorMessage(loadError, errorMessages, "通知設定の取得に失敗しました") : "")

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      setDraft(null)
      setSaveError("")
    }
    onOpenChange(nextOpen)
  }

  const updateSetting = (
    notificationType: NotificationType,
    patch: Partial<NotificationSetting>,
  ) => {
    setDraft(
      settings.map((s) => (s.notification_type === notificationType ? { ...s, ...patch } : s)),
    )
  }

  const handleSave = async () => {
    setIsWorking(true)
    setSaveError("")
    try {
      const res = await apiClient.updateNotificationSettings(formId, { settings })
      // 回答一覧の確認ダイアログも同じキャッシュを読むため、保存した時点で反映される。
      queryClient.setQueryData(queryKeys.notificationSettings(formId), res)
      handleOpenChange(false)
    } catch (error) {
      setSaveError(getApiErrorMessage(error, errorMessages, "通知設定の保存に失敗しました"))
    } finally {
      setIsWorking(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>通知設定</DialogTitle>
          <DialogDescription>
            チケットが更新されたときに、回答者へメールで通知します。
          </DialogDescription>
        </DialogHeader>

        {isEmailCollectionDisabled(emailCollectionType) && (
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              このフォームは回答者のメールアドレスを収集していないため、通知は送信されません。
              Google フォーム側で収集を有効にしたあと、同期すると反映されます。
            </AlertDescription>
          </Alert>
        )}

        {errorMessage && (
          <Alert variant="destructive">
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        )}

        {isLoading ? (
          <p className="py-8 text-center text-sm text-muted-foreground">読み込み中...</p>
        ) : (
          <div className="space-y-6">
            {settings.map((setting) => (
              <div key={setting.notification_type} className="space-y-3 rounded-lg border p-4">
                <h3 className="font-medium">{TYPE_LABELS[setting.notification_type]}</h3>

                <RadioGroup
                  value={setting.mode}
                  onValueChange={(value) =>
                    updateSetting(setting.notification_type, { mode: value as NotificationMode })
                  }
                  disabled={!canEdit || isWorking}
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
                      updateSetting(setting.notification_type, { include_detail: checked })
                    }
                    disabled={!canEdit || isWorking || setting.mode === "off"}
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

            {canEdit ? (
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => handleOpenChange(false)}
                  disabled={isWorking}
                >
                  キャンセル
                </Button>
                <Button onClick={handleSave} disabled={isWorking}>
                  {isWorking ? "保存中..." : "保存"}
                </Button>
              </div>
            ) : (
              <p className="text-xs text-muted-foreground">
                通知設定を変更できるのは管理者のみです。
              </p>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
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
