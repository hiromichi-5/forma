# 認可・権限モデル

## ロールモデル

フォームごとにメンバーシップとロールが存在する。

| ロール | 値 | 権限 |
| --- | --- | --- |
| **Admin** | `"admin"` | フォームに対するすべての操作（メンバー管理、招待、設定変更含む） |
| **Editor** | `"editor"` | チケット操作、ステータス操作、同期、フォーム閲覧 |

ロールは `entity/member.go` に定数として定義されている:

```go
const (
    RoleAdmin  = "admin"
    RoleEditor = "editor"
)
```

## 認証フロー

`middleware/session.go` の `SessionMiddleware` が認証を担当する。

1. リクエストから Cookie（`forma_token`）を取得
2. Cookie 値を UUID としてパースし、セッション ID として扱う
3. `UserRepository.GetSessionByID()` でセッションの有効性を検証
4. 有効であればユーザー ID を `gin.Context` に格納（`"userID"` キー）
5. 無効であれば `CodeInvalidSession`（HTTP 401）を返して中断

認証不要のエンドポイント（`/v1/auth/*`）はこのミドルウェアを経由しない。

## 権限チェックパターン

`usecase/authorization.go` に2つのヘルパー関数を定義し、各 UseCase から呼び出す。

### requireEditor

Admin または Editor ロールを持つメンバーのみ許可する。チケット閲覧・操作、ステータス操作、同期など多くの操作で使用。

```go
func requireEditor(ctx, memberRepo, formID, userID) error
```

### requireAdmin

Admin ロールのみ許可する。メンバー管理、招待管理で使用。

```go
func requireAdmin(ctx, memberRepo, formID, userID) error
```

### エラーの使い分け

| 状況 | 返すエラー | HTTP | 理由 |
| --- | --- | --- | --- |
| ユーザーがフォームのメンバーでない | `RESOURCE_HIDDEN` | 404 | フォームの存在自体を隠す |
| `requireEditor` でロール不足 | `RESOURCE_HIDDEN` | 404 | フォームの存在自体を隠す |
| `requireAdmin` でメンバーでない | `RESOURCE_HIDDEN` | 404 | フォームの存在自体を隠す |
| `requireAdmin` で Editor がアクセス | `FORBIDDEN` | 403 | メンバーであることは分かっているため、権限不足を伝える |

`requireEditor` は非メンバーでもロール不足でも一律 `RESOURCE_HIDDEN` を返す。メンバーであること自体が確認できていないため、存在を隠すのが適切。

`requireAdmin` はメンバーが見つからない場合は `RESOURCE_HIDDEN`、メンバーだが Admin でない場合は `FORBIDDEN` を返す。メンバーであることを既に知っているユーザーには「権限がない」と明示する方が有用。

## 各エンドポイントの権限要件

### 認証不要

| エンドポイント | 説明 |
| --- | --- |
| `POST /v1/auth/signup` | ユーザー登録 |
| `POST /v1/auth/login` | ログイン |
| `POST /v1/auth/logout` | ログアウト |
| `POST /v1/auth/verify-email` | メール認証 |
| `POST /v1/auth/verify-email/resend` | メール認証再送 |
| `POST /v1/auth/password-reset` | パスワードリセット要求 |
| `POST /v1/auth/password-reset/confirm` | パスワードリセット確認 |

### 認証のみ（ロール不問）

| エンドポイント | 説明 |
| --- | --- |
| `GET /v1/me` | 自分のプロフィール取得 |
| `PATCH /v1/me` | 表示名変更 |
| `DELETE /v1/me` | アカウント削除 |
| `PATCH /v1/me/password` | パスワード変更 |
| `GET /v1/forms` | 自分がアクセスできるフォーム一覧 |
| `POST /v1/forms` | フォーム登録（登録者が Admin になる） |
| `POST /v1/invites/:id/accept` | 招待の承諾（メールアドレス一致が必要） |

### Editor 以上（`requireEditor`）

| エンドポイント | 説明 |
| --- | --- |
| `GET /v1/forms/:id` | フォーム詳細 |
| `PATCH /v1/forms/:id` | タイトル質問 ID の設定 |
| `GET /v1/forms/:id/questions` | 質問一覧 |
| `GET /v1/forms/:id/statuses` | ステータス一覧 |
| `POST /v1/forms/:id/statuses` | ステータス作成 |
| `PATCH /v1/forms/:id/statuses/:id` | ステータス更新 / デフォルトステータス設定 |
| `DELETE /v1/forms/:id/statuses/:id` | ステータス削除 |
| `POST /v1/forms/:id/sync` | フォーム同期 |
| `GET /v1/forms/:id/members` | メンバー一覧 |
| `GET /v1/tickets` | チケット一覧 |
| `GET /v1/tickets/:id` | チケット詳細 |
| `PATCH /v1/tickets/:id` | チケット更新 |
| `GET /v1/tickets/:id/histories` | チケット変更履歴 |

### Admin のみ（`requireAdmin`）

| エンドポイント | 説明 |
| --- | --- |
| `POST /v1/forms/:id/members` | メンバー追加 |
| `PUT /v1/forms/:id/members/:id` | ロール変更 |
| `DELETE /v1/forms/:id/members/:id` | メンバー削除 |
| `GET /v1/forms/:id/invites` | 招待一覧 |
| `POST /v1/forms/:id/invites` | 招待作成 |
| `DELETE /v1/forms/:id/invites/:id` | 招待削除 |
