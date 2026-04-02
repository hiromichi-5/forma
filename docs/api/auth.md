# 認証 API

## 概要

ユーザー登録、ログイン・ログアウト、メール認証、パスワードリセットを提供する。Cookie ベースのセッション管理を使用。

## POST /v1/auth/signup

ユーザーを新規登録し、メール認証トークンを発行する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/auth/signup` |
| 認証 | 不要 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `email` | string | Yes | メールアドレス（email 形式） |
| `password` | string | Yes | パスワード（8文字以上） |
| `display_name` | string | Yes | 表示名 |

```json
{
  "email": "user@example.com",
  "password": "password123",
  "display_name": "山田太郎"
}
```

### レスポンス

#### 201 Created

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string | 作成されたユーザーの UUID |

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | 入力値が不正 |
| `CONFLICT` | 409 | メールアドレスが既に登録済み |

### 補足

- メール認証トークンの有効期限は 24 時間

## POST /v1/auth/login

メールアドレスとパスワードで認証し、セッション Cookie を発行する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/auth/login` |
| 認証 | 不要 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `email` | string | Yes | メールアドレス |
| `password` | string | Yes | パスワード |

### レスポンス

#### 200 OK

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `session_id` | string | セッション ID（UUID） |

レスポンスに加えて `Set-Cookie` ヘッダーでセッション Cookie を設定する。

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | 入力値が不正 |
| `INVALID_CREDENTIALS` | 401 | メールアドレスまたはパスワードが正しくない |
| `EMAIL_NOT_VERIFIED` | 403 | メール認証が完了していない |

### 補足

- Cookie 設定: `HttpOnly=true`, `SameSite=Lax`, `Secure=true`

## POST /v1/auth/logout

セッションを無効化し、Cookie を削除する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/auth/logout` |
| 認証 | Cookie（`forma_token`） |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `INVALID_SESSION` | 401 | Cookie がない・無効 |

## POST /v1/auth/verify-email

メール認証トークンを検証し、ユーザーのメール認証を完了する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/auth/verify-email` |
| 認証 | 不要 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `token` | string | Yes | メール認証トークン |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | トークンが空 |
| `TOKEN_NOT_FOUND` | 404 | トークンが存在しない・既に使用済み |

## POST /v1/auth/verify-email/resend

メール認証トークンを再発行する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/auth/verify-email/resend` |
| 認証 | 不要 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `email` | string | Yes | メールアドレス |

### レスポンス

#### 202 Accepted

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | メールアドレスが空 |

### 補足

- ユーザーが存在しない場合や既に認証済みの場合も 202 を返すことで、ユーザーの存在を隠蔽する

## POST /v1/auth/password-reset

パスワードリセットトークンを発行する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/auth/password-reset` |
| 認証 | 不要 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `email` | string | Yes | メールアドレス |

### レスポンス

#### 202 Accepted

### 補足

- ユーザーが存在しない場合も 202 を返すことで、ユーザーの存在を隠蔽する

---

## POST /v1/auth/password-reset/confirm

パスワードリセットトークンを検証し、新しいパスワードを設定する。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/auth/password-reset/confirm` |
| 認証 | 不要 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `token` | string | Yes | パスワードリセットトークン |
| `new_password` | string | Yes | 新しいパスワード（8文字以上） |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | 入力値が不正・パスワードが8文字未満 |
| `TOKEN_NOT_FOUND` | 404 | トークンが存在しない・既に使用済み |
