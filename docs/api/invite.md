# 招待 API

## 概要

フォームへの招待の作成・一覧・削除・承諾を提供する。招待は 7 日間有効で、メールアドレスベースで発行される。

## POST /v1/forms/:form_id/invites

フォームへの招待を作成する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/forms/:form_id/invites` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Admin のみ |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `email` | string | Yes | 招待先メールアドレス |
| `role` | string | Yes | 付与するロール（`admin` または `editor`） |

### レスポンス

#### 201 Created

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `invite_id` | string | 招待の UUID |
| `expires_at` | string | 有効期限（RFC3339） |

```json
{
  "invite_id": "550e8400-...",
  "expires_at": "2026-04-07T12:00:00Z"
}
```

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | ロールが不正 |
| `RESOURCE_HIDDEN` | 404 | 操作者がメンバーでない |
| `FORBIDDEN` | 403 | 操作者が Admin でない |
| `ALREADY_MEMBER` | 409 | 招待先が既にメンバー |
| `ACTIVE_INVITE_ALREADY_EXISTS` | 409 | 同一メールへの有効な招待が既に存在 |

### 補足

- 招待有効期限は 7 日間（`inviteTTL = 7 * 24 * time.Hour`）
- 未登録のメールアドレスへの招待も可能。ユーザーが登録後に承諾できる
- 同一フォームかつメールアドレスに対する有効な招待は一意制約で1件に制限される

## GET /v1/forms/:form_id/invites

フォームの有効な招待一覧を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/forms/:form_id/invites` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Admin のみ |

### レスポンス

#### 200 OK

```json
{
  "invites": [
    {
      "id": "550e8400-...",
      "email": "invited@example.com",
      "role": "editor",
      "invited_by": "660e8400-...",
      "expires_at": "2026-04-07T12:00:00Z",
      "created_at": "2026-03-31T12:00:00Z"
    }
  ]
}
```

## DELETE /v1/forms/:form_id/invites/:invite_id

招待を削除（取り消し）する。

| 項目 | 値 |
| --- | --- |
| メソッド | `DELETE` |
| パス | `/v1/forms/:form_id/invites/:invite_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Admin のみ |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `RESOURCE_HIDDEN` | 404 | 操作者がメンバーでない |
| `FORBIDDEN` | 403 | 操作者が Admin でない |
| `INVITE_NOT_FOUND` | 404 | 招待が存在しない、または別フォームの招待 |

### 補足

- トランザクション内で `GetForUpdate` → Delete の順で実行し、競合を防ぐ

## POST /v1/invites/:invite_id/accept

招待を承諾し、フォームのメンバーになる。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/invites/:invite_id/accept` |
| 認証 | 必要（SessionMiddleware） |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `INVITE_NOT_FOUND` | 404 | 招待が存在しない・既に承諾済み |
| `INVITE_EXPIRED` | 404 | 招待の有効期限切れ |
| `RESOURCE_HIDDEN` | 404 | ログインユーザーのメールと招待メールが一致しない |
| `USER_NOT_FOUND` | 404 | ユーザーが存在しない |
| `ALREADY_MEMBER` | 409 | 既にメンバーである |

### 補足

- フラットなパス（`/v1/invites/:id/accept`）を採用。
  - 招待承諾は受信者の視点で行う操作であり、フォーム ID を知っている必要がないため
