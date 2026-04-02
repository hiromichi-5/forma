# エラーコード設計

## 背景・課題 (Background/Problem)

- 旧設計のセンチネルエラー（16個）では `ErrValidation` が万能すぎ、フロントが原因を判別できなかった
- エラーの粒度がバラバラだった（`ErrForbidden` は認可カテゴリ、`ErrInviteExpired` はビジネスルール）
- usecase が `return ErrValidation` だけ返すため「何が不正か」が handler に届かなかった
- 全ハンドラーで `errors.Is()` → HTTP ステータスの switch 文が重複していた

## 決定事項 (Decision)

型付き `Code` 定数 + `Error` 構造体方式を採用する。

```go
// entity/errors.go
type Code string
type Error struct {
    Code   Code
    Fields []FieldError
    Err    error
}
```

- entity は「何が起きたか」をドメインの語彙（`Code` 定数）で表現する。HTTP もメッセージも知らない
- handler が `Code` → HTTP ステータス + ユーザー向けメッセージを1ファイルのマッピングテーブルに集約する
- repository は技術的エラー（`ErrNotFound`, `ErrConflict`）のみを返し、usecase がビジネスコンテキストに応じたドメインコードに変換する

現在のエラーコード一覧とレスポンス形式は `docs/architecture/error-handling.md` を参照。

### 理由 (Reasons)

- `Code` が型付き定数のため、entity に一元管理して全語彙を見渡せる
- handler は endpoint ごとの特例分岐を持たず、機械的にテーブルを引くだけで済む
- フロントが `code` フィールドで正確にエラー種別を判別し、適切な UI を出し分けられる
- repository エラーがビジネス語彙に変換されるため、技術的エラーが上位層に漏れない

### 受け入れるトレードオフ (Accepted Trade-offs)

- Code 定数の数が多くなる（汎用コードに比べて語彙が増える）
- usecase で repository エラーをドメインコードに変換するボイラープレートが発生する
- 新しいビジネスエラーを追加するたびに entity の Code 定数と handler のマッピングテーブルの両方を更新する必要がある

## 検討した別の選択肢 (Alternatives Considered)

### 案A: ErrorKind（カテゴリ）+ Code + Message

Error 構造体に `Kind ErrorKind`（Validation, Forbidden, NotFound 等）を持たせ、handler が Kind → HTTP ステータスに変換する。

- メリット: HTTP ステータスの決定がシンプル
- デメリット: ErrorKind は実質 HTTP ステータスコードの別名であり、entity 層に HTTP の概念が入り込む

### 案B: センチネル + fmt.Errorf

現在のセンチネルエラーを維持しつつ、`fmt.Errorf("パスワードは8文字以上: %w", ErrValidation)` でメッセージを付与。

- メリット: 既存設計からの変更が最小限
- デメリット: Code の型安全性がない、フロントが判別に使う機械可読コードを付与しにくい

## 参考 (References)

## 議論 (Discussion)

### RESOURCE_HIDDEN と FORBIDDEN の分離

実装を進める中で、「存在隠蔽」と「権限不足」を同じ `FORBIDDEN` に載せると handler 側で endpoint ごとの特例が必要になることが判明した。

- `RESOURCE_HIDDEN`: その actor に対してリソースの存在を隠蔽したい場合（HTTP 404）
- `FORBIDDEN`: actor はリソースを知っているが、操作権限だけが不足している場合（HTTP 403）

具体的な使い分け:

- `requireEditor` 失敗 → フォームへの基礎アクセス権がないので `RESOURCE_HIDDEN`
- `requireAdmin` 失敗で非メンバー → `RESOURCE_HIDDEN`
- `requireAdmin` 失敗で Editor → メンバーであることは分かっているため `FORBIDDEN`

この分離により、usecase がアクセス制御の判断を完結させ、handler は機械的変換のみを行う構造を維持できる。

### repository.ErrConflict の粒度

`ErrConflict`（DB の一意制約違反）を usecase でそのまま汎用 `CONFLICT` にすると、業務上「何が重複したのか」が失われる。逆に `VALIDATION_ERROR` に落とすと、入力不正と競合状態が混ざる。

- 案A（すべて `VALIDATION_ERROR`）→ 却下。「入力が悪い」のか「既に存在する」のかが区別できない
- 案B（すべて `CONFLICT`）→ 却下。何の競合なのかが不明確
- 案C（採用）: 文脈別コードに分解し、必要な場合のみ汎用 `CONFLICT` を使う
  - フォーム登録重複 → `FORM_ALREADY_REGISTERED`
  - 招待作成重複 → `ACTIVE_INVITE_ALREADY_EXISTS`
  - ステータス名・表示順重複 → `STATUS_CONFLICT`
