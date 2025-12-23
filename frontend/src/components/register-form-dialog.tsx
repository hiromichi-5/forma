"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Plus } from "lucide-react"
import { apiClient } from "@/lib/api"
import { ApiError } from "@/lib/api"

type RegisterFormDialogProps = {
  onRegistered: () => Promise<void>
}

export function RegisterFormDialog({ onRegistered }: RegisterFormDialogProps) {
  const [formUrl, setFormUrl] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [errorMessage, setErrorMessage] = useState("")
  const [isOpen, setIsOpen] = useState(false)

  const handleClose = () => {
    if (isLoading) return
    setFormUrl("")
    setErrorMessage("")
    setIsOpen(false)
  }

  const handleRegister = async () => {
    if (!formUrl.trim()) return
    setIsLoading(true)
    setErrorMessage("")

    try {
      await apiClient.registerForm({ url: formUrl.trim() })
      await onRegistered()
      handleClose()
    } catch (error) {
      if (error instanceof ApiError && error.isValidationError) {
        setErrorMessage("フォームURLを確認してください")
      } else {
        setErrorMessage("フォーム連携に失敗しました")
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => (open ? setIsOpen(true) : handleClose())}>
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
            <p className="text-xs text-muted-foreground">
              Googleフォームの編集画面URLを貼り付けてください。
            </p>
          </div>
          {errorMessage && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">{errorMessage}</div>
          )}
          <Button onClick={handleRegister} disabled={!formUrl.trim() || isLoading} className="w-full">
            {isLoading ? "連携中..." : "連携する"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
