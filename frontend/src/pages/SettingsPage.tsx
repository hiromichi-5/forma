import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { Settings, User, Shield, Save, RefreshCw } from "lucide-react";
import { Button } from "../components/ui/Button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/Card";
import { Input } from "../components/ui/Input";
import { Label } from "../components/ui/Label";
import { useAuth } from "../hooks/useAuth";
import { ApiError } from "../lib/api";

export default function SettingsPage() {
  const {
    profile,
    isProfileLoading,
    refreshProfile,
    updateProfile,
    changePassword,
  } = useAuth();
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState(false);
  const [passwordError, setPasswordError] = useState<string | null>(null);

  useEffect(() => {
    if (profile?.display_name) {
      setDisplayName(profile.display_name);
    }
  }, [profile]);

  useEffect(() => {
    if (!profile && !isProfileLoading) {
      refreshProfile();
    }
  }, [profile, isProfileLoading, refreshProfile]);

  const handleSave = async () => {
    if (!displayName.trim()) {
      setError("表示名を入力してください");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await updateProfile({ display_name: displayName.trim() });
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message || "プロフィールの更新に失敗しました");
      } else {
        setError("プロフィールの更新に失敗しました");
      }
    } finally {
      setLoading(false);
    }
  };

  const resetPasswordState = () => {
    setCurrentPassword("");
    setNewPassword("");
    setConfirmPassword("");
  };

  const handlePasswordUpdate = async () => {
    if (!currentPassword || !newPassword || !confirmPassword) {
      setPasswordError("未入力のフィールドがあります");
      return;
    }
    if (newPassword.length < 8) {
      setPasswordError("新しいパスワードは8文字以上で入力してください");
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError("新しいパスワードと確認用パスワードが一致しません");
      return;
    }

    setPasswordLoading(true);
    setPasswordError(null);
    setPasswordSuccess(false);

    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setPasswordSuccess(true);
      resetPasswordState();
      setTimeout(() => setPasswordSuccess(false), 3000);
    } catch (err) {
      if (err instanceof ApiError) {
        setPasswordError(err.message || "パスワードの更新に失敗しました");
      } else {
        setPasswordError("パスワードの更新に失敗しました");
      }
    } finally {
      setPasswordLoading(false);
    }
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div>
        <h1 className="text-3xl font-bold flex items-center gap-3">
          <Settings className="h-8 w-8 text-blue-600" />
          設定
        </h1>
      </div>

      {success && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          className="bg-green-50 border border-green-200 rounded-lg p-4"
        >
          <div className="flex items-center gap-2 text-green-800">
            <Save className="h-5 w-5" />
            設定が保存されました
          </div>
        </motion.div>
      )}

      {error && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          className="bg-red-50 border border-red-200 rounded-lg p-4"
        >
          <div className="text-red-800">{error}</div>
        </motion.div>
      )}
      {passwordSuccess && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          className="bg-green-50 border border-green-200 rounded-lg p-4"
        >
          <div className="flex items-center gap-2 text-green-800">
            <Shield className="h-5 w-5" />
            パスワードを更新しました
          </div>
        </motion.div>
      )}

      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              プロフィール設定
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isProfileLoading ? (
              <div className="flex items-center justify-center py-8">
                <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
                <span className="ml-2 text-gray-600">
                  プロフィールを読み込み中...
                </span>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="display-name">表示名</Label>
                  <Input
                    id="display-name"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    placeholder="表示名を入力"
                  />
                </div>
                <div>
                  <Label htmlFor="email">メールアドレス</Label>
                  <Input
                    id="email"
                    type="email"
                    value={profile?.email || ""}
                    disabled
                  />
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              セキュリティ設定
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {passwordError && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-red-800">
                {passwordError}
              </div>
            )}
            <div>
              <Label htmlFor="current-password">現在のパスワード</Label>
              <Input
                id="current-password"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="現在のパスワードを入力"
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label htmlFor="new-password">新しいパスワード</Label>
                <Input
                  id="new-password"
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="新しいパスワードを入力"
                />
              </div>
              <div>
                <Label htmlFor="confirm-password">
                  新しいパスワード（確認）
                </Label>
                <Input
                  id="confirm-password"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="パスワードを再入力"
                />
              </div>
            </div>
            <Button
              variant="secondary"
              size="sm"
              onClick={handlePasswordUpdate}
              disabled={
                passwordLoading ||
                !currentPassword ||
                !newPassword ||
                !confirmPassword
              }
              className="flex items-center gap-2"
            >
              {passwordLoading ? (
                <RefreshCw className="h-4 w-4 animate-spin" />
              ) : (
                <Shield className="h-4 w-4" />
              )}
              {passwordLoading ? "更新中..." : "パスワードを更新"}
            </Button>
          </CardContent>
        </Card>

        <div className="flex justify-end">
          <Button
            onClick={handleSave}
            disabled={
              loading ||
              isProfileLoading ||
              !displayName.trim() ||
              displayName === profile?.display_name
            }
            className="flex items-center gap-2"
          >
            {loading ? (
              <RefreshCw className="h-4 w-4 animate-spin" />
            ) : (
              <Save className="h-4 w-4" />
            )}
            {loading ? "保存中..." : "設定を保存"}
          </Button>
        </div>
      </div>
    </div>
  );
}
