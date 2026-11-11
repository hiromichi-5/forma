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
import { Alert, AlertDescription } from "@/components/ui/alert"
import { AlertTriangle } from "lucide-react"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import type { NotificationEmailSample } from "@/lib/notification-email-preview"
import { isEmailCollectionDisabled } from "@/hooks/use-notification-settings"
import { NotificationSettingsFields } from "@/components/notification-settings-fields"
import type { NotificationSetting, NotificationType } from "@/types"
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
            <NotificationSettingsFields
              settings={settings}
              onChange={updateSetting}
              disabled={!canEdit || isWorking}
              emailSample={emailSample}
            />

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
