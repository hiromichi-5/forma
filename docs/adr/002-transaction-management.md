# トランザクション管理パターンの選定

## 背景・課題 (Background/Problem)

- クリーンアーキテクチャ移行（ADR-001）において、usecase 層でのトランザクション管理方式を決定する必要があった
- usecase は DB の実装詳細（`pgx.Tx` 等）に依存してはならない
- トランザクションが必要な操作が複数存在する（フォーム登録、招待受諾、デフォルトステータスを伴うステータス更新、チケット更新など）
- commit/rollback の漏れを構造的に防ぎたい

## 決定事項 (Decision)

**Unit of Work** パターンを採用する。

```go
// repository/uow.go
type UnitOfWork[T any] interface {
    Do(ctx context.Context, fn func(repos T) error) error
}

type FormRepos struct {
    Form   FormRepository
    Member MemberRepository
    Status StatusRepository
}
// AuthRepos, InviteRepos, StatusRepos, TicketRepos も同様に定義
```

Go Generics を用いた `UnitOfWork[T]` interface により、各 UseCase はトランザクション内で必要なリポジトリのみを含む型安全な Repos struct を受け取る。`fn` が error を返せば自動ロールバック、nil を返せばコミットされる。

現在の仕様の詳細は `docs/architecture/transaction.md` を参照。

### 理由 (Reasons)

- tx に参加する repository がコード上で明示的に見える（`r.Form.Create` と書いた時点で tx 内であることが自明）
- repository interface に tx の概念を持ち込まなくて良い
- commit/rollback が `Do` 内に一元管理され、漏れがない
- 各 UseCase が必要な repository のみ受け取る
- 新しいリポジトリ追加が影響する UseCase の Repos struct のみ変更すればよい

### 受け入れるトレードオフ (Accepted Trade-offs)

- **デュアルアクセスパス:** tx 不要な操作では `uc.formRepo`（DI 注入）、tx 必要な操作では `repos.Form`（UnitOfWork から取得）の2ルートが存在する
- **ユースケースごとの Repos struct:** UseCase ごとに Repos struct を定義するため、struct の数が UseCase 数に比例して増える
- **同一 UseCase 内の余剰フィールド:** per-usecase 粒度のため、一部のメソッドでは不要なリポジトリフィールドが Repos に含まれる場合がある

## 検討した別の選択肢 (Alternatives Considered)

### repository層での管理
- メリット: repository 層で完結するため、usecase はトランザクション管理を意識しなくて良い
- デメリット: repository のインターフェースにトランザクション管理の概念が入り込む。複数 repository を跨いだトランザクション管理が難しい。

**不採用理由:** repository 層でトランザクション管理の概念を持ち込むのは現在のアーキテクチャの方針に反するため、採用しない。

### usecase層での管理
- メリット: usecase 層で完結するため、トランザクション管理が明示的になる。複数 repository を跨いだトランザクション管理が容易。
- デメリット: usecaseにdbの実装詳細が入り込む。

**不採用理由:* usecase 層で DB の実装詳細に依存するのは現在のアーキテクチャの方針に反するため、採用しない。

### Context 埋め込み方式

```go
type Transaction interface {
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}
```

tx を `context.WithValue` で埋め込み、各 repository が暗黙的に取り出す。

- メリット: repository のインターフェースを変更せずにトランザクション管理が可能である。実装がシンプルである。
- デメリット: 暗黙的なトランザクション管理になる。　contextの乱用ともいえる。

**不採用理由:** 明示的なトランザクション管理を重視するため、採用しない。

## 参考 (References)

- https://tech.gree-x.com/golang-transaction-pattern/
- https://zenn.dev/hacobell_dev/articles/0ae114500cf974
- https://karamaru-alpha.com/posts/layered-tx/
- https://tech.yappli.io/entry/ddd_usecase
