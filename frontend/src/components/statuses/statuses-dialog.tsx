import { useMemo, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { getApiErrorMessage, type ApiErrorMessageMap } from "@/lib/api-error"
import { sortStatuses } from "@/lib/ticket-display"
import { formQueries } from "@/queries/forms"
import type { FormStatus } from "@/types"
import { ColorInput } from "./color-input"
import { colorValueOrNull, createColorState } from "./color-state"
import { StatusEditRow, StatusRow } from "./status-row"
import { useStatusMutations } from "./use-status-mutations"

const EMPTY_STATUSES: FormStatus[] = []

const statusErrorMessages: ApiErrorMessageMap = {
  RESOURCE_HIDDEN: "フォームまたはステータスが見つかりません",
  STATUS_CONFLICT: "ステータス名または表示順が重複しています",
}

type StatusesDialogProps = {
  formId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function StatusesDialog({ formId, open, onOpenChange }: StatusesDialogProps) {
  const { data: statuses = EMPTY_STATUSES } = useQuery(formQueries.statuses(formId))
  const sortedStatuses = useMemo(() => sortStatuses(statuses), [statuses])
  const { isWorking, create, update, remove, setDefault, swapOrder } = useStatusMutations(
    formId,
    statuses,
  )

  const [newName, setNewName] = useState("")
  const [newColor, setNewColor] = useState(() => createColorState(null))
  const [editingId, setEditingId] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<FormStatus | null>(null)
  const [deleteError, setDeleteError] = useState("")
  const [errorMessage, setErrorMessage] = useState("")

  /** 操作を実行し、失敗したら文言をダイアログに表示する。成功したら true を返す。 */
  const attempt = async (
    operation: () => Promise<void>,
    messages: ApiErrorMessageMap,
    fallback: string,
    showError: (message: string) => void = setErrorMessage,
  ) => {
    showError("")
    try {
      await operation()
      return true
    } catch (error) {
      console.error(fallback, error)
      showError(getApiErrorMessage(error, { ...statusErrorMessages, ...messages }, fallback))
      return false
    }
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && deleteTarget) return
    if (!nextOpen) {
      setEditingId(null)
      setNewName("")
      setNewColor(createColorState(null))
      setErrorMessage("")
    }
    onOpenChange(nextOpen)
  }

  const handleCreate = async () => {
    if (!newName.trim()) return
    const created = await attempt(
      () => create({ name: newName.trim(), color: colorValueOrNull(newColor) }),
      { VALIDATION_ERROR: "ステータス名と表示順を確認してください" },
      "ステータスの追加に失敗しました",
    )
    if (created) {
      setNewName("")
      setNewColor(createColorState(null))
    }
  }

  const handleUpdate = async (statusId: string, values: { name: string; color: string | null }) => {
    const updated = await attempt(
      () => update(statusId, values),
      { VALIDATION_ERROR: "ステータス名・色・表示順を確認してください" },
      "ステータスの更新に失敗しました",
    )
    if (updated) setEditingId(null)
  }

  const handleMove = (index: number, direction: "up" | "down") => {
    const target = sortedStatuses[direction === "up" ? index - 1 : index + 1]
    if (!target) return
    void attempt(() => swapOrder(sortedStatuses[index], target), {}, "表示順の更新に失敗しました")
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    if (deleteTarget.is_default) {
      setDeleteError("デフォルトステータスは削除できません")
      return
    }
    const deleted = await attempt(
      () => remove(deleteTarget.id),
      {
        VALIDATION_ERROR:
          "このステータスは削除できません。使用中のチケットがないか確認してください",
      },
      "ステータスの削除に失敗しました",
      setDeleteError,
    )
    if (deleted) setDeleteTarget(null)
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
                aria-label="新しいステータス名"
                className="flex-1"
              />
              <ColorInput value={newColor} onChange={setNewColor} />
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
                <h3 className="text-sm font-medium text-muted-foreground">ステータス一覧</h3>
                <span className="text-xs text-muted-foreground">{sortedStatuses.length}件</span>
              </div>
              <div className="max-h-[400px] space-y-1 overflow-y-auto">
                {sortedStatuses.map((status, index) => (
                  <div
                    key={status.id}
                    className="rounded-md border bg-card transition-colors hover:bg-accent/50"
                  >
                    {editingId === status.id ? (
                      <StatusEditRow
                        status={status}
                        disabled={isWorking}
                        onSave={(values) => handleUpdate(status.id, values)}
                        onCancel={() => setEditingId(null)}
                      />
                    ) : (
                      <StatusRow
                        status={status}
                        isFirst={index === 0}
                        isLast={index === sortedStatuses.length - 1}
                        disabled={isWorking}
                        onEdit={() => setEditingId(status.id)}
                        onSetDefault={() =>
                          void attempt(
                            () => setDefault(status.id),
                            {},
                            "デフォルトステータスの設定に失敗しました",
                          )
                        }
                        onMove={(direction) => handleMove(index, direction)}
                        onDelete={() => setDeleteTarget(status)}
                      />
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={deleteTarget !== null}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) {
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
            <Button onClick={handleDelete} disabled={isWorking} variant="destructive">
              削除する
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
