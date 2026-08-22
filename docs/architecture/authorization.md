# 認可・権限モデル

## ロールモデル

フォームごとにメンバーシップとロールが存在する。

| ロール | 値 | 権限 |
| --- | --- | --- |
| **Admin** | `"admin"` | フォームに対するすべての操作（メンバー管理、招待、設定変更含む） |
| **Editor** | `"editor"` | チケット操作、ステータス操作、同期、フォーム閲覧 |

ロールは `entity/member.go` に型として定義されている:

```go
type Role string

const (
    RoleAdmin  Role = "admin"
    RoleEditor Role = "editor"
)

func (r Role) Valid() bool    // 入力として受理できる値か
func (r Role) CanEdit() bool  // 編集権限があるか
func (r Role) CanAdmin() bool // 管理権限があるか
```

## 認証フロー

`middleware/session.go` の `SessionMiddleware` が認証を担当する。

1. リクエストから Cookie（`forma_token`）を取得
2. Cookie 値を UUID としてパースし、セッション ID として扱う
3. `SessionRepository.GetByID()` でセッションの有効性を検証
4. 有効であればユーザー ID を `gin.Context` に格納（`"userID"` キー）
5. 無効であれば `CodeInvalidSession`（HTTP 401）を返して中断

認証不要のエンドポイント（`/v1/auth/*`）はこのミドルウェアを経由しない。

## 権限チェックパターン

`usecase/authorization.go` の `Authorizer` が担当する。`MemberRepository` を保持し、各 UseCase に注入される。

```go
type Authorizer struct {
    memberRepo repository.MemberRepository
}

// Admin または Editor。チケット・ステータス操作、同期、フォーム閲覧で使う
func (a *Authorizer) RequireEditor(ctx context.Context, formID, userID uuid.UUID) error

// Admin のみ。メンバー管理、招待、フォーム登録解除、通知設定の変更で使う
func (a *Authorizer) RequireAdmin(ctx context.Context, formID, userID uuid.UUID) error
```

UseCase は対象のメソッド冒頭で呼ぶ。

```go
func (uc *StatusUseCase) ListStatuses(
    ctx context.Context,
    formID, userID uuid.UUID,
) ([]entity.FormStatus, error) {
    if err := uc.authz.RequireEditor(ctx, formID, userID); err != nil {
        return nil, err
    }
    return uc.statusRepo.List(ctx, formID)
}
```

SSE のように対応する UseCase メソッドを持たないエンドポイントでは、ハンドラが `Authorizer` を直接受け取る（`handler/stream.go` の `FormAuthorizer`）。

`SyncUseCase` / `StatusUseCase` / `NotificationUseCase` は `MemberRepository` を認可以外に使わないため、`Authorizer` のみを持つ。`FormUseCase` / `TicketUseCase` / `MemberUseCase` / `InviteUseCase` は「対象ユーザーの在籍確認」などで `MemberRepository` も併せて必要とする。

### エラーの使い分け

| 状況 | 返すエラー | HTTP | 理由 |
| --- | --- | --- | --- |
| ユーザーがフォームのメンバーでない | `RESOURCE_HIDDEN` | 404 | フォームの存在自体を隠す |
| `RequireEditor` でロール不足 | `RESOURCE_HIDDEN` | 404 | フォームの存在自体を隠す |
| `RequireAdmin` でメンバーでない | `RESOURCE_HIDDEN` | 404 | フォームの存在自体を隠す |
| `RequireAdmin` で Editor がアクセス | `FORBIDDEN` | 403 | メンバーであることは分かっているため、権限不足を伝える |

`RequireEditor` は非メンバーでもロール不足でも一律 `RESOURCE_HIDDEN` を返す。メンバーであること自体が確認できていないため、存在を隠すのが適切。

`RequireAdmin` はメンバーが見つからない場合は `RESOURCE_HIDDEN`、メンバーだが Admin でない場合は `FORBIDDEN` を返す。メンバーであることを既に知っているユーザーには「権限がない」と明示する方が有用。
