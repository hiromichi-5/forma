# トランザクション管理

## UnitOfWork パターン

複数のリポジトリ操作をアトミックに実行するために、Generic `UnitOfWork[T]` interface を使用する。

### Interface 定義

`repository/uow.go`:

```go
type UnitOfWork[T any] interface {
    Do(ctx context.Context, fn func(repos T) error) error
}
```

各 UseCase は、トランザクション内で必要なリポジトリのみを含む Repos struct を型パラメータ `T` として受け取る。

```go
type AuthRepos struct {
    User UserRepository
}

type FormRepos struct {
    Form   FormRepository
    Member MemberRepository
    Status StatusRepository
}

type InviteRepos struct {
    Invite InviteRepository
    User   UserRepository
    Member MemberRepository
}

type StatusRepos struct {
    Status StatusRepository
}

type TicketRepos struct {
    Ticket TicketRepository
    Status StatusRepository
    User   UserRepository
}
```

### 実装

`infra/postgres/uow.go` で PostgreSQL トランザクションとして実装する。

```go
type unitOfWork[T any] struct {
    pool    *pgxpool.Pool
    factory func(q *db.Queries) T
}

func (u *unitOfWork[T]) Do(ctx context.Context, fn func(repos T) error) error {
    tx, err := u.pool.Begin(ctx)
    // ...
    q := db.New(tx)
    repos := u.factory(q)  // factory が必要なリポジトリのみ初期化
    if err := fn(repos); err != nil {
        _ = tx.Rollback(ctx)
        return err
    }
    return tx.Commit(ctx)
}
```

各 UseCase 用のコンストラクタが提供される:

```go
func NewAuthUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.AuthRepos]
func NewFormUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.FormRepos]
func NewInviteUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.InviteRepos]
func NewStatusUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.StatusRepos]
func NewTicketUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.TicketRepos]
```

`fn` が error を返せば自動ロールバック、nil を返せばコミットされる。

## TX あり / TX なしの使い分け

### TX が必要な場合

`uc.uow.Do()` を使う。
複数のリポジトリ操作を不可分に実行する必要がある場合。コールバック内では `repos.Xxx` を使う。

```go
err := uc.uow.Do(ctx, func(repos repository.FormRepos) error {
    form, err := repos.Form.Create(ctx, ...)
    if err != nil { return err }
    if err := repos.Member.Upsert(ctx, ...); err != nil { return err }
    return repos.Status.Create(ctx, ...)
})
```

### TX が不要な場合 — `uc.xxxRepo` 直接使用

単一のリポジトリ操作、または複数操作だが中間状態が許容される場合。DI で注入されたリポジトリを直接呼ぶ。

```go
form, err := uc.formRepo.GetByID(ctx, formID)
```

## 新しい TX 対象操作の追加手順

### 既存の UnitOfWork を使う場合

UseCase メソッド内で `uc.uow.Do(ctx, func(repos repository.XxxRepos) error { ... })` を使用するだけ。

### 新しいリポジトリを TX 対象に追加する場合

1. `repository/uow.go` の該当 Repos struct にフィールドを追加
2. `infra/postgres/uow.go` の該当 `NewXxxUoW` の factory 内で新しいリポジトリのインスタンスを生成

### 新しい UseCase 用の UnitOfWork を追加する場合

1. `repository/uow.go` に新しい Repos struct を定義
2. `infra/postgres/uow.go` に `NewXxxUoW` コンストラクタを追加
3. `cmd/api/main.go` で UseCase に注入
