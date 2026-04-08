# バックエンド ログ設計方針

## 目的

Backend APIのログの設計方針を定める。

## 方針

### 基本原則

1. **slog標準ライブラリを使用する**
2. **JSON Lines形式で標準出力に出力する**
3. **ローカルではText形式を許容する** （開発時の可読性を向上させるため）
4. **機密情報をログに含めない**

### ログレベル運用

| レベル | 用途 | ローカル | 本番 |
| --- | --- | --- | --- |
| ERROR | リクエスト処理の予期せぬ失敗 | o | o |
| WARN | 将来的に問題になりうる状態 | o | o |
| INFO | アクセスログ、起動ログ、重要なビジネスイベント | o | o |
| DEBUG | 開発時の詳細情報（ドメインエラーの詳細など） | o | - |

- `LOG_LEVEL` 環境変数で切り替え可能（デフォルト: `INFO`）
- WARNは現時点で使用していない

## アーキテクチャ

### loggerの伝搬方式

`context.Context`の`Value`を使い、request_id付きのloggerをミドルウェアからusecase層まで伝搬する。

```text
RequestLoggerミドルウェア
  ├─ request_id付きloggerを生成
  └─ context.WithValue でリクエストのcontextに格納
        │
        ▼
SessionMiddleware
  ├─ logger.From(ctx) でloggerを取得
  └─ user_id付きに拡張して再格納
        │
        ▼
Handler層 (gin.Context)
  ├─ handler は c を usecase に渡す
  │   ※ *gin.Context は context.Context を実装しているため直接渡せる
  └─ handleError で logger.From(ctx) を使いエラーログ出力
        │
        ▼
UseCase層 (context.Context)
  ├─ logger.From(ctx) で request_id + user_id 付きloggerを取得
  └─ ビジネスイベントのINFOログを出力
        │
        ▼
Repository層 (context.Context)
  └─ 基本的にログ出力しない（エラーは呼び出し元に返す）
```

### ログ出力キー名規則

OTel Semantic Conventionsに準拠しつつ、slogの標準キー（`time`, `level`, `msg`）はそのまま使用する。

| キー名 | 付与タイミング | 例 |
| --- | --- | --- |
| `service.name` | 起動時 | `"forma-api"` |
| `deployment.environment` | 起動時 | `"local"` |
| `request_id` | リクエスト受信時 | `"550e8400-..."` |
| `user_id` | 認証成功時 | `"7c9e6679-..."` |
| `http.request.method` | レスポンス時 | `"POST"` |
| `url.path` | レスポンス時 | `"/v1/forms"` |
| `http.response.status_code` | レスポンス時 | `200` |
| `http.server.request.duration` | レスポンス時 | `"12.345ms"` |
| `client.address` | レスポンス時 | `"192.168.1.1"` |
| `error` | エラー発生時 | エラーメッセージ |
| `form_id` | フォーム操作時 | `"550e8400-..."` |

## ログ出力例

### JSON形式での出力例

```jsonc
// 起動時
{"time":"2026-04-02T13:43:25.208263+09:00","level":"INFO","msg":"application startup initiated","service.name":"forma-api","deployment.environment":"production","env":"production","log_level":"DEBUG"}
{"time":"2026-04-02T13:43:25.210354+09:00","level":"INFO","msg":"database connection established","service.name":"forma-api","deployment.environment":"production"}
{"time":"2026-04-02T13:43:25.21257+09:00","level":"INFO","msg":"server listening","service.name":"forma-api","deployment.environment":"production","addr":":8080"}
// 認証成功時
{"time":"2026-04-02T13:43:34.230049+09:00","level":"INFO","msg":"user authenticated","service.name":"forma-api","deployment.environment":"production","user_id":"278b249c-7903-48c6-9b3a-4d4b09223e6c"}
{"time":"2026-04-02T13:43:34.230186+09:00","level":"INFO","msg":"request completed","service.name":"forma-api","deployment.environment":"production","request_id":"99802b40-58ee-43a5-941c-aef70d99571f","http.request.method":"POST","url.path":"/v1/auth/login","http.response.status_code":200,"http.server.request.duration":"202.853958ms","client.address":"::1"}
// フォーム登録時
{"time":"2026-04-02T13:44:45.112922+09:00","level":"INFO","msg":"form registered","service.name":"forma-api","deployment.environment":"production","form_id":"dd849826-f03d-45f3-b415-60c950beae25","google_form_id":"1Bw87YOacw4nXtAsSAQJWl-KLNM1QTrvwnxpYauoTIWQ"}
{"time":"2026-04-02T13:44:45.113162+09:00","level":"INFO","msg":"request completed","service.name":"forma-api","deployment.environment":"production","request_id":"e3da1eec-b00c-4d84-8183-40bde3357ab1","http.request.method":"POST","url.path":"/v1/forms","http.response.status_code":201,"http.server.request.duration":"2.034762917s","client.address":"::1","user_id":"278b249c-7903-48c6-9b3a-4d4b09223e6c"}
// ドメインエラー発生時
{"time":"2026-04-02T13:45:12.128997+09:00","level":"DEBUG","msg":"domain error","service.name":"forma-api","deployment.environment":"production","request_id":"46e35d3a-1ab6-4553-9950-e271316b5454","user_id":"278b249c-7903-48c6-9b3a-4d4b09223e6c","code":"RESOURCE_HIDDEN","status":404}
{"time":"2026-04-02T13:45:12.133032+09:00","level":"INFO","msg":"request completed","service.name":"forma-api","deployment.environment":"production","request_id":"46e35d3a-1ab6-4553-9950-e271316b5454","http.request.method":"GET","url.path":"/v1/tickets","http.response.status_code":404,"http.server.request.duration":"26.324042ms","client.address":"::1","user_id":"278b249c-7903-48c6-9b3a-4d4b09223e6c"}
```

### TEXT形式の出力例

```text
// 起動時
time=2026-04-02T13:18:10.754+09:00 level=INFO msg="application startup initiated" service.name=forma-api deployment.environment=local env=local log_level=DEBUG
time=2026-04-02T13:18:10.756+09:00 level=INFO msg="database connection established" service.name=forma-api deployment.environment=local
time=2026-04-02T13:18:10.756+09:00 level=INFO msg="server listening" service.name=forma-api deployment.environment=local addr=:8080
// ドメインエラー発生時
time=2026-04-02T13:28:39.495+09:00 level=DEBUG msg="domain error" service.name=forma-api deployment.environment=local request_id=0d10c205-ce5a-4962-a210-39e714af98db user_id=278b249c-7903-48c6-9b3a-4d4b09223e6c code=RESOURCE_HIDDEN status=404
time=2026-04-02T13:28:39.500+09:00 level=INFO msg="request completed" service.name=forma-api deployment.environment=local request_id=0d10c205-ce5a-4962-a210-39e714af98db http.request.method=GET url.path=/v1/tickets http.response.status_code=404 http.server.request.duration=42.254792ms client.address=::1 user_id=278b249c-7903-48c6-9b3a-4d4b09223e6c
```

## 実装例

usecase層では以下のように記述することで、自動的にrequest_idとuser_idが付与されたloggerを取得できる。

```go
log := logger.From(ctx)
log.Info("form sync started", "form_id", formID.String())
```

## 参考

- <https://future-architect.github.io/arch-guidelines/documents/forLog/>
