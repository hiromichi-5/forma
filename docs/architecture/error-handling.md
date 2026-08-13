# エラーハンドリング

## 設計方針

エラーコードはビジネスロジック上の意味に基づき命名する。

例えば、HTTP 404 エラーは「リソースが見つからない」という技術的な意味を持つが、ビジネスロジック上では「非メンバーへの情報隠蔽」と「実際に存在しない」の2種類の意味がある。
この設計により、同じ HTTP 404 でも `RESOURCE_HIDDEN`（非メンバーへの情報隠蔽）と `FORM_NOT_FOUND`（実際に存在しない）を区別でき、クライアント側で適切なメッセージを出し分けられる。

## 各層のエラー表現

### entity 層

`entity/errors.go` でエラーの語彙を定義する。

```go
type Code string      // エラーコード
type FieldCode string // フィールドレベルのエラーコード

type Error struct {
    Code   Code
    Fields []FieldError  // バリデーションエラー時のみ使用
    Err    error         // ラップされた元エラー
}

type FieldError struct {
    Field string
    Code  FieldCode
}
```

### repository 層

`repository/errors.go` で技術的エラーを定義する。

```go
var (
    ErrNotFound  = errors.New("not found")
    ErrConflict  = errors.New("conflict")
    ErrForbidden = errors.New("forbidden")
)
```

これらはビジネス上の意味を持たず、「行が見つからなかった」「一意制約に違反した」「外部 API が 403 を返した」という事実のみを表す。

### infra/postgres 層

`infra/postgres/errors.go` で PostgreSQL 固有のエラーを repository エラーに変換する。

```go
func repoError(err error) error {
    // pgx.ErrNoRows → repository.ErrNotFound
    // PG code 23505 (unique_violation) → repository.ErrConflict
}

func rowsError(n int64) error {
    // 0 行更新 → repository.ErrNotFound
}
```

### usecase 層

usecase は repository エラーを受け取り、ビジネスコンテキストに応じたドメインエラーに変換する。

```go
form, err := uc.formRepo.GetByID(ctx, formID)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        return entity.NewError(entity.CodeFormNotFound)
    }
    return err
}
```

例えば、同じ `repository.ErrNotFound` でも、コンテキストによって変換先が異なる。

| UseCase の状況 | 変換先コード |
| --- | --- |
| フォームが見つからない | `CodeFormNotFound` |
| 非メンバーがフォームにアクセス | `CodeResourceHidden` |
| 招待が見つからない | `CodeInviteNotFound` |
| 同一フォームの重複登録 | `CodeFormAlreadyRegistered` |

### handler 層

`handler/error.go` の `errorDefs` マップで変換する。

```go
var errorDefs = map[entity.Code]errorDef{
    entity.CodeInvalidCredentials:        {401, "メールアドレスまたはパスワードが正しくありません"},
    entity.CodeForbidden:                 {403, "この操作を実行する権限がありません"},
    // ...
}
```

すべてのハンドラーは `handleError(c, err)` を呼ぶだけでよい。エンドポイントごとの分岐は行わない。

## エラーフロー全体図

```text
PostgreSQL エラー
  ↓ repoError()
repository.ErrNotFound / ErrConflict
  ↓ usecase 内で errors.Is() 判定
entity.Error{Code: CodeXxx}
  ↓ handleError()
HTTP レスポンス {code, message, fields?}
```

## エラーコード一覧

### 認証・認可

| コード | HTTP | 意味 |
| --- | --- | --- |
| `INVALID_CREDENTIALS` | 401 | メールアドレスまたはパスワードが正しくない |
| `INVALID_SESSION` | 401 | セッションが無効・期限切れ |
| `EMAIL_NOT_VERIFIED` | 403 | メール認証が完了していない |
| `FORBIDDEN` | 403 | 権限不足（メンバーだが権限がない） |
| `RESOURCE_HIDDEN` | 404 | 非メンバーに対してリソースの存在を隠す |

### リソース不在

| コード | HTTP | 意味 |
| --- | --- | --- |
| `USER_NOT_FOUND` | 404 | ユーザーが存在しない |
| `FORM_NOT_FOUND` | 404 | フォームが存在しない |
| `FORM_NOT_SHARED` | 404 | フォームがサービスアカウントに共有されていない |
| `TOKEN_NOT_FOUND` | 404 | 検証トークンが存在しない |
| `INVITE_NOT_FOUND` | 404 | 招待が存在しない |

### ビジネスルール違反

| コード | HTTP | 意味 |
| --- | --- | --- |
| `INVITE_EXPIRED` | 404 | 招待の有効期限切れ |
| `ALREADY_MEMBER` | 409 | 既にメンバーである |
| `INCORRECT_PASSWORD` | 403 | 現在のパスワードが正しくない |
| `LAST_ADMIN` | 409 | 最後の管理者は削除・降格できない |
| `CONFLICT` | 409 | 汎用的なリソース競合 |
| `FORM_ALREADY_REGISTERED` | 409 | 同一 Google フォームの重複登録 |
| `ACTIVE_INVITE_ALREADY_EXISTS` | 409 | 同一メールへの有効な招待が既に存在 |
| `STATUS_CONFLICT` | 409 | ステータス名または表示順の重複 |
| `NOTIFICATION_DISABLED` | 409 | 該当種別の通知設定が `off` である |
| `RESPONDENT_EMAIL_MISSING` | 409 | チケットに回答者のメールアドレスがない |
| `NOTIFICATION_RATE_LIMITED` | 429 | 通知の手動送信が短時間に繰り返された |

### バリデーション

| コード | HTTP | 意味 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | 入力値の検証エラー（`fields` 配列で詳細を返す） |

**フィールドエラーコード:** `REQUIRED`, `TOO_SHORT`, `INVALID_FORMAT`, `INVALID_VALUE`

## レスポンス形式

```json
{
  "code": "VALIDATION_ERROR",
  "message": "入力内容に誤りがあります",
  "fields": [
    { "field": "email", "code": "REQUIRED" },
    { "field": "password", "code": "TOO_SHORT" }
  ]
}
```

- `fields` は `VALIDATION_ERROR` の場合のみ含まれる。
- マッピングに存在しないコードの場合は HTTP 500 + `{"code": "INTERNAL"}` を返す。
