"use client"

import { useMemo, useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { ArrowDown, ArrowUp, Check, Pencil, Plus, Star, Trash2, X } from "lucide-react"
import type { FormStatus } from "@/types"
import { apiClient } from "@/lib/api"
import { getApiErrorMessage } from "@/lib/api-error"

const colorSwitchTooltip = "ONにするとステータスに表示色を設定できます"

type StatusesDialogProps = {
  formId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  statuses: FormStatus[]
  onStatusesChange: (statuses: FormStatus[]) => void
}

type ColorInputState = {
  enabled: boolean
  value: string
}

const defaultColor = "#3B82F6"

const statusColorPresets = [
  { name: "レッド", value: "#EF4444" },
  { name: "オレンジ", value: "#F97316" },
  { name: "アンバー", value: "#F59E0B" },
  { name: "イエロー", value: "#EAB308" },
  { name: "ライム", value: "#84CC16" },
  { name: "グリーン", value: "#22C55E" },
  { name: "エメラルド", value: "#10B981" },
  { name: "ティール", value: "#14B8A6" },
  { name: "シアン", value: "#06B6D4" },
  { name: "スカイ", value: "#0EA5E9" },
  { name: "ブルー", value: "#3B82F6" },
  { name: "インディゴ", value: "#6366F1" },
  { name: "バイオレット", value: "#8B5CF6" },
  { name: "パープル", value: "#A855F7" },
  { name: "ピンク", value: "#EC4899" },
  { name: "グレー", value: "#6B7280" },
]

const createColorState = (color?: string | null): ColorInputState => ({
  enabled: Boolean(color),
  value: color ?? defaultColor,
})

const colorValueOrNull = (state: ColorInputState): string | null =>
  state.enabled ? state.value : null

type ColorSwatchPickerProps = {
  value: string
  disabled?: boolean
  onChange: (color: string) => void
  className?: string
}

function ColorSwatchPicker({
  value,
  disabled,
  onChange,
  className,
}: ColorSwatchPickerProps) {
  const [open, setOpen] = useState(false)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          disabled={disabled}
          aria-label="色を選択"
          className={`rounded-md border border-border disabled:opacity-40 disabled:cursor-not-allowed ${className ?? "h-7 w-7"
            }`}
          style={{ backgroundColor: value }}
        />
      </PopoverTrigger>
      <PopoverContent className="w-48 p-2" align="start">
        <div className="grid grid-cols-6 gap-1">
          {statusColorPresets.map((preset) => (
            <button
              key={preset.value}
              type="button"
              title={preset.name}
              aria-label={preset.name}
              onClick={() => {
                onChange(preset.value)
                setOpen(false)
              }}
              className="relative h-5 w-5 rounded-full border border-border/50"
              style={{ backgroundColor: preset.value }}
            >
              {value === preset.value && (
                <Check className="absolute inset-0 m-auto h-3 w-3 text-white drop-shadow" />
              )}
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}

const defaultStatusErrorMessages = {
  RESOURCE_HIDDEN: "フォームまたはステータスが見つかりません",
  STATUS_CONFLICT: "ステータス名または表示順が重複しています",
  INVALID_SESSION: "セッションの有効期限が切れました。ログインし直してください",
  NETWORK_ERROR: "ネットワークエラーが発生しました",
} as const

export function StatusesDialog({
  formId,
  open,
  onOpenChange,
  statuses,
  onStatusesChange,
}: StatusesDialogProps) {
  const [newName, setNewName] = useState("")
  const [newColor, setNewColor] = useState<ColorInputState>({
    enabled: false,
    value: defaultColor,
  })
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState("")
  const [editColor, setEditColor] = useState<ColorInputState>({
    enabled: false,
    value: defaultColor,
  })
  const [deleteTarget, setDeleteTarget] = useState<FormStatus | null>(null)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const [isWorking, setIsWorking] = useState(false)
  const [errorMessage, setErrorMessage] = useState("")

  const sortedStatuses = useMemo(
    () => [...statuses].sort((a, b) => a.display_order - b.display_order),
    [statuses]
  )

  const resetNewFields = () => {
    setNewName("")
    setNewColor({ enabled: false, value: defaultColor })
  }

  const startEdit = (status: FormStatus) => {
    setEditingId(status.id)
    setEditName(status.name)
    setEditColor(createColorState(status.color))
  }

  const cancelEdit = () => {
    setEditingId(null)
    setEditName("")
    setEditColor({ enabled: false, value: defaultColor })
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && isDeleteDialogOpen) {
      return
    }
    if (!nextOpen) {
      cancelEdit()
      resetNewFields()
      setDeleteTarget(null)
      setDeleteError("")
      setErrorMessage("")
    }
    onOpenChange(nextOpen)
  }

  const handleCreate = async () => {
    if (!newName.trim()) return
    setIsWorking(true)
    setErrorMessage("")
    try {
      const maxOrder = statuses.reduce(
        (max, status) => Math.max(max, status.display_order),
        0
      )
      const created = await apiClient.createFormStatus(formId, {
        name: newName.trim(),
        color: colorValueOrNull(newColor),
        display_order: maxOrder + 1,
      })
      onStatusesChange([...statuses, created])
      resetNewFields()
    } catch (error) {
      console.error("Failed to create status:", error)
      setErrorMessage(
        getApiErrorMessage(
          error,
          {
            ...defaultStatusErrorMessages,
            VALIDATION_ERROR: "ステータス名と表示順を確認してください",
          },
          "ステータスの追加に失敗しました"
        )
      )
    } finally {
      setIsWorking(false)
    }
  }

  const handleUpdate = async (statusId: string) => {
    if (!editName.trim()) return
    setIsWorking(true)
    setErrorMessage("")
    try {
      const updated = await apiClient.updateFormStatus(formId, statusId, {
        name: editName.trim(),
        color: colorValueOrNull(editColor),
      })
      onStatusesChange(
        statuses.map((status) => (status.id === statusId ? updated : status))
      )
      cancelEdit()
    } catch (error) {
      console.error("Failed to update status:", error)
      setErrorMessage(
        getApiErrorMessage(
          error,
          {
            ...defaultStatusErrorMessages,
            VALIDATION_ERROR: "ステータス名・色・表示順を確認してください",
          },
          "ステータスの更新に失敗しました"
        )
      )
    } finally {
      setIsWorking(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    if (deleteTarget.is_default) {
      setDeleteError("デフォルトステータスは削除できません")
      return
    }
    setIsWorking(true)
    setDeleteError("")
    try {
      await apiClient.deleteFormStatus(formId, deleteTarget.id)
      onStatusesChange(
        statuses.filter((status) => status.id !== deleteTarget.id)
      )
      setIsDeleteDialogOpen(false)
      setDeleteTarget(null)
    } catch (error) {
      console.error("Failed to delete status:", error)
      setDeleteError(
        getApiErrorMessage(
          error,
          {
            ...defaultStatusErrorMessages,
            VALIDATION_ERROR:
              "このステータスは削除できません。使用中のチケットがないか確認してください",
          },
          "ステータスの削除に失敗しました"
        )
      )
    } finally {
      setIsWorking(false)
    }
  }

  const handleSetDefault = async (statusId: string) => {
    setIsWorking(true)
    setErrorMessage("")
    try {
      const updatedDefault = await apiClient.updateFormStatus(
        formId,
        statusId,
        { is_default: true }
      )
      onStatusesChange(
        statuses.map((status) => ({
          ...status,
          is_default: status.id === updatedDefault.id,
        }))
      )
    } catch (error) {
      console.error("Failed to set default status:", error)
      setErrorMessage(
        getApiErrorMessage(
          error,
          defaultStatusErrorMessages,
          "デフォルトステータスの設定に失敗しました"
        )
      )
    } finally {
      setIsWorking(false)
    }
  }

  const handleMove = async (statusId: string, direction: "up" | "down") => {
    const index = sortedStatuses.findIndex((status) => status.id === statusId)
    const targetIndex = direction === "up" ? index - 1 : index + 1
    if (index < 0 || targetIndex < 0 || targetIndex >= sortedStatuses.length) {
      return
    }

    const current = sortedStatuses[index]
    const target = sortedStatuses[targetIndex]
    setIsWorking(true)
    setErrorMessage("")
    try {
      const tempOrderBase = Math.max(
        ...sortedStatuses.map((status) => status.display_order),
        0
      ) + 1000
      const tempOrder = tempOrderBase + index

      await apiClient.updateFormStatus(formId, current.id, {
        display_order: tempOrder,
      })
      const updatedTarget = await apiClient.updateFormStatus(formId, target.id, {
        display_order: current.display_order,
      })
      const updatedCurrent = await apiClient.updateFormStatus(formId, current.id, {
        display_order: target.display_order,
      })
      onStatusesChange(
        statuses.map((status) => {
          if (status.id === updatedCurrent.id) return updatedCurrent
          if (status.id === updatedTarget.id) return updatedTarget
          return status
        })
      )
    } catch (error) {
      console.error("Failed to reorder status:", error)
      setErrorMessage(
        getApiErrorMessage(
          error,
          defaultStatusErrorMessages,
          "表示順の更新に失敗しました"
        )
      )
    } finally {
      setIsWorking(false)
    }
  }

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>ステータス管理</DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="新しいステータス名"
                className="flex-1"
              />
              <div className="flex items-center gap-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className="inline-flex">
                      <Switch
                        checked={newColor.enabled}
                        onCheckedChange={(checked) =>
                          setNewColor((prev) => ({ ...prev, enabled: checked }))
                        }
                      />
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>{colorSwitchTooltip}</TooltipContent>
                </Tooltip>
                <ColorSwatchPicker
                  value={newColor.value}
                  disabled={!newColor.enabled}
                  onChange={(value) =>
                    setNewColor((prev) => ({ ...prev, value }))
                  }
                  className="h-7 w-7"
                />
              </div>
              <Button
                onClick={handleCreate}
                size="sm"
                className="gap-1.5"
                disabled={isWorking || !newName.trim()}
              >
                <Plus className="h-4 w-4" />
                追加する
              </Button>
            </div>

            {errorMessage && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {errorMessage}
              </div>
            )}

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium text-muted-foreground">
                  ステータス一覧
                </h3>
                <span className="text-xs text-muted-foreground">
                  {sortedStatuses.length}件
                </span>
              </div>
              <div className="space-y-1 max-h-[400px] overflow-y-auto">
                {sortedStatuses.map((status, index) => {
                  const isEditing = editingId === status.id
                  return (
                    <div
                      key={status.id}
                      className="rounded-md border bg-card hover:bg-accent/50 transition-colors"
                    >
                      {!isEditing ? (
                        <div className="flex items-center gap-2 p-2">
                          <div
                            className="flex items-center gap-2 flex-1 min-w-0 cursor-pointer select-none"
                            onClick={() => startEdit(status)}
                          >
                            <div
                              className="h-3 w-3 rounded-full flex-shrink-0"
                              style={{
                                backgroundColor: status.color ?? "#9CA3AF",
                              }}
                            />
                            <span className="text-sm flex-1 min-w-0 truncate">
                              {status.name}
                            </span>
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 flex-shrink-0"
                            disabled={isWorking}
                            onClick={() =>
                              status.is_default
                                ? undefined
                                : handleSetDefault(status.id)
                            }
                          >
                            <Star
                              className={`h-3.5 w-3.5 ${status.is_default
                                ? "fill-yellow-400 text-yellow-400"
                                : "text-muted-foreground"
                                }`}
                            />
                          </Button>
                          <div className="flex items-center gap-0.5 flex-shrink-0">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              disabled={isWorking || index === 0}
                              onClick={() => handleMove(status.id, "up")}
                            >
                              <ArrowUp className="h-3.5 w-3.5" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              disabled={
                                isWorking || index === sortedStatuses.length - 1
                              }
                              onClick={() => handleMove(status.id, "down")}
                            >
                              <ArrowDown className="h-3.5 w-3.5" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              disabled={isWorking}
                              onClick={() => startEdit(status)}
                            >
                              <Pencil className="h-3.5 w-3.5" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              disabled={isWorking || status.is_default}
                              onClick={() => {
                                setDeleteTarget(status)
                                setIsDeleteDialogOpen(true)
                              }}
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        </div>
                      ) : (
                        <div className="p-2 space-y-2">
                          <div className="flex items-center gap-2">
                            <Input
                              value={editName}
                              onChange={(e) => setEditName(e.target.value)}
                              className="flex-1 h-8"
                              placeholder="ステータス名"
                            />
                            <div className="flex items-center gap-1">
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <span className="inline-flex">
                                    <Switch
                                      checked={editColor.enabled}
                                      onCheckedChange={(checked) =>
                                        setEditColor((prev) => ({
                                          ...prev,
                                          enabled: checked,
                                        }))
                                      }
                                    />
                                  </span>
                                </TooltipTrigger>
                                <TooltipContent>
                                  {colorSwitchTooltip}
                                </TooltipContent>
                              </Tooltip>
                              <ColorSwatchPicker
                                value={editColor.value}
                                disabled={!editColor.enabled}
                                onChange={(value) =>
                                  setEditColor((prev) => ({ ...prev, value }))
                                }
                                className="h-6 w-6"
                              />
                            </div>
                          </div>
                          <div className="flex gap-1.5">
                            <Button
                              size="sm"
                              onClick={() => handleUpdate(status.id)}
                              disabled={isWorking || !editName.trim()}
                              className="h-7 gap-1"
                            >
                              <Check className="h-3.5 w-3.5" />
                              保存
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={cancelEdit}
                              disabled={isWorking}
                              className="h-7 gap-1"
                            >
                              <X className="h-3.5 w-3.5" />
                              キャンセル
                            </Button>
                          </div>
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={isDeleteDialogOpen}
        onOpenChange={(open) => {
          if (!open) {
            setIsDeleteDialogOpen(false)
            setDeleteTarget(null)
            setDeleteError("")
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>ステータスを削除しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              「{deleteTarget?.name}」を削除すると元に戻せません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {deleteError && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {deleteError}
            </div>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isWorking}>キャンセル</AlertDialogCancel>
            <Button
              onClick={handleDelete}
              disabled={isWorking}
              variant="destructive"
            >
              削除する
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
