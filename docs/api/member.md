# メンバー API

## 概要

フォームのメンバー管理（一覧、追加、ロール変更、削除）を提供する。メンバー管理は原則 Admin のみ、一覧は Editor 以上。

---

## GET /v1/forms/:form_id/members

フォームのメンバー一覧を取得する。

| 項目 | 値 |
|------|-----|
| メソッド | `GET` |
| パス | `/v1/forms/:form_id/members` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

**200 OK**

```json
{
  "members": [
    {
      "id": "550e8400-...",
      "email": "user@example.com",
      "display_name": "山田太郎",
      "role": "admin"
    }
  ]
}
```

## POST /v1/forms/:form_id/members

メールアドレスでユーザーを検索し、フォームのメンバーとして追加する。

| 項目 | 値 |
|------|-----|
| メソッド | `POST` |
| パス | `/v1/forms/:form_id/members` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Admin のみ |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `email` | string | Yes | 追加するユーザーのメールアドレス |
| `role` | string | Yes | ロール（`admin` または `editor`） |

### レスポンス

**201 Created**（ボディなし）

### エラー

| コード | HTTP | 条件 |
|---|---|---|
| `VALIDATION_ERROR` | 400 | ロールが不正 |
| `RESOURCE_HIDDEN` | 404 | 操作者がメンバーでない |
| `FORBIDDEN` | 403 | 操作者が Admin でない |
| `USER_NOT_FOUND` | 404 | 指定メールのユーザーが存在しない |
| `ALREADY_MEMBER` | 409 | 既にメンバーである |

### 補足
- **招待APIへの一本化を検討中。** 現行APIはメールアドレスでユーザーを検索して追加するが、招待APIはメールアドレスベースで招待を発行し、ユーザーが承諾するフロー。セキュリティとユーザー体験の観点から、招待APIへの一本化を検討中。

## PUT /v1/forms/:form_id/members/:user_id

メンバーのロールを変更する。

| 項目 | 値 |
|------|-----|
| メソッド | `PUT` |
| パス | `/v1/forms/:form_id/members/:user_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Admin のみ |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `role` | string | Yes | 新しいロール（`admin` または `editor`） |

### レスポンス

**204 No Content**

### エラー

| コード | HTTP | 条件 |
|---|---|---|
| `VALIDATION_ERROR` | 400 | ロールが不正 |
| `RESOURCE_HIDDEN` | 404 | 操作者がメンバーでない、または対象がメンバーでない |
| `FORBIDDEN` | 403 | 操作者が Admin でない |
| `LAST_ADMIN` | 409 | 最後の Admin を Editor に降格しようとした |

### 補足

- 同じロールへの変更は何もせず成功を返すことで冪等にする

---

## DELETE /v1/forms/:form_id/members/:user_id

メンバーをフォームから削除する。

| 項目 | 値 |
|------|-----|
| メソッド | `DELETE` |
| パス | `/v1/forms/:form_id/members/:user_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Admin のみ |

### レスポンス

**204 No Content**

### エラー

| コード | HTTP | 条件 |
|---|---|---|
| `RESOURCE_HIDDEN` | 404 | 操作者がメンバーでない |
| `FORBIDDEN` | 403 | 操作者が Admin でない |
| `LAST_ADMIN` | 409 | 最後の Admin を削除しようとした |

### 補足

- 対象が既にメンバーでない場合も 204 を返すことで冪等にする
