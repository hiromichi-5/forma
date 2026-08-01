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

Admin ロールのみ許可する。

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
