/** ステータスの色の入力。色を使わない（null）場合も、最後に選んだ色は保持しておく。 */
export type ColorInputState = {
  enabled: boolean
  value: string
}

const DEFAULT_COLOR = "#3B82F6"

export const createColorState = (color?: string | null): ColorInputState => ({
  enabled: Boolean(color),
  value: color ?? DEFAULT_COLOR,
})

export const colorValueOrNull = (state: ColorInputState): string | null =>
  state.enabled ? state.value : null
