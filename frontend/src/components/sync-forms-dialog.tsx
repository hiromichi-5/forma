"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { RefreshCw, CloudDownload } from "lucide-react"

export function SyncFormsDialog() {
  const [formId, setFormId] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [result, setResult] = useState<{ success: boolean; count?: number; error?: string } | null>(null)

  const handleSync = async () => {
    if (!formId.trim()) return

    setIsLoading(true)
    setResult(null)

    try {
      const response = await fetch("/api/sync-forms", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ formId: formId.trim() }),
      })

      const data = await response.json()

      if (response.ok) {
        setResult({ success: true, count: data.count })
      } else {
        setResult({ success: false, error: data.error })
      }
    } catch (error) {
      console.error("[v0] Sync error:", error)
      setResult({ success: false, error: "同期中にエラーが発生しました" })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" className="gap-2 bg-transparent">
          <CloudDownload className="h-4 w-4" />
          Googleフォーム連携
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Googleフォームと連携</DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="formId">フォームID</Label>
            <Input
              id="formId"
              placeholder="例: 1FAIpQLSe..."
              value={formId}
              onChange={(e) => setFormId(e.target.value)}
              disabled={isLoading}
            />
            <p className="text-xs text-muted-foreground">
              GoogleフォームのURLから取得できます。
              <br />
              例: https://docs.google.com/forms/d/[フォームID]/edit
            </p>
          </div>

          {result && (
            <div
              className={`p-3 rounded-lg ${result.success ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"}`}
            >
              {result.success ? (
                <p className="text-sm font-medium">{result.count}件の回答を同期しました</p>
              ) : (
                <p className="text-sm font-medium">{result.error}</p>
              )}
            </div>
          )}

          <Button onClick={handleSync} disabled={!formId.trim() || isLoading} className="w-full gap-2">
            {isLoading ? (
              <>
                <RefreshCw className="h-4 w-4 animate-spin" />
                同期中...
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4" />
                回答を同期
              </>
            )}
          </Button>

          <div className="bg-muted p-3 rounded-lg text-sm space-y-2">
            <p className="font-medium">設定が必要な環境変数:</p>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li>GOOGLE_SERVICE_ACCOUNT_EMAIL</li>
              <li>GOOGLE_SERVICE_ACCOUNT_PRIVATE_KEY</li>
            </ul>
            <p className="text-xs text-muted-foreground mt-2">
              Service Accountをフォームの編集者として追加してください。
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
