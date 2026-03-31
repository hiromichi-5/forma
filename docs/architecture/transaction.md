# トランザクション管理

## TxManager パターン

複数のリポジトリ操作をアトミックに実行するために、`TxManager` interface を使用する。

### Interface 定義

`repository/transaction.go`:

```go
type TxManager interface {
    Do(ctx context.Context, fn func(repos Repos) error) error
}

type Repos struct {
    User   UserRepository
    Form   FormRepository
    Member MemberRepository
    Status StatusRepository
    Invite InviteRepository
    Ticket TicketRepository
}
```

### 実装

`infra/postgres/transaction.go` で PostgreSQL トランザクションとして実装する。

```go
func (m *TxManager) Do(ctx context.Context, fn func(repos repository.Repos) error) error {
    tx, err := m.pool.Begin(ctx)
    // ...
    q := db.New(tx)  // トランザクションに紐づく Queries
    repos := repository.Repos{
        User:   &UserRepository{q: q},
        Form:   &FormRepository{q: q},
        // ... 全リポジトリが同じ tx を共有
    }
    if err := fn(repos); err != nil {
        _ = tx.Rollback(ctx)
        return err
    }
    return tx.Commit(ctx)
}
```

`fn` が error を返せば自動ロールバック、nil を返せばコミットされる。

## TX あり / TX なしの使い分け

### TX が必要な場合

`uc.txm.Do()`を使う。
複数のリポジトリ操作を不可分に実行する必要がある場合。コールバック内では `repos.Xxx` を使う。

```go
err := uc.txm.Do(ctx, func(repos repository.Repos) error {
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

## TX が必要な操作一覧

| UseCase | 操作 | TX 内で行うこと |
|---|---|---|
| **FormUseCase.RegisterForm** | フォーム登録 | フォーム作成 + Admin メンバー作成 + デフォルトステータス3件作成 |
| **InviteUseCase.AcceptInvite** | 招待承諾 | 招待を SELECT FOR UPDATE → 有効性検証 → Accept + メンバー追加 |
| **InviteUseCase.DeleteInvite** | 招待削除 | 招待を SELECT FOR UPDATE → フォーム所属確認 → 削除 |
| **StatusUseCase.UpdateStatus** | ステータス更新（isDefault=true の場合） | ステータス更新 + 現在のデフォルトを解除 + 新しいデフォルトを設定 |
| **StatusUseCase.CreateStatus** | ステータス作成（isDefault=true の場合） | ステータス作成 + デフォルト設定 |
| **TicketUseCase.UpdateTicket** | チケット更新 | 最新状態を再取得 → ステータス/担当者/優先度更新 + 変更履歴記録 |
| **AuthUseCase.Signup** | ユーザー登録 | ユーザー作成 + メール検証トークン作成 |
| **AuthUseCase.VerifyEmail** | メール認証 | トークン使用済みにする + ユーザーの認証日時を設定 |
| **AuthUseCase.ConfirmPasswordReset** | パスワードリセット | トークン使用済みにする + パスワード更新 |

## 設計上のトレードオフ

この設計では以下のトレードオフを受け入れる。
- **デュアルアクセスパス**: UseCase は DI で注入されたリポジトリ（`uc.xxxRepo`）と TxManager 経由のリポジトリ（`repos.Xxx`）の2つの経路でリポジトリにアクセスする。TX 内では必ず `repos.Xxx` を使う必要があり、`uc.xxxRepo` を使うとトランザクション外の操作になる。
- **Repos 構造体のメンテナンス**: 新しいリポジトリを追加した場合、`Repos` 構造体と `TxManager.Do()` の実装の両方を更新する必要がある。

## 新しい TX 対象操作の追加手順

1. UseCase メソッド内で `uc.txm.Do(ctx, func(repos repository.Repos) error { ... })` を使用
2. コールバック内ではすべてのリポジトリ操作に `repos.Xxx` を使う（`uc.xxxRepo` は使わない）
3. エラーを返せば自動ロールバック、nil を返せばコミット

新しいリポジトリ interface を追加する場合は以下の手順で実装する。
1. `repository/transaction.go` の `Repos` 構造体にフィールドを追加
2. `infra/postgres/transaction.go` の `Do()` 内で新しいリポジトリのインスタンスを生成
