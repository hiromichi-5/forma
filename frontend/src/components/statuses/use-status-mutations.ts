import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { apiClient } from "@/lib/api"
import { queryKeys } from "@/queries/keys"
import type { FormStatus } from "@/types"

/**
 * ステータスの変更操作。結果は API の応答で確定しているため、取り直さずにキャッシュへ直接書き込む。
 * 一覧やカンバンも同じキャッシュを読むので、ダイアログを閉じる前から反映される。
 * 失敗した場合は例外をそのまま投げるので、呼び出し側で文言を選んで表示する。
 */
export function useStatusMutations(formId: string, statuses: FormStatus[]) {
  const queryClient = useQueryClient()
  const [isWorking, setIsWorking] = useState(false)

  const save = (next: FormStatus[]) =>
    queryClient.setQueryData(queryKeys.formStatuses(formId), { statuses: next })

  const run = async (operation: () => Promise<void>) => {
    setIsWorking(true)
    try {
      await operation()
    } finally {
      setIsWorking(false)
    }
  }

  const create = (values: { name: string; color: string | null }) =>
    run(async () => {
      const maxOrder = statuses.reduce((max, s) => Math.max(max, s.display_order), 0)
      const created = await apiClient.createFormStatus(formId, {
        ...values,
        display_order: maxOrder + 1,
      })
      save([...statuses, created])
    })

  const update = (statusId: string, values: { name: string; color: string | null }) =>
    run(async () => {
      const updated = await apiClient.updateFormStatus(formId, statusId, values)
      save(statuses.map((s) => (s.id === statusId ? updated : s)))
    })

  const remove = (statusId: string) =>
    run(async () => {
      await apiClient.deleteFormStatus(formId, statusId)
      save(statuses.filter((s) => s.id !== statusId))
    })

  const setDefault = (statusId: string) =>
    run(async () => {
      const updated = await apiClient.updateFormStatus(formId, statusId, { is_default: true })
      save(statuses.map((s) => ({ ...s, is_default: s.id === updated.id })))
    })

  /**
   * 隣のステータスと表示順を入れ替える。表示順は重複できないため、
   * 一方をいったん使われていない値に退避してから入れ替える。
   */
  const swapOrder = (current: FormStatus, target: FormStatus) =>
    run(async () => {
      const tempOrder = Math.max(...statuses.map((s) => s.display_order), 0) + 1000
      await apiClient.updateFormStatus(formId, current.id, { display_order: tempOrder })
      const updatedTarget = await apiClient.updateFormStatus(formId, target.id, {
        display_order: current.display_order,
      })
      const updatedCurrent = await apiClient.updateFormStatus(formId, current.id, {
        display_order: target.display_order,
      })
      save(
        statuses.map((s) =>
          s.id === updatedCurrent.id
            ? updatedCurrent
            : s.id === updatedTarget.id
              ? updatedTarget
              : s,
        ),
      )
    })

  return { isWorking, create, update, remove, setDefault, swapOrder }
}
