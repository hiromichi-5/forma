"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Plus } from "lucide-react";
import { apiClient } from "@/lib/api";
import { ApiError } from "@/lib/api";

type RegisterFormDialogProps = {
  onRegistered: () => Promise<void>;
};

export function RegisterFormDialog({ onRegistered }: RegisterFormDialogProps) {
  const [formUrl, setFormUrl] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [isOpen, setIsOpen] = useState(false);

  const handleClose = () => {
    if (isLoading) return;
    setFormUrl("");
    setErrorMessage("");
    setIsOpen(false);
  };

  const handleRegister = async () => {
    if (!formUrl.trim()) return;
    setIsLoading(true);
    setErrorMessage("");

    try {
      await apiClient.registerForm({ url: formUrl.trim() });
      await onRegistered();
      handleClose();
    } catch (error) {
      if (error instanceof ApiError && error.isValidationError) {
        setErrorMessage("フォームURLを確認してください");
      } else {
        setErrorMessage("フォーム連携に失敗しました");
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => (open ? setIsOpen(true) : handleClose())}
    >
      <DialogTrigger asChild>
        <Button className="gap-2">
          <Plus className="h-4 w-4" />
          フォーム連携
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>フォーム連携</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="formUrl">フォームURL</Label>
            <Input
              id="formUrl"
              placeholder="https://docs.google.com/forms/d/..."
              value={formUrl}
              onChange={(e) => setFormUrl(e.target.value)}
              disabled={isLoading}
            />
            <div className="space-y-1.5 text-xs text-muted-foreground">
              <p>連携する前に、以下の操作が必要です</p>
              <ol className="list-decimal list-inside space-y-1 ml-1">
                <li>Googleフォームの編集画面を開く</li>
                <li>
                  右上の「公開」ボタンの左にある共有ボタン（人型のアイコン）を選択する
                </li>
                <li>
                  以下のメールアドレスを「編集者」として追加する
                  <div className="mt-1 p-2 bg-muted rounded font-mono text-[10px] break-all">
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
          <Button
            onClick={handleRegister}
            disabled={!formUrl.trim() || isLoading}
            className="w-full"
          >
            {isLoading ? "連携中..." : "連携する"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
