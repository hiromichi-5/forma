# アーキテクチャ概観

## レイヤー構成

```
backend/internal/
├── entity/          ドメインモデル・ドメインエラー
├── repository/      データアクセス interface
├── usecase/         ビジネスロジック
├── interfaces/
│   ├── handler/     HTTP ハンドラー・レスポンス DTO
│   └── middleware/  認証ミドルウェア
└── infra/
    ├── postgres/    repository 実装・型変換
    ├── google/      Google Forms API クライアント
    └── db/          sqlc 生成コード
```

## 依存の方向

```
handler → usecase → repository (interface)
                  → entity

middleware → repository (interface)

infra/postgres → repository (interface) を実装
               → entity（pgtype ↔ Go 型変換）
               → infra/db（sqlc）

infra/google → repository (interface) を実装
```

すべての依存は内側（entity）に向かう。外側の層（handler, infra）が内側の層（usecase, repository, entity）に依存し、逆方向の依存は存在しない。`infra/postgres` と `infra/google` は repository interface を実装することで依存性逆転を実現している。

## 各層の責務

| 層 | やること | やらないこと |
|---|---|---|
| **entity** | Go 標準型でのドメインモデル定義、ドメインエラー定義 | ビジネスロジック、DB 型、JSON タグ |
| **repository** | データアクセス・外部 API の interface 定義 | 実装、SQL |
| **usecase** | ビジネスロジック、権限チェック、トランザクション境界 | HTTP、DB 型、レスポンス整形 |
| **handler** | リクエストパース、entity → DTO 変換、エラー → HTTP ステータス変換 | ビジネスロジック |
| **middleware** | セッション検証、認証済みユーザー ID の注入 | 業務エラー変換 |
| **infra/postgres** | repository 実装、pgtype ↔ Go 型変換、トランザクション管理 | ビジネスロジック |
| **infra/google** | Google Forms API 呼び出し | ビジネスロジック |
| **infra/db** | sqlc 生成コード（自動生成、手動変更なし） | — |

## DI（依存性注入）

`cmd/api/main.go` でアプリケーション全体のワイヤリングを行う。現時点では以下のように手動でDIをしているが、将来的に Wire などのDIフレームワークを導入する可能性もある。

```go
// --- Repository ---
userRepo := postgres.NewUserRepository(pool)
formRepo := postgres.NewFormRepository(pool)
// ...

// --- UseCase ---
authUC := usecase.NewAuthUseCase(userRepo, postgres.NewAuthUoW(pool))
formUC := usecase.NewFormUseCase(
    formRepo,
    memberRepo,
    statusRepo,
    fetcher,
    postgres.NewFormUoW(pool),
)
// ...

// --- Handler ---
ah := handler.NewAuthHandler(authUC, cookieCfg)
fh := handler.NewFormHandler(formUC)
// ...
```


## ルーティング

認証が不要なエンドポイント（`/v1/auth/*`）はルーターに直接登録する。認証が必要なエンドポイントは `SessionMiddleware` を適用したグループに登録する。

```go
// 認証不要
r.POST("/v1/auth/login", ah.PostV1AuthLogin)
r.POST("/v1/auth/signup", ah.PostV1AuthSignup)

// 認証必要
authz := r.Group("/v1")
authz.Use(middleware.SessionMiddleware(userRepo, cookieCfg.Name))
authz.GET("/forms", fh.GetV1Forms)
authz.POST("/forms", fh.PostV1Forms)
// ...
```
