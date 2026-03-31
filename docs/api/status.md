# ステータス API

## 概要

フォームに紐づくチケットステータスの CRUD とデフォルトステータスの設定を提供する。

---

## GET /v1/forms/:form_id/statuses

フォームのステータス一覧を取得する。

| 項目 | 値 |
|------|-----|
| メソッド | `GET` |
| パス | `/v1/forms/:form_id/statuses` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

**200 OK**

```json
{
  "statuses": [
    {
      "id": "550e8400-...",
      "form_id": "660e8400-...",
      "name": "未対応",
      "color": "#E53935",
      "display_order": 1,
      "is_default": true
    }
  ]
}
```

## POST /v1/forms/:form_id/statuses

新しいステータスを作成する。

| 項目 | 値 |
|------|-----|
| メソッド | `POST` |
| パス | `/v1/forms/:form_id/statuses` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `name` | string | Yes | ステータス名 |
| `color` | string? | No | カラーコード（例: `#E53935`） |
| `display_order` | int | Yes | 表示順（1以上） |
| `is_default` | bool | No | デフォルトにするか（デフォルト: `false`） |

### レスポンス

**201 Created** — 作成されたステータス

```json
{
  "id": "550e8400-...",
  "form_id": "660e8400-...",
  "name": "保留",
  "color": "#FFC107",
  "display_order": 4,
  "is_default": false
}
```

### エラー

| コード | HTTP | 条件 |
|---|---|---|
| `VALIDATION_ERROR` | 400 | 名前が空、display_order が 0 以下 |
| `RESOURCE_HIDDEN` | 404 | メンバーでない |
| `STATUS_CONFLICT` | 409 | 名前または表示順が既存ステータスと重複 |

### 補足

- `is_default=true` の場合、トランザクション内で作成 → 既存デフォルト解除 → 新デフォルト設定の順で実行してデフォルトステータスを設定する

---

## PATCH /v1/forms/:form_id/statuses/:status_id

ステータスの名前・カラー・表示順を更新する。

| 項目 | 値 |
|------|-----|
| メソッド | `PATCH` |
| パス | `/v1/forms/:form_id/statuses/:status_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### リクエストボディ

すべてオプショナル。指定されたフィールドのみ更新される。

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `name` | string? | No | 新しいステータス名 |
| `color` | string? | No | 新しいカラーコード。空文字で null に設定 |
| `display_order` | int? | No | 新しい表示順 |

### レスポンス

**200 OK** — 更新後のステータス

### エラー

| コード | HTTP | 条件 |
|---|---|---|
| `VALIDATION_ERROR` | 400 | 名前が空文字、display_order が 0 以下 |
| `RESOURCE_HIDDEN` | 404 | メンバーでない、ステータスが存在しない、別フォームのステータス |
| `STATUS_CONFLICT` | 409 | 名前または表示順が重複 |

## POST /v1/forms/:form_id/statuses/:status_id/default

ステータスをデフォルトに設定する。

| 項目 | 値 |
|------|-----|
| メソッド | `POST` |
| パス | `/v1/forms/:form_id/statuses/:status_id/default` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

**200 OK**（ボディなし）

### エラー

| コード | HTTP | 条件 |
|---|---|---|
| `RESOURCE_HIDDEN` | 404 | メンバーでない、ステータスが存在しない |

### 内部フロー

```
handler.PostV1FormsIdStatusesStatusIdDefault
  → StatusUseCase.SetDefaultStatus(formID, userID, statusID)
    → requireEditor(memberRepo, formID, userID)
    → txm.Do:
      → repos.Status.ClearDefault(formID)          // 現在のデフォルトを解除
      → repos.Status.SetDefault(formID, statusID)  // 新しいデフォルトを設定
        → ErrNotFound → CodeResourceHidden
  → 200
```

## DELETE /v1/forms/:form_id/statuses/:status_id

ステータスを削除する。使用中のステータスやデフォルトステータスは削除できない。

| 項目 | 値 |
|------|-----|
| メソッド | `DELETE` |
| パス | `/v1/forms/:form_id/statuses/:status_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

**204 No Content**

### エラー

| コード | HTTP | 条件 |
|---|---|---|
| `VALIDATION_ERROR` | 400 | デフォルトステータスの削除、使用中のステータスの削除 |
| `RESOURCE_HIDDEN` | 404 | メンバーでない、ステータスが存在しない、別フォームのステータス |

### 補足

- デフォルトステータスを削除したい場合は、先に別のステータスをデフォルトに設定する必要がある
- チケットが紐づいているステータスは削除できない。先にチケットのステータスを変更する必要がある
