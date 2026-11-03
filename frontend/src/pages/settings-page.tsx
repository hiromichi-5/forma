import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import type { z } from "zod"
import { useNavigate } from "react-router"
import { AppLayout } from "@/components/app-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { TextField } from "@/components/text-field"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { useAuth } from "@/hooks/use-auth"
import { changePasswordSchema, PASSWORD_MIN_LENGTH, profileSchema } from "@/lib/form-schemas"
import { toastApiError } from "@/lib/api-error"
import { toast } from "sonner"
import { Loader2 } from "lucide-react"

export default function SettingsPage() {
  const navigate = useNavigate()
  const { profile, deleteAccount } = useAuth()
  const [isDeletingAccount, setIsDeletingAccount] = useState(false)

  const handleDeleteAccount = async () => {
    try {
      setIsDeletingAccount(true)
      // ログアウトまで行い、認証状態とキャッシュを破棄してからログイン画面へ戻す。
      await deleteAccount()
      toast.success("アカウントを削除しました")
      void navigate("/login")
    } catch (error) {
      console.error("Failed to delete account:", error)
      toastApiError(
        error,
        {
          USER_NOT_FOUND: "アカウントが見つかりません",
        },
        "アカウントの削除に失敗しました",
      )
    } finally {
      setIsDeletingAccount(false)
    }
  }

  return (
    <AppLayout>
      <div className="max-w-3xl space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">設定</h1>
          <p className="mt-1 text-sm text-muted-foreground">アカウント情報・セキュリティ設定</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>プロフィール</CardTitle>
            <CardDescription>アカウントの基本情報を変更できます</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <TextField
              id="email"
              label="メールアドレス"
              type="email"
              value={profile?.email ?? ""}
              disabled
              className="bg-muted"
              description="メールアドレスは変更できません"
            />
            {profile && <ProfileForm key={profile.id} displayName={profile.display_name} />}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>パスワード変更</CardTitle>
            <CardDescription>アカウントのパスワードを変更できます</CardDescription>
          </CardHeader>
          <CardContent>
            <PasswordForm />
          </CardContent>
        </Card>

        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle className="text-destructive">アカウント削除</CardTitle>
            <CardDescription>この操作は取り消すことができません</CardDescription>
          </CardHeader>
          <CardContent>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive" disabled={isDeletingAccount}>
                  {isDeletingAccount ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      削除中...
                    </>
                  ) : (
                    "アカウントを削除"
                  )}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>本当に削除しますか？</AlertDialogTitle>
                  <AlertDialogDescription>
                    この操作は取り消せません。アカウントとすべてのデータが完全に削除されます。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>キャンセル</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={handleDeleteAccount}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    削除する
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}

function ProfileForm({ displayName }: { displayName: string }) {
  const { updateProfile } = useAuth()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting, isDirty },
  } = useForm<z.infer<typeof profileSchema>>({
    resolver: zodResolver(profileSchema),
    defaultValues: { display_name: displayName },
  })

  const onSubmit = async (values: z.infer<typeof profileSchema>) => {
    try {
      // AuthProvider 経由で更新し、画面全体で使うプロフィールも最新にする。
      const updated = await updateProfile(values)
      reset({ display_name: updated.display_name })
      toast.success("プロフィールを更新しました")
    } catch (error) {
      console.error("Failed to update profile:", error)
      toastApiError(
        error,
        {
          VALIDATION_ERROR: "表示名を入力してください",
          USER_NOT_FOUND: "アカウントが見つかりません",
        },
        "プロフィールの更新に失敗しました",
      )
    }
  }

  return (
    <form className="space-y-4" noValidate onSubmit={handleSubmit(onSubmit)}>
      <TextField
        id="displayName"
        label="表示名"
        autoComplete="name"
        placeholder="表示名を入力"
        error={errors.display_name}
        {...register("display_name")}
      />
      <Button type="submit" disabled={isSubmitting || !isDirty}>
        {isSubmitting ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            保存中...
          </>
        ) : (
          "保存"
        )}
      </Button>
    </form>
  )
}

function PasswordForm() {
  const { changePassword } = useAuth()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<z.infer<typeof changePasswordSchema>>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: { current_password: "", new_password: "", confirm_password: "" },
  })

  const onSubmit = async ({
    current_password,
    new_password,
  }: z.infer<typeof changePasswordSchema>) => {
    try {
      await changePassword({ current_password, new_password })
      reset()
      toast.success("パスワードを変更しました")
    } catch (error) {
      console.error("Failed to change password:", error)
      toastApiError(
        error,
        {
          VALIDATION_ERROR: "入力内容を確認してください",
          INCORRECT_PASSWORD: "現在のパスワードが正しくありません",
          USER_NOT_FOUND: "アカウントが見つかりません",
        },
        "パスワードの変更に失敗しました",
      )
    }
  }

  return (
    <form className="space-y-4" noValidate onSubmit={handleSubmit(onSubmit)}>
      <TextField
        id="currentPassword"
        label="現在のパスワード"
        type="password"
        autoComplete="current-password"
        placeholder="現在のパスワードを入力"
        error={errors.current_password}
        {...register("current_password")}
      />
      <TextField
        id="newPassword"
        label="新しいパスワード"
        type="password"
        autoComplete="new-password"
        placeholder={`新しいパスワードを入力（${PASSWORD_MIN_LENGTH}文字以上）`}
        error={errors.new_password}
        {...register("new_password")}
      />
      <TextField
        id="confirmPassword"
        label="新しいパスワード（確認）"
        type="password"
        autoComplete="new-password"
        placeholder="新しいパスワードを再入力"
        error={errors.confirm_password}
        {...register("confirm_password")}
      />
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            変更中...
          </>
        ) : (
          "パスワードを変更"
        )}
      </Button>
    </form>
  )
}
