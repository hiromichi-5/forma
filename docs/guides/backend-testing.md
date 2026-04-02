# バックエンドテストの設計方針

## 目的

Backend API のテストの設計方針を定める。

## テストレイヤー別の方針

### Unit Test

- 外部依存を持たない純粋なロジックを対象とする。Parametrized Test による網羅的な検証はこのレイヤーで行う。
- バリデーションや計算ロジックが追加された場合は、境界値を含む Parametrized Test で網羅する

### Integration Test

usecase 層と infra/postgres 層を対象に、テスト用 DB に接続してテストする。

#### usecase 層

- テストの効果を確保するため**DB はモックしない**。テスト用 DB に接続した実 Repository を注入する
- ケースによって期待する振る舞いやデータの前提が大きく変わるため、Table Driven Test ではなく、**サブテスト関数でケースを分ける**
- 正常系 + **準正常系（仕様範囲内の異常）に焦点**を絞る
  - 権限不足、リソース未存在など
  - 仕様範囲外の異常（DB接続エラー等）は大域的エラーハンドリングに任せるため、あえてテストしない
- 認可ロジック（`requireAdmin`, `requireEditor`）はこのレイヤーで検証する

以下のような実装を想定している。

    ```go
    func TestTicketUseCase_UpdateStatus(t *testing.T) {
        t.Run("正常系: admin がステータスを変更できること", func(t *testing.T) { // 日本語でケースを説明する
            // arrange: テスト用DBにデータ投入
            // act: usecase 呼び出し
            // assert: 結果検証
        })
        t.Run("準正常系: editor は権限不足でエラーになること", func(t *testing.T) {
            // ...
        })
        t.Run("準正常系: 存在しないチケットはエラーになること", func(t *testing.T) {
            // ...
        })
    }
    ```

#### infra/postgres 層

- 単純な CRUD は usecase の Integration Test で間接的にカバーされるため、あえてテストしない。
- 今後、複雑なクエリを実装する場合は、テストを追加する可能性がある。

### Backend E2E

`httptest.NewServer` + Gin ルーター全体を起動し、テスト用 DB + モック外部サービスを注入する。
実際の HTTP リクエスト/レスポンスでシナリオを検証する。


## テストインフラ

### テスト用 DB

- **testcontainers-go** で PostgreSQL コンテナを起動する
- `TestMain` でコンテナ起動 → goose でマイグレーション適用
- テスト関数ごとにテーブルを **TRUNCATE** してデータを分離する

### テスト用 DI

以下の設計により、プロダクションとテストでルーティングを共通化することで、テストの信頼性を高める。

```go
// internal/app/setup.go
type Deps struct {
    Pool    *pgxpool.Pool
    Fetcher repository.FormFetcher  // テスト時にモック差し替え
}

type Option struct {
    CookieSecure bool
}

func NewRouter(deps Deps, opt Option) *gin.Engine {
    // Repository → UseCase → Handler → ルーティングの全配線
}
```

プロダクション側:
```go
// cmd/api/main.go
func run() error {
    // 環境変数読み込み、pool/fetcher 生成
    r := app.NewRouter(app.Deps{Pool: pool, Fetcher: fetcher}, app.Option{...})
    return r.Run(addr)
}
```

テスト側:
```go
func setupTestRouter(t *testing.T, pool *pgxpool.Pool) *gin.Engine {
    return app.NewRouter(app.Deps{
        Pool:    pool,
        Fetcher: &mockFormFetcher{},
    }, app.Option{CookieSecure: false})
}
```


### テストライブラリ

- `github.com/stretchr/testify` — assert/require
- `github.com/testcontainers/testcontainers-go` — PostgreSQL コンテナ管理

# 参考
- https://developers.freee.co.jp/entry/testing-strategy-based-on-software-architecture
