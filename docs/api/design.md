# API 設計ガイド

## エンドポイント設計方針

### リソースベース REST

基本はリソース指向で設計する。URL がリソースを表し、HTTP メソッドが操作を表す。

```text
GET    /v1/forms              フォーム一覧
POST   /v1/forms              フォーム登録
GET    /v1/forms/:id          フォーム詳細
PATCH  /v1/forms/:id          フォーム更新
```

サブリソースは親リソースの下にネストする。

```text
GET    /v1/forms/:id/members
POST   /v1/forms/:id/statuses
DELETE /v1/forms/:id/invites/:id
```

### RPC スタイルの例外

アクションとしての性質が強い操作は RPC スタイルを許容する。

```text
POST /v1/auth/login                          ログイン
POST /v1/auth/signup                         ユーザー登録
POST /v1/forms/:id/sync                      フォーム同期
POST /v1/invites/:id/accept                  招待承諾
```

### チケットのフラットルーティング

チケットはフォームのサブリソースだが、URL はフラットにしている。

```text
GET   /v1/tickets?form_id=xxx          フォーム指定でチケット一覧
GET   /v1/tickets/:id                  チケット詳細
PATCH /v1/tickets/:id                  チケット更新
GET   /v1/tickets/:id/histories        変更履歴
```

チケット詳細・更新は `ticket_id` だけで一意に特定できるため、`/v1/forms/:id/tickets/:id` のようなネストは冗長であるため、フラットなパスにしている。

### 招待承諾のフラットルーティング

```text
POST /v1/invites/:id/accept
```

招待承諾は受信者の視点から行う操作であり、フォーム ID は招待データ内に含まれている。受信者がフォーム ID を知っている必要はないため、フラットなパスにしている。

## レスポンス DTO

レスポンス DTO は `handler/response.go` に定義する。以下の理由から、entity とは別の構造体としている。

- **HTTP 層の関心事**: JSON タグ、フィールドの省略（`omitempty`）、ネスト構造はクライアントとの契約であり entity が持つべきではない
- **バージョニング容易性**: API バージョンを変える場合、handler の DTO だけ変えれば済む
- **変換の明示性**: entity → DTO の変換関数がハンドラーに存在することで、何がクライアントに露出するかが明確になる

## エラーレスポンス

すべてのエラーは統一された形式で返す。詳細は [エラーハンドリング](../architecture/error-handling.md) を参照。

```json
{
  "code": "ERROR_CODE",
  "message": "人間向けメッセージ",
  "fields": [...]
}
```

## 新しいエンドポイント追加時のチェックリスト

- `openapi/openapi.yaml` にエンドポイントを定義
- `make openapi` で仕様とフロントエンドの型を生成
- UseCase にビジネスロジックを実装
- Handler でリクエストパース・レスポンス変換を実装
- 必要に応じてレスポンス DTO を `handler/response.go` に追加
- `cmd/api/main.go` でルーティングを登録（認証グループの内側/外側を判断）
- 権限要件を決定し、UseCase 内で `requireEditor` / `requireAdmin` を呼ぶ
