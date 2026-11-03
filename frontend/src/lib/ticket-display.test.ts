import { describe, expect, it } from "vitest"
import {
  hexToRgba,
  respondentEmailLabel,
  sortStatuses,
  toPriorityValue,
} from "@/lib/ticket-display"
import type { FormStatus } from "@/types"

describe("toPriorityValue", () => {
  it("既知の優先度はそのまま返す", () => {
    expect(toPriorityValue("high")).toBe("high")
  })

  it("想定外の値は medium として扱う", () => {
    expect(toPriorityValue("urgent")).toBe("medium")
  })
})

describe("respondentEmailLabel", () => {
  it("メールアドレスが無い場合は代替の文言を返す", () => {
    expect(respondentEmailLabel(null)).toBe("メールアドレス未登録")
    expect(respondentEmailLabel(undefined)).toBe("メールアドレス未登録")
  })
})

describe("hexToRgba", () => {
  it("16 進の色を rgba に変換する", () => {
    expect(hexToRgba("#10B981", 0.1)).toBe("rgba(16, 185, 129, 0.1)")
  })

  it("色が未設定ならグレーにする", () => {
    expect(hexToRgba(null, 0.5)).toBe("rgba(156, 163, 175, 0.5)")
  })
})

describe("sortStatuses", () => {
  it("表示順に並べ、元の配列は変更しない", () => {
    const statuses = [
      { id: "b", display_order: 2 },
      { id: "a", display_order: 1 },
    ] as FormStatus[]

    expect(sortStatuses(statuses).map((s) => s.id)).toEqual(["a", "b"])
    expect(statuses.map((s) => s.id)).toEqual(["b", "a"])
  })
})
