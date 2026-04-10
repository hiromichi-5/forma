import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { AppLayout } from "@/components/app-layout";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
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
} from "@/components/ui/alert-dialog";
import { apiClient } from "@/lib/api";
import { getApiErrorMessage } from "@/lib/api-error";
import { toast } from "sonner";
import type { UserProfile } from "@/types";
import { Loader2 } from "lucide-react";

const sessionExpiredMessage =
  "セッションの有効期限が切れました。ログインし直してください";

export default function SettingsPage() {
  const navigate = useNavigate();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [isChangingPassword, setIsChangingPassword] = useState(false);
  const [isDeletingAccount, setIsDeletingAccount] = useState(false);

  const [displayName, setDisplayName] = useState("");

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  useEffect(() => {
    const loadProfile = async () => {
      try {
        setIsLoading(true);
        const data = await apiClient.getProfile();
        setProfile(data);
        setDisplayName(data.display_name);
      } catch (error) {
        console.error("Failed to load profile:", error);
        toast.error(
          getApiErrorMessage(
            error,
            {
              INVALID_SESSION: sessionExpiredMessage,
              USER_NOT_FOUND: "アカウントが見つかりません",
              NETWORK_ERROR: "ネットワークエラーが発生しました",
            },
            "プロフィールの読み込みに失敗しました"
          )
        );
      } finally {
        setIsLoading(false);
      }
    };
    loadProfile();
  }, []);

  const handleUpdateProfile = async () => {
    if (!displayName.trim()) {
      toast.error("表示名を入力してください");
      return;
    }

    try {
      setIsSavingProfile(true);
      const updatedProfile = await apiClient.updateProfile({
        display_name: displayName,
      });
      setProfile(updatedProfile);
      toast.success("プロフィールを更新しました");
    } catch (error) {
      console.error("Failed to update profile:", error);
      toast.error(
        getApiErrorMessage(
          error,
          {
            INVALID_SESSION: sessionExpiredMessage,
            VALIDATION_ERROR: "表示名を入力してください",
            USER_NOT_FOUND: "アカウントが見つかりません",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "プロフィールの更新に失敗しました"
        )
      );
    } finally {
      setIsSavingProfile(false);
    }
  };

  const handleChangePassword = async () => {
    if (!currentPassword || !newPassword || !confirmPassword) {
      toast.error("すべての項目を入力してください");
      return;
    }

    if (newPassword !== confirmPassword) {
      toast.error("新しいパスワードが一致しません");
      return;
    }

    if (newPassword.length < 8) {
      toast.error("パスワードは8文字以上で設定してください");
      return;
    }

    try {
      setIsChangingPassword(true);
      await apiClient.changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      toast.success("パスワードを変更しました");
    } catch (error) {
      console.error("Failed to change password:", error);
      toast.error(
        getApiErrorMessage(
          error,
          {
            INVALID_SESSION: sessionExpiredMessage,
            VALIDATION_ERROR: "入力内容を確認してください",
            INCORRECT_PASSWORD: "現在のパスワードが正しくありません",
            USER_NOT_FOUND: "アカウントが見つかりません",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "パスワードの変更に失敗しました"
        )
      );
    } finally {
      setIsChangingPassword(false);
    }
  };

  const handleDeleteAccount = async () => {
    try {
      setIsDeletingAccount(true);
      await apiClient.deleteProfile();
      toast.success("アカウントを削除しました");
      navigate("/login");
    } catch (error) {
      console.error("Failed to delete account:", error);
      toast.error(
        getApiErrorMessage(
          error,
          {
            INVALID_SESSION: sessionExpiredMessage,
            USER_NOT_FOUND: "アカウントが見つかりません",
            NETWORK_ERROR: "ネットワークエラーが発生しました",
          },
          "アカウントの削除に失敗しました"
        )
      );
    } finally {
      setIsDeletingAccount(false);
    }
  };

  if (isLoading) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="space-y-6 max-w-3xl">
        <div>
          <h1 className="text-2xl font-bold text-foreground">設定</h1>
          <p className="text-sm text-muted-foreground mt-1">
            アカウント情報・セキュリティ設定
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>プロフィール</CardTitle>
            <CardDescription>
              アカウントの基本情報を変更できます
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">メールアドレス</Label>
              <Input
                id="email"
                type="email"
                value={profile?.email || ""}
                disabled
                className="bg-muted"
              />
              <p className="text-xs text-muted-foreground">
                メールアドレスは変更できません
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="displayName">表示名</Label>
              <Input
                id="displayName"
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="表示名を入力"
              />
            </div>

            <Button
              onClick={handleUpdateProfile}
              disabled={
                isSavingProfile || displayName === profile?.display_name
              }
            >
              {isSavingProfile ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  保存中...
                </>
              ) : (
                "保存"
              )}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>パスワード変更</CardTitle>
            <CardDescription>
              アカウントのパスワードを変更できます
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="currentPassword">現在のパスワード</Label>
              <Input
                id="currentPassword"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="現在のパスワードを入力"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="newPassword">新しいパスワード</Label>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="新しいパスワードを入力（8文字以上）"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="confirmPassword">新しいパスワード（確認）</Label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="新しいパスワードを再入力"
              />
            </div>

            <Button
              onClick={handleChangePassword}
              disabled={isChangingPassword}
            >
              {isChangingPassword ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  変更中...
                </>
              ) : (
                "パスワードを変更"
              )}
            </Button>
          </CardContent>
        </Card>

        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle className="text-destructive">アカウント削除</CardTitle>
            <CardDescription>
              この操作は取り消すことができません
            </CardDescription>
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
  );
}
