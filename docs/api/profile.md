# プロフィール API

## 概要

認証済みユーザーの自身のプロフィール取得・更新・削除、およびパスワード変更を提供する。

---

## GET /v1/me

自分のプロフィールを取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/me` |
| 認証 | 必要（SessionMiddleware） |

### レスポンス

#### 200 OK

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string | ユーザー UUID |
| `email` | string | メールアドレス |
| `display_name` | string | 表示名 |
| `verified_at` | string? | メール認証日時（RFC3339Nano）。未認証なら `null` |

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "display_name": "山田太郎",
  "verified_at": "2026-01-15T10:30:00.000Z"
}
```

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `INVALID_SESSION` | 401 | セッションが無効 |
| `USER_NOT_FOUND` | 404 | ユーザーが存在しない |

## PATCH /v1/me

表示名を更新する。

| 項目 | 値 |
| --- | --- |
| メソッド | `PATCH` |
| パス | `/v1/me` |
| 認証 | 必要（SessionMiddleware） |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `display_name` | string | Yes | 新しい表示名 |

### レスポンス

**200 OK** — 更新後のプロフィール（GET /v1/me と同じ形式）

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `INVALID_SESSION` | 401 | セッションが無効 |
| `VALIDATION_ERROR` | 400 | 表示名が空 |
| `USER_NOT_FOUND` | 404 | ユーザーが存在しない |

## DELETE /v1/me

自分のアカウントを削除する。

| 項目 | 値 |
| --- | --- |
| メソッド | `DELETE` |
| パス | `/v1/me` |
| 認証 | 必要（SessionMiddleware） |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `INVALID_SESSION` | 401 | セッションが無効 |
| `USER_NOT_FOUND` | 404 | ユーザーが存在しない |

## PATCH /v1/me/password

パスワードを変更する。現在のパスワードの検証が必要。

| 項目 | 値 |
| --- | --- |
| メソッド | `PATCH` |
| パス | `/v1/me/password` |
| 認証 | 必要（SessionMiddleware） |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `current_password` | string | Yes | 現在のパスワード |
| `new_password` | string | Yes | 新しいパスワード（8文字以上） |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `INVALID_SESSION` | 401 | セッションが無効 |
| `VALIDATION_ERROR` | 400 | 入力値が不正・新パスワードが8文字未満 |
| `INCORRECT_PASSWORD` | 403 | 現在のパスワードが正しくない |
| `USER_NOT_FOUND` | 404 | ユーザーが存在しない |
