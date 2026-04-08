# クリーンアーキテクチャの採用

## 背景・課題 (Background/Problem)

- `Service` 構造体が forms, tickets, members, invites, statuses, sync のすべてを抱える God Object になっていた
- ハンドラーの依存パターンが不統一（interface 経由と直接参照が混在）
- `pgtype.UUID`, `pgtype.Text` などの DB 型がサービス層のビジネスロジックに侵食していた
- 各ハンドラーでドメインエラー → HTTP ステータスの switch 文が重複していた
- ハンドラーにビジネスロジックが漏れていた（URL パース等）
- 複数クエリを伴う操作（フォーム登録 + ロール作成 + ステータス作成）にトランザクションが適用されていなかった

## 決定事項 (Decision)

以下の5層構成に移行する。

```text
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

各層の詳細な責務は `docs/architecture/overview.md` を参照。

### 理由 (Reasons)

- God Object の分割により、各 usecase が自身に必要な repository のみに依存し、テスタビリティと変更容易性が向上する
- entity 層で pgtype を排除することで、ビジネスロジックが DB 実装の詳細から独立する
- エラーミドルウェアにより、ハンドラーごとのエラーマッピング重複を解消できる
- repository interface を通じた依存性逆転により、infra の差し替え（テスト用 mock、DB 移行等）が容易になる

### 受け入れるトレードオフ (Accepted Trade-offs)

- ファイル数が増える
- entity ↔ DB モデルの変換コードが infra/postgres に必要（pgtype ↔ Go 標準型）
- entity ↔ レスポンス DTO の変換コードが handler に必要

## 検討した別の選択肢 (Alternatives Considered)

### 選択肢1: 現構造の改善のみ（Service 分割 + エラーミドルウェア）

Service を分割し、エラーミドルウェアを追加するだけの最小限の変更。

- メリット: 変更量が小さい、既存テストへの影響が少ない
- デメリット: pgtype の侵食が残る、DB 層との結合が残る

## 参考 (References)

## 議論 (Discussion)

### Google Forms クライアントの interface の置き場所

- 案A（採用）: `repository/` に `FormFetcher` interface を置く。外部依存の interface が1箇所にまとまる
- 案B: usecase 内で定義。interface が分散する。
- 案C: 独立パッケージ `gateway/` を作る。1つの interface のためにパッケージを作る過剰さがある

### レスポンス DTO の置き場所

- 案A（採用）: `handler/` に置く。レスポンスの形は HTTP 層の関心事。API バージョニング時に handler だけ変えれば済む。
- 案B: usecase の戻り値型として定義。handler が薄くなるが usecase が暗黙的にレスポンスの形を意識してしまう。
- 案C: entity に置く。ドメインモデルとレスポンスが混在してしまう。

### entity に何を置くか

- 結論: entity はドメインモデルのみ。ビュー型（TicketSummary, TicketDetail 等）は usecase 層に定義する
- 根拠: TicketSummary は複数エンティティの結合 + アプリケーションロジックで導出した表示用データであり、ドメインの概念ではない。
