import { Controller, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import type { z } from "zod"
import { Mail } from "lucide-react"
import { Button } from "@/components/ui/button"
import { FieldError } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { inviteSchema } from "@/lib/form-schemas"
import { ROLE_LABELS, toRoleValue } from "./role"

export type InviteValues = z.infer<typeof inviteSchema>

type InviteFormProps = {
  /** 招待を送る。成功したら true を返し、入力欄を空に戻す。 */
  onInvite: (values: InviteValues) => Promise<boolean>
}

export function InviteForm({ onInvite }: InviteFormProps) {
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<InviteValues>({
    resolver: zodResolver(inviteSchema),
    defaultValues: { email: "", role: "editor" },
  })

  const onSubmit = async (values: InviteValues) => {
    if (await onInvite(values)) reset()
  }

  return (
    <form className="space-y-2" noValidate onSubmit={handleSubmit(onSubmit)}>
      <div className="flex gap-2">
        <div className="flex-1 space-y-2">
          <Label htmlFor="inviteEmail" className="text-xs">
            メールアドレス
          </Label>
          <Input
            id="inviteEmail"
            type="email"
            placeholder="example@company.com"
            aria-invalid={!!errors.email}
            {...register("email")}
          />
        </div>
        {/* フォーム内の Select は送信用の隠れた select を追加するため、space-y ではなく gap で間隔を空ける。 */}
        <div className="flex w-[140px] flex-col gap-2">
          <Label htmlFor="inviteRole" className="text-xs">
            権限
          </Label>
          <Controller
            control={control}
            name="role"
            render={({ field }) => (
              <Select value={field.value} onValueChange={(v) => field.onChange(toRoleValue(v))}>
                <SelectTrigger id="inviteRole" className="w-full bg-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">{ROLE_LABELS.admin}</SelectItem>
                  <SelectItem value="editor">{ROLE_LABELS.editor}</SelectItem>
                </SelectContent>
              </Select>
            )}
          />
        </div>
        <div className="flex items-end">
          <Button type="submit" className="gap-2" disabled={isSubmitting}>
            <Mail className="h-4 w-4" />
            招待
          </Button>
        </div>
      </div>
      <FieldError errors={[errors.email]} />
    </form>
  )
}
