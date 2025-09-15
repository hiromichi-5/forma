import { useState } from "react";
import { motion } from "framer-motion";
import {
  Settings,
  User,
  Bell,
  Shield,
  Palette,
  Globe,
  Save,
  RefreshCw,
} from "lucide-react";
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

export default function SettingsPage() {
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const handleSave = async () => {
    setLoading(true);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000));
    setLoading(false);
    setSuccess(true);
    setTimeout(() => setSuccess(false), 3000);
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <Settings className="h-8 w-8 text-blue-600" />
          設定
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          アカウントとアプリケーションの設定を管理
        </p>
      </div>

      {/* Success Message */}
      {success && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          className="bg-green-50 dark:bg-green-900/50 border border-green-200 dark:border-green-800 rounded-lg p-4"
        >
          <div className="flex items-center gap-2 text-green-800 dark:text-green-200">
            <Save className="h-5 w-5" />
            設定が保存されました
          </div>
        </motion.div>
      )}

      <div className="grid gap-6">
        {/* Profile Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              プロフィール設定
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label htmlFor="display-name">表示名</Label>
                <Input
                  id="display-name"
                  defaultValue={user?.email?.split("@")[0] || ""}
                  placeholder="表示名を入力"
                />
              </div>
              <div>
                <Label htmlFor="email">メールアドレス</Label>
                <Input
                  id="email"
                  type="email"
                  defaultValue={user?.email || ""}
                  disabled
                  className="bg-gray-50 dark:bg-gray-800"
                />
              </div>
            </div>
            <div>
              <Label htmlFor="bio">自己紹介</Label>
              <Input id="bio" placeholder="自己紹介を入力" />
            </div>
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              通知設定
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label className="text-base">新規チケット通知</Label>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  新しいチケットが作成されたときに通知を受け取る
                </p>
              </div>
              <input type="checkbox" defaultChecked className="toggle" />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <Label className="text-base">フォーム同期通知</Label>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  フォームが同期されたときに通知を受け取る
                </p>
              </div>
              <input type="checkbox" defaultChecked className="toggle" />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <Label className="text-base">メンバー変更通知</Label>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  メンバーが追加・削除されたときに通知を受け取る
                </p>
              </div>
              <input type="checkbox" className="toggle" />
            </div>
          </CardContent>
        </Card>

        {/* Security Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              セキュリティ設定
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="current-password">現在のパスワード</Label>
              <Input
                id="current-password"
                type="password"
                placeholder="現在のパスワードを入力"
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label htmlFor="new-password">新しいパスワード</Label>
                <Input
                  id="new-password"
                  type="password"
                  placeholder="新しいパスワードを入力"
                />
              </div>
              <div>
                <Label htmlFor="confirm-password">パスワード確認</Label>
                <Input
                  id="confirm-password"
                  type="password"
                  placeholder="パスワードを再入力"
                />
              </div>
            </div>
            <Button variant="secondary" size="sm">
              パスワードを更新
            </Button>
          </CardContent>
        </Card>

        {/* Appearance Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Palette className="h-5 w-5" />
              外観設定
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label>テーマ</Label>
              <div className="flex items-center gap-4 mt-2">
                <label className="flex items-center gap-2">
                  <input
                    type="radio"
                    name="theme"
                    value="light"
                    defaultChecked
                  />
                  <span className="text-sm">ライト</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="radio" name="theme" value="dark" />
                  <span className="text-sm">ダーク</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="radio" name="theme" value="auto" />
                  <span className="text-sm">システム設定に従う</span>
                </label>
              </div>
            </div>
            <div>
              <Label>言語</Label>
              <div className="flex items-center gap-2 mt-2">
                <Globe className="h-4 w-4 text-gray-400" />
                <span className="text-sm">日本語</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Save Button */}
        <div className="flex justify-end">
          <Button
            onClick={handleSave}
            disabled={loading}
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
