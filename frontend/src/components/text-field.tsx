import type { ComponentProps, ReactNode } from "react"
import type { FieldError as FormFieldError } from "react-hook-form"
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"

type TextFieldProps = ComponentProps<typeof Input> & {
  id: string
  label: ReactNode
  /** react-hook-form の検証エラー。あれば入力欄の下に表示し、入力欄を invalid にする。 */
  error?: FormFieldError
  /** 入力の補足（例: 8文字以上）。エラーがある間はエラーを優先して表示する。 */
  description?: ReactNode
  /** ラベルの右側に置く要素（例: パスワードを忘れた方へのリンク）。 */
  labelAction?: ReactNode
}

/** ラベル・入力欄・エラーをひとまとめにした入力項目。react-hook-form の register と組み合わせて使う。 */
export function TextField({
  id,
  label,
  error,
  description,
  labelAction,
  ...inputProps
}: TextFieldProps) {
  const errorId = `${id}-error`
  const descriptionId = `${id}-description`
  // 既存の Label と行の高さを揃える（FieldLabel の既定は leading-snug）。
  const labelElement = (
    <FieldLabel htmlFor={id} className="leading-none">
      {label}
    </FieldLabel>
  )

  return (
    <Field className="gap-2" data-invalid={!!error}>
      {labelAction ? (
        <div className="flex items-center justify-between">
          {labelElement}
          {labelAction}
        </div>
      ) : (
        labelElement
      )}
      <Input
        id={id}
        aria-invalid={!!error}
        aria-describedby={error ? errorId : description ? descriptionId : undefined}
        {...inputProps}
      />
      {error ? (
        <FieldError id={errorId} errors={[error]} />
      ) : (
        description && (
          <FieldDescription id={descriptionId} className="text-xs">
            {description}
          </FieldDescription>
        )
      )}
    </Field>
  )
}
