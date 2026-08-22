# アーキテクチャ概観

## レイヤー構成

```text
backend/internal/
├── app/             ルーティングと DI のワイヤリング
├── entity/          ドメインモデル・ドメインエラー
├── repository/      データアクセス interface
├── usecase/         ビジネスロジック
├── interfaces/
│   ├── handler/     HTTP ハンドラー・レスポンス DTO
│   └── middleware/  認証ミドルウェア
└── infra/
    ├── postgres/    repository 実装・型変換
    ├── google/      Google Forms API クライアント
    ├── resend/      Resend メール送信
    ├── pubsub/      チケット更新イベントの配信（SSE 用）
    └── db/          sqlc 生成コード
```

## 依存の方向

```text
handler → usecase → repository → entity
usecase → usecase（interface 経由のみ）
middleware → repository
infra/postgres → repository
infra/google → repository
infra/resend → repository
infra/pubsub → usecase
```

すべての依存は内側（entity）に向かう。外側の層（handler, infra）が内側の層（usecase, repository, entity）に依存し、逆方向の依存は存在しない。`infra/postgres` と `infra/google` は repository interface を実装することで依存性逆転を実現している。

usecase 同士の依存は、**呼び出す側が必要な操作だけを interface として定義する**ことで具象への依存を避けている。現時点では2箇所ある。

| 呼び出す側 | interface | 実装 |
| --- | --- | --- |
| `FormUseCase` | `FormSyncer` | `SyncUseCase` |
| `TicketUseCase` | `TicketNotifier` | `NotificationUseCase` |

`EventPublisher`（`TicketUseCase` → `infra/pubsub`）も同じ形で、usecase 側が interface を持ち infra が実装する。

## 各層の責務

| 層 | やること | やらないこと |
| --- | --- | --- |
| **entity** | Go 標準型でのドメインモデル定義、ドメインエラー定義、**自身で完結する不変条件と状態遷移** | **IO を伴う判断**、複数エンティティの結合、DB 型、JSON タグ |
| **repository** | データアクセス・外部 API の interface 定義 | 実装、SQL |
| **usecase** | ユースケースの手順、権限チェック、トランザクション境界、複数エンティティの結合と表示用の導出 | HTTP、DB 型、レスポンス整形 |
| **handler** | リクエストパース、entity → DTO 変換、エラー → HTTP ステータス変換 | ビジネスロジック |
| **middleware** | セッション検証、認証済みユーザー ID の注入 | 業務エラー変換 |
| **infra/postgres** | repository 実装、pgtype ↔ Go 型変換、トランザクション管理 | ビジネスロジック |
| **infra/google** | Google Forms API 呼び出し | ビジネスロジック |
| **infra/resend** | Resend メール送信 | ビジネスロジック |
| **infra/pubsub** | チケット更新イベントの配信 | ビジネスロジック |
| **infra/db** | sqlc 生成コード（自動生成、手動変更なし） | — |

### entity と usecase の境界

判断の材料が**そのエンティティの中だけで揃うか**で分ける。

```go
// entity: 自身の値だけで答えられる
func (r Role) CanEdit() bool                                        // ロールの値だけで決まる
func (p Priority) Valid() bool                                      // 優先度の値だけで決まる
func (t *Ticket) ChangeStatus(status *FormStatus) (uuid.UUID, bool) // 自身の状態遷移

// usecase: 他のエンティティや IO が要る
requireEditor(ctx, memberRepo, formID, userID)  // メンバーシップの問い合わせが要る
deriveTitle(titleQID, answers, questions, ...)  // フォーム・質問・回答の結合
resolveUserName(ctx, repos.User, userID)        // 表示名の解決に IO が要る
```

ビュー型（`TicketSummary`, `TicketDetail` 等）を usecase に置く方針は変わらない。複数エンティティの結合とアプリケーションロジックで導出した表示用データであり、ドメインの概念ではないため。

詳細な経緯は `docs/adr/001-clean-architecture.md` と `docs/design/domain-modeling.md` を参照。

## DI（依存性注入）

`internal/app/setup.go` の `NewRouter` でアプリケーション全体のワイヤリングを行う。`cmd/api/main.go` は環境変数の読み込みと外部クライアント（`infra/google`, `infra/resend`）の生成までを担い、それらを `app.Deps` として渡す。現時点では以下のように手動でDIをしているが、将来的に Wire などのDIフレームワークを導入する可能性もある。

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
