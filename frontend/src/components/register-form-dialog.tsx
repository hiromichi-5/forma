import type React from "react"
import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import type { z } from "zod"
import { Button } from "@/components/ui/button"
import { TextField } from "@/components/text-field"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Plus } from "lucide-react"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"
import { queryKeys } from "@/queries/keys"
import { registerFormSchema } from "@/lib/form-schemas"

type RegisterFormDialogProps = {
  trigger?: React.ReactNode
}

export function RegisterFormDialog({ trigger }: RegisterFormDialogProps) {
  const queryClient = useQueryClient()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<z.infer<typeof registerFormSchema>>({
    resolver: zodResolver(registerFormSchema),
    defaultValues: { url: "" },
  })
  const [errorMessage, setErrorMessage] = useState("")
  const [isOpen, setIsOpen] = useState(false)
  const registerForm = useMutation({
    mutationFn: (url: string) => apiClient.registerForm({ url }),
    // サイドバーと一覧ページは同じキャッシュを読むため、ここで無効化すれば両方が更新される。
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.forms(), exact: true }),
  })
  const isLoading = registerForm.isPending

  const handleClose = () => {
    if (isLoading) return
    reset()
    setErrorMessage("")
    setIsOpen(false)
  }

  const onSubmit = async ({ url }: z.infer<typeof registerFormSchema>) => {
    setErrorMessage("")

    try {
      await registerForm.mutateAsync(url)
      reset()
      setIsOpen(false)
    } catch (error) {
      setErrorMessage(
        getApiErrorMessage(
          error,
          {
            VALIDATION_ERROR: "Googleフォームの編集画面のURLを入力してください",
            FORM_NOT_FOUND: "指定されたフォームが見つかりません",
            FORM_NOT_SHARED:
              "フォームがサービスアカウントに共有されていません。共有設定を確認してください",
            FORM_ALREADY_REGISTERED: "このフォームは既に登録されています",
          },
          "フォーム連携に失敗しました",
        ),
      )
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => (open ? setIsOpen(true) : handleClose())}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            フォーム連携
          </Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>フォーム連携</DialogTitle>
        </DialogHeader>
        <form className="space-y-4 py-2" noValidate onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-2">
            <TextField
              id="formUrl"
              label="フォームURL"
              placeholder="https://docs.google.com/forms/d/..."
              disabled={isLoading}
              error={errors.url}
              {...register("url")}
            />
            <div className="space-y-1.5 text-xs text-muted-foreground">
              <p>連携する前に、以下の操作が必要です</p>
              <ol className="ml-1 list-inside list-decimal space-y-1">
                <li>Googleフォームの編集画面を開く</li>
                <li>右上の「公開」ボタンの左にある共有ボタン（人型のアイコン）を選択する</li>
                <li>
                  以下のメールアドレスを「編集者」として追加する
                  <div className="mt-1 rounded bg-muted p-2 font-mono text-[10px] break-all">
                    forma-service@forma-470418.iam.gserviceaccount.com
                  </div>
                </li>
                <li>フォームの編集画面URLをここに貼り付ける</li>
              </ol>
            </div>
          </div>
          {errorMessage && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {errorMessage}
            </div>
          )}
          <Button type="submit" disabled={isLoading} className="w-full">
            {isLoading ? "連携中..." : "連携する"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
