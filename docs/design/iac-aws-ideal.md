# AWS 上の理想構成と IaC 設計

## 目的

Forma を AWS 上で運用する場合の理想的な構成と、それを Terraform で管理する設計をまとめる。
現行構成（Cloud Run / Supabase / Vercel）の IaC 化は `docs/design/iac-current-infra.md` を参照。本書はそこからの移行先候補として位置付ける。

### Non Goals

- 移行の実施判断（本書は比較・判断材料）
- マイクロサービス化や Kubernetes（EKS）の採用。現行のモノリス 1 コンテナを前提とする
- マルチリージョン構成

## 背景

### 現行アプリの特性（構成を決める要件）

| 特性 | 実装 | AWS 構成への影響 |
| --- | --- | --- |
| Go の単一バイナリ API + マイグレーション用バイナリ | `Dockerfile`（distroless） | コンテナ実行基盤が必要。マイグレーションはワンショットタスク |
| SSE の長時間接続（30 秒ごとに ping） | `handler/stream.go` | **Lambda / API Gateway / App Runner は不向き**（タイムアウト・ストリーミング制約）。ALB + ECS が素直 |
| インメモリの pubsub / レート制限 | `infra/pubsub/memory.go`, `infra/ratelimit/memory.go` | そのままでは 1 タスク固定。冗長化するにはアプリ改修が必要（後述） |
| PostgreSQL（pgx, goose, sqlc） | `infra/postgres` | RDS for PostgreSQL がそのまま使える |
| セッションは DB 保存、Cookie は `SameSite=Lax` | `app/setup.go` | ステートレスに水平スケール可能（pubsub / rate limit を除く） |
| 外部 API: Google Forms API, Resend | `infra/google`, `infra/resend` | プライベートサブネットからのアウトバウンド経路（NAT）が必要 |
| Google のサービスアカウント鍵ファイルを読む | `main.go`, `google/form_client.go` | 鍵を Secrets Manager に置くか、AWS → GCP の Workload Identity 連携でキーレス化 |
| フロントは Vite の SPA | `frontend/` | S3 + CloudFront の静的配信で十分 |

## 概要

```text
                         Route 53（または Cloudflare DNS）
                 ┌──────────────────┴───────────────────┐
       forma.hiromichi.dev                     api.forma.hiromichi.dev
                 │                                      │
        ┌────────┴────────┐                    ┌────────┴────────┐
        │   CloudFront    │                    │  ALB (+ WAF)    │  public subnet × 2AZ
        │   + WAF         │                    └────────┬────────┘
        └────────┬────────┘                             │
          OAC    │                             ┌────────┴────────┐
        ┌────────┴────────┐                    │  ECS Fargate    │  private subnet × 2AZ
        │  S3 (SPA)       │                    │  api タスク      │──┐
        └─────────────────┘                    └───┬─────────┬───┘  │ NAT GW
                                                   │         │      ▼
                                   ┌───────────────┘         │   Google Forms API
                                   ▼                         ▼   Resend
                         ┌──────────────────┐   ┌──────────────────┐
                         │ RDS PostgreSQL   │   │ Secrets Manager  │
                         │ (Multi-AZ)       │   │ / SSM Parameter  │
                         └──────────────────┘   └──────────────────┘
                           isolated subnet

   GitHub Actions ──OIDC──▶ IAM Role ──▶ ECR push / ECS RunTask(migrate) / ECS UpdateService
```

- **コンピュート**: ECS on Fargate。ALB のアイドルタイムアウト（既定 60 秒）は ping 間隔 30 秒で維持できるが、余裕を見て 120 秒に設定する
- **DB**: RDS for PostgreSQL（Graviton, gp3）。本番は Multi-AZ。ECS からのみ到達可能な isolated subnet に置く
- **フロント**: S3 + CloudFront（OAC）。SPA のリライトは CloudFront Functions で行う（`vercel.json` 相当）
- **DNS / 証明書**: ACM（ALB 用は ap-northeast-1、CloudFront 用は us-east-1）。DNS は Cloudflare のままでも Route 53 に寄せてもよい（後述）
- **IaC**: Terraform。アカウント・ネットワークなどの「土台」とアプリ用リソースを state 分割して管理する

## 設計詳細

### アカウント・環境

AWS Organizations で用途ごとにアカウントを分ける（小規模でも後から分けるのは大変なため最初から分ける）。

| アカウント | 用途 |
| --- | --- |
| management | Organizations, IAM Identity Center, 請求 |
| forma-stg | ステージング |
| forma-prod | 本番 |

人のアクセスは IAM Identity Center（SSO）のみ。IAM ユーザーとアクセスキーは作らない。

### Terraform 構成

```text
infra/aws/
├── bootstrap/            # state 用 S3 バケット、GitHub OIDC Provider、Terraform 実行ロール
├── modules/
│   ├── network/          # VPC, subnet, NAT, VPC endpoint, flow logs
│   ├── database/         # RDS, パラメータグループ, SG, 自動バックアップ
│   ├── api/              # ECR, ECS cluster/service/taskdef, ALB, ACM, SG, IAM, ロググループ
│   ├── migration/        # マイグレーション用タスク定義（ECS RunTask で実行）
│   ├── frontend/         # S3, CloudFront, CloudFront Function, ACM(us-east-1)
│   ├── secrets/          # Secrets Manager / SSM の「入れ物」
│   ├── observability/    # CloudWatch アラーム, ダッシュボード, SNS
│   └── github-oidc/      # Actions 用 IAM ロール（デプロイ / plan / apply）
└── envs/
    ├── stg/
    │   ├── network/      # state 分割: 変更頻度とブラスト半径で分ける
    │   ├── data/         # database, secrets
    │   └── app/          # api, migration, frontend, observability
    └── prod/
        └── （同上）
```

- **state**: S3 バックエンド（バージョニング・SSE-KMS・パブリックアクセスブロック）。ロックは S3 ネイティブロック（`use_lockfile = true`）を使い DynamoDB は不要
- **state 分割**: `network` → `data` → `app` の順に依存。下位の出力は `terraform_remote_state` ではなく **SSM Parameter Store 経由**か data source（タグ検索）で参照し、state 間の結合を緩くする
- **環境差分**: `envs/<env>/*/terraform.tfvars` のみで表現する（インスタンスサイズ、Multi-AZ、タスク数）
- **プロバイダ**: `hashicorp/aws ~> 6.0`（CloudFront 用に `us-east-1` の alias provider を併用）。DNS を Cloudflare に残す場合は `cloudflare/cloudflare` も使う

### ネットワーク（`modules/network`）

| 項目 | 設定 |
| --- | --- |
| VPC | /16、2 AZ（ap-northeast-1a, 1c） |
| public subnet | ALB, NAT Gateway |
| private subnet | ECS タスク（外向き通信は NAT 経由） |
| isolated subnet | RDS（インターネット経路なし） |
| NAT Gateway | prod は AZ ごと、stg は 1 個に節約 |
| VPC Endpoint | ECR (api/dkr), S3 (gateway), Secrets Manager, CloudWatch Logs。NAT のデータ処理料金削減とイメージ pull の高速化 |
| VPC Flow Logs | S3 へ（保持 90 日） |

セキュリティグループは「ALB ← 0.0.0.0/0:443」「ECS ← ALB SG:8080」「RDS ← ECS SG:5432（＋マイグレーションタスク SG）」の参照関係のみで定義する。

### API（`modules/api`）

| リソース | 主な設定 |
| --- | --- |
| ECR | immutable tag、push 時スキャン、ライフサイクルで直近 30 イメージ保持 |
| ECS Cluster | Container Insights 有効 |
| Task Definition | `cpu=256, memory=512`（初期値）、ARM64（Graviton）、`readonlyRootFilesystem=true`、ログは awslogs（JSON の slog をそのまま） |
| ECS Service | Fargate、`desired_count` は後述、デプロイサーキットブレーカー＋自動ロールバック、`enable_execute_command` は stg のみ |
| ALB | HTTPS のみ（HTTP→HTTPS リダイレクト）、`idle_timeout = 120`、ターゲットのヘルスチェック `/health`、`deregistration_delay = 30` |
| WAF (ALB) | AWS マネージドルール（Common, KnownBadInputs）＋ `/v1/auth/*` へのレートベースルール |
| IAM | 実行ロール（ECR pull、Secrets 読み取り、Logs）とタスクロール（アプリが使う AWS API のみ。現状ほぼ不要）を分離 |

シークレットはタスク定義の `secrets` で Secrets Manager から注入する。

| 環境変数 | 取得元 |
| --- | --- |
| `PG_DSN` | Secrets Manager（RDS マネージドシークレットから組み立て。後述） |
| `RESEND_API_KEY`, `SWAGGER_PASSWORD` | Secrets Manager（値は手動投入） |
| `APP_ENV`, `FRONTEND_BASE_URL`, `ALLOWED_ORIGINS`, `COOKIE_DOMAIN`, `RESEND_FROM_ADDRESS` 等 | タスク定義の `environment`（Terraform 変数） |
| Google 認証情報 | 下記「Google Forms API の認証」 |

アプリのデプロイは Terraform ではなく GitHub Actions が行う（現行と同じ責務分担）。Terraform 側は `aws_ecs_service` の `task_definition` を `ignore_changes` にし、Actions が「タスク定義の新リビジョン登録 → サービス更新」を行う。

#### タスク数とアプリ側の前提変更

インメモリ実装のままでは `desired_count = 1`（＋デプロイ時は `minimum_healthy_percent=100, maximum_percent=200` で一時的に 2）。この状態では 2 タスクが重なる数十秒間、片方に繋いだ SSE クライアントへイベントが届かない可能性がある点に注意。

理想形（冗長化して `desired_count = 2` 以上、AZ 分散）にするには以下のアプリ改修が前提となる。

| 対象 | 現状 | 改修案 | 追加インフラ |
| --- | --- | --- | --- |
| SSE イベント配信 (`EventPublisher`) | `MemoryHub` | **PostgreSQL `LISTEN/NOTIFY` 実装**に差し替え | なし（推奨） |
| ログインのレート制限 | `ratelimit.memory` | DB テーブルでの固定窓カウント、または WAF のレートベースルールに寄せる | なし |

`EventPublisher` は interface 化済み（`docs/design/sse-ticket-realtime.md`）のため差し替えは infra 層に閉じる。Redis（ElastiCache Serverless）でも実現できるが、イベント量が小さく追加コストと運用対象が増えるため、まずは Postgres で十分。

### マイグレーション（`modules/migration`）

- API と同じイメージで `command = ["/bin/migrate"]` を指定したタスク定義を別に用意する
- GitHub Actions から `aws ecs run-task` → `aws ecs wait tasks-stopped` → 終了コード確認、の順で実行し、成功時のみサービスを更新する（現行の Cloud Run Job と同じフロー）
- マイグレーションタスクには DDL 権限を持つ DB ユーザーを使い、API には DML のみのユーザーを使うのが理想（`PG_DSN` を分ける）

### データベース（`modules/database`）

| 項目 | stg | prod |
| --- | --- | --- |
| エンジン | PostgreSQL（現行と同じメジャーバージョン） | 同左 |
| インスタンス | `db.t4g.micro` | `db.t4g.small`〜 |
| Multi-AZ | なし | あり |
| ストレージ | gp3 20GB、自動拡張上限設定 | 同左 |
| 暗号化 | KMS | KMS |
| バックアップ | 7 日 | 14〜35 日、PITR |
| 削除保護 | なし | あり（`deletion_protection = true`, `prevent_destroy`） |
| 認証情報 | `manage_master_user_password = true`（Secrets Manager で自動ローテーション） | 同左 |
| パラメータ | `rds.force_ssl = 1`, `log_min_duration_statement` | 同左 |
| Performance Insights | 有効（無料枠） | 有効 |

マスターユーザーはアプリに使わない。アプリ用ユーザーとマイグレーション用ユーザーは初回に手動（またはブートストラップ用タスク）で作成し、その DSN を Secrets Manager に入れる。

Aurora Serverless v2 も候補だが、最小 ACU の常時課金で RDS の小さいインスタンスより高くつき、現状の負荷規模ではメリットが薄いため RDS を推奨する。

### Google Forms API の認証

| 方式 | 内容 | 評価 |
| --- | --- | --- |
| A. SA 鍵を Secrets Manager に保存 | 現行と同じ。JSON 全体をシークレットにし、起動時にファイルとして書き出す（distroless のため環境変数で受けるようアプリ改修するのが簡単） | 移行が最小。ただし長期鍵が残る |
| B. **GCP Workload Identity Federation（AWS プロバイダ）** | ECS タスクロールの認証情報で GCP の SA を借用する。鍵ファイルが不要になる | **推奨**。`google/form_client.go` の `option.WithAuthCredentialsFile(option.ServiceAccount, ...)` を `external_account` 形式も受けられる形に変更する必要あり |

B の場合、GCP 側の WIF Pool / Provider も同じ Terraform（`google` プロバイダ）で管理する。

### フロントエンド（`modules/frontend`）

- S3: パブリックアクセスブロック、OAC 経由のみ許可、バージョニング
- CloudFront: HTTP/3、`index.html` は `no-cache`、`/assets/*` は長期キャッシュ（Vite のハッシュ付きファイル名）
- CloudFront Function（viewer-request）: 拡張子のないパスを `/index.html` に書き換え（SPA リライト）
- レスポンスヘッダーポリシーで HSTS, CSP, X-Frame-Options などを付与
- デプロイは Actions で `vite build` → `aws s3 sync` → CloudFront invalidation（`/index.html` のみ）

#### 代替: 同一オリジン化

CloudFront のビヘイビアで `/v1/*` を ALB オリジンに振り分けると、フロントと API が同一オリジンになり CORS と `COOKIE_DOMAIN` が不要になる。ただし SSE が CloudFront を経由することになる（CloudFront のオリジン応答タイムアウト・キャッシュ無効化設定が必要）ため、初期は現行どおり `api.` サブドメイン分離を維持し、必要になったら検討する。

### DNS

- **Cloudflare 継続**（推奨・初期）: `hiromichi.dev` ゾーンを Forma 以外でも使っているなら移さない。ALB / CloudFront への CNAME と ACM の DNS 検証レコードを `cloudflare` プロバイダで作る。Cloudflare プロキシは CloudFront / WAF と役割が重複するため、Forma のレコードは **DNS only** にする
- **Route 53 へ委任**: `forma.hiromichi.dev` サブドメインだけ Route 53 に NS 委任すれば、ACM 検証・Alias レコードを AWS 内で完結できる

### メール

アプリが Resend に依存しているため **Resend 継続**を基本とする。Amazon SES に寄せる場合は `infra/resend` と同じ interface（`repository` 側）で SES 実装を追加し、SES のドメイン検証・DKIM・設定セット・バウンス通知（SNS）を Terraform で管理する。サンドボックス解除申請が必要。

### 可観測性（`modules/observability`）

| 種別 | 内容 |
| --- | --- |
| ログ | CloudWatch Logs（API 30 日、マイグレーション 90 日）。必要に応じて Logs Insights のクエリを保存 |
| メトリクス / アラーム | ALB 5xx 率・ターゲット応答時間、ECS の稼働タスク数 < 期待値、CPU / メモリ、RDS の CPU・空きストレージ・接続数 |
| 通知 | SNS → メール（または Chatbot → Slack） |
| 監査 | CloudTrail（Organizations のトレイル）、AWS Config（任意） |
| コスト | AWS Budgets で月額上限アラート |

### CI/CD

```text
.github/workflows/
├── ci.yaml          # 既存
├── deploy.yaml      # 既存を AWS 向けに書き換え
│                    #   OIDC → ECR push → RunTask(migrate) → タスク定義登録 → UpdateService → wait services-stable
├── deploy-fe.yaml   # vite build → s3 sync → invalidation
└── infra.yaml       # PR: fmt/validate/tflint/trivy(config)/plan コメント
                     # main: plan → Environment 承認 → apply（envs/<env>/<layer> ごと、変更があったものだけ）
```

GitHub OIDC の IAM ロールは用途別に 3 つに分け、信頼ポリシーの `sub` 条件で絞る。

| ロール | `sub` 条件 | 権限 |
| --- | --- | --- |
| deploy | `repo:hiromichi-5/forma:ref:refs/heads/main` | ECR push、ECS RunTask / RegisterTaskDefinition / UpdateService、S3 sync、CloudFront invalidation、`iam:PassRole`（対象ロールのみ） |
| tf-plan | `repo:hiromichi-5/forma:pull_request` | ReadOnly + state バケット読み取り |
| tf-apply | `repo:hiromichi-5/forma:environment:prod-infra` | 管理対象リソースの変更権限 |

### 概算コスト（prod、ap-northeast-1、月額・税抜の目安）

| 項目 | 目安 |
| --- | --- |
| ECS Fargate（0.25 vCPU / 0.5GB, ARM, 2 タスク常時） | 約 $15 |
| ALB | 約 $20〜25 |
| NAT Gateway（2 AZ） | 約 $90 ＋データ処理料（1 AZ にすれば約 $45） |
| VPC Interface Endpoint（4 種 × 2 AZ） | 約 $55（NAT 経由トラフィックが少ないなら省略してコスト優先も可） |
| RDS `db.t4g.small` Multi-AZ + gp3 20GB | 約 $60〜70 |
| CloudFront / S3 / WAF / Secrets Manager / CloudWatch | 約 $15〜25 |
| **合計** | **約 $200〜280** |

正確な値は導入時点の料金表で再見積もりする。**NAT と VPC エンドポイントが固定費の大半**を占めるため、規模に応じて下記「コスト重視案」を検討する。

#### コスト重視案（同じ Terraform モジュールの変数で切り替え）

- ECS タスクを public subnet に置き `assign_public_ip = true`（インバウンドは SG で ALB からのみ許可）→ NAT・Interface Endpoint を廃止
- RDS は Single-AZ（自動バックアップ + PITR で RPO を担保）
- この場合の合計は月 $50〜70 程度

### 現行構成からの移行手順（概略）

```text
1. bootstrap / network / data / app を stg に構築し、E2E（backend/e2e）を通す
2. prod を構築（DNS はまだ切り替えない）
3. Supabase から pg_dump → RDS に pg_restore（メンテナンス画面 or 読み取り専用化して短時間停止）
4. Secrets を投入、Google 認証方式を切り替え
5. DNS の TTL を事前に短縮 → api. / forma. のレコードを ALB / CloudFront に向ける
6. 1〜2 週間並行で旧環境を停止状態のまま保持 → 問題なければ GCP / Supabase / Vercel を削除
```

セッションは DB に保存されているため、データ移行後もユーザーは再ログイン不要。SSE は再接続で復帰する。

## 検討した別の選択肢

- **App Runner**: 最も手軽だが、リクエストタイムアウトが 120 秒固定で SSE の長時間接続に向かない。VPC 内 RDS への接続も VPC コネクタが必要
- **Lambda + API Gateway / Function URL**: Gin アプリのままでは載せにくく、SSE（長時間ストリーミング）と DB コネクションプールの扱いが課題
- **EKS**: 単一サービスにはオーバースペック。コントロールプレーンだけで月 $70 以上
- **EC2 + Docker Compose**: 安価だが、OS パッチ・デプロイ・冗長化を自前で持つことになる
- **Aurora Serverless v2**: 上記「データベース」参照。負荷が大きく変動するようになったら再検討
- **AWS CDK**: AWS 単独なら有力だが、Cloudflare DNS や GCP 側の WIF 設定を同じツールで扱えない。現行構成の IaC（Terraform）とツールを揃えるメリットも大きい
- **ElastiCache（Redis）で pubsub / rate limit**: 上記「タスク数とアプリ側の前提変更」参照

## 注意点・懸念事項

- **シークレットの値を Terraform に入れない**: Secrets Manager の「入れ物」だけ作り、値は CLI で投入する。RDS のマスターパスワードは `manage_master_user_password` で Terraform state に残らないようにする
- **SSE とデプロイ**: ローリング更新時に旧タスクの SSE 接続は切断される（`deregistration_delay` 経過後）。フロントの再接続処理が前提
- **単一タスク運用中の可用性**: アプリ改修前は AZ 障害・タスク再起動時に短時間停止する。Multi-AZ RDS だけでは API の可用性は上がらない
- **コールドスタートの消失とコスト増**: Cloud Run の scale-to-zero に比べ、常時稼働の固定費が発生する。利用規模が小さいうちは現行構成のほうが圧倒的に安い
- **リージョン間レイテンシ**: Google Forms API と Resend へのアウトバウンドはインターネット経由。同期処理（`POST /v1/forms/:form_id/sync`）のレイテンシは現行と同程度を想定するが、移行時に計測する
