# 現行インフラの IaC 化（GCP / Supabase / Vercel / Cloudflare）

## 目的

現在コンソール・`gcloud` で手作業構築されている本番インフラ（Cloud Run, Supabase, Vercel, Cloudflare など）を Terraform で宣言的に管理し、再現性・レビュー可能性・変更履歴を確保する。
**構成そのものは変えずに**コード化することを第一段階のゴールとする。

### Non Goals

- アーキテクチャの変更（AWS への移行は `docs/design/iac-aws-ideal.md` で扱う）
- アプリケーションのデプロイ（イメージのビルド・リビジョン更新）を Terraform に移すこと。これは引き続き `.github/workflows/deploy.yaml` が担う
- ステージング環境の新設（ディレクトリ構成だけは追加できる形にしておく）
- シークレットの「値」を Terraform で管理すること

## 背景

### 現行構成（`docs/guides/deploy.md` と実装から読み取れる範囲）

```text
                     ┌─────────────┐
                     │  Cloudflare │  DNS（+ プロキシ）
                     └──────┬──────┘
              ┌─────────────┴──────────────┐
    forma.hiromichi.dev             api.forma.hiromichi.dev
              │                            │
       ┌──────┴──────┐           ┌─────────┴─────────┐
       │   Vercel    │           │  Cloud Run (api)  │◀── Cloud Run Job (migrate)
       │ (Frontend)  │           └─────────┬─────────┘
       └─────────────┘                     │
                           ┌───────────────┼────────────────┐
                           │               │                │
                   ┌───────┴──────┐ ┌──────┴──────┐ ┌───────┴───────┐
                   │  Supabase    │ │   Resend    │ │ Google Forms  │
                   │ (PostgreSQL) │ │  (メール)   │ │     API       │
                   └──────────────┘ └─────────────┘ └───────────────┘
```

| 要素 | 実装上の根拠 | 備考 |
| --- | --- | --- |
| Artifact Registry | `deploy.yaml` の `AR_REPOSITORY` | イメージタグは `GITHUB_SHA::7` |
| Cloud Run Service | `deploy.yaml` の `CLOUD_RUN_SERVICE` | `gcloud run services update --image` のみ実行。それ以外の設定は手作業 |
| Cloud Run Job | `deploy.yaml` の `CLOUD_RUN_MIGRATION_JOB` | 同一イメージを `/bin/migrate` で実行（goose） |
| Workload Identity Federation | `deploy.yaml` の `WIF_PROVIDER` / `WIF_SERVICE_ACCOUNT` | GitHub Actions → GCP のキーレス認証 |
| Google サービスアカウント鍵 | `main.go` が `/run/secrets/google_sa.json` を読む | Secret Manager のボリュームマウントと推測 |
| 環境変数 | `main.go` | `APP_ENV`, `PG_DSN`, `RESEND_*`, `FRONTEND_BASE_URL`, `ALLOWED_ORIGINS`, `COOKIE_DOMAIN`, `LOG_LEVEL`, `SWAGGER_*` |
| Vercel | `frontend/vercel.json` | SPA リライト。ビルド時に `VITE_API_URL` が必要 |
| Resend | `infra/resend` | 送信ドメインの SPF / DKIM レコードが Cloudflare にあるはず |

### アプリ側の制約（インフラ設定に効くもの）

- **単一インスタンス前提**: SSE のイベント配信（`infra/pubsub/memory.go`）とログインのレート制限（`infra/ratelimit/memory.go`）がインメモリ実装。`docs/design/sse-ticket-realtime.md` でも複数インスタンス対応は Non Goal。したがって Cloud Run の `max_instance_count = 1` をコードで明示的に固定する。
- **SSE の長時間接続**: `GET /v1/forms/:form_id/stream` は 30 秒ごとに `ping` を送る。Cloud Run のリクエストタイムアウト（最大 3600 秒）と Cloudflare プロキシのアイドルタイムアウト（100 秒）の範囲内に収まっている。
- **Cookie**: `forma.hiromichi.dev` と `api.forma.hiromichi.dev` は同一サイト（SameSite=Lax）。`COOKIE_DOMAIN` の値をドメイン設計と合わせて管理する必要がある。

## 概要

ツールは **Terraform**（OpenTofu 互換を維持）を採用する。必要なプロバイダ（`google`, `cloudflare`, `vercel`, `supabase`）がすべて公式またはベンダー公式で揃っており、1つのツールで複数クラウドを横断できるため。

既存リソースは作り直さず、Terraform 1.5+ の `import` ブロックで取り込む。`terraform plan` が差分ゼロになることを「現状をコード化できた」の完了条件とする。

責務の境界は以下のとおり。

| 管理対象 | 管理するもの |
| --- | --- |
| Terraform | リソースの存在・設定（スケーリング、IAM、ドメイン、環境変数の「名前と参照先」） |
| GitHub Actions（`deploy.yaml`） | イメージのビルド、マイグレーション実行、Cloud Run へのイメージ反映 |
| 手作業（`gcloud secrets versions add` 等） | シークレットの値 |

イメージタグは Actions が更新するため、Terraform 側では `lifecycle.ignore_changes` で無視し、両者が互いの変更を巻き戻さないようにする。

## 設計詳細

### ディレクトリ構成

```text
infra/
├── bootstrap/                # tfstate 用 GCS バケットと Terraform 実行用 SA（ローカルで一度だけ apply）
│   └── main.tf
├── modules/
│   ├── gcp-api/              # Artifact Registry, Cloud Run Service/Job, SA, Secret Manager
│   ├── gcp-github-oidc/      # WIF Pool/Provider, デプロイ用 SA と権限
│   ├── cloudflare-dns/       # DNS レコード（api, frontend, Resend の SPF/DKIM/DMARC）
│   ├── vercel-frontend/      # Vercel プロジェクト, ドメイン, 環境変数
│   └── supabase/             # Supabase プロジェクト設定
└── envs/
    └── prod/
        ├── backend.tf        # GCS backend
        ├── providers.tf
        ├── main.tf           # modules の呼び出し
        ├── imports.tf        # 既存リソースの import ブロック（取り込み完了後に削除）
        ├── variables.tf
        └── terraform.tfvars  # 非機密値のみ
```

`envs/stg/` を足せばステージングを同じモジュールで作れる形にしておく。

### State 管理

- バックエンド: GCS バケット（バージョニング有効、Uniform bucket-level access、パブリックアクセス防止）
- ロック: GCS バックエンドのネイティブロック
- State にはプロバイダが返す機密値（Supabase の DB パスワード等）が平文で入りうるため、バケットへのアクセスは Terraform 実行用 SA と管理者のみに絞る

### プロバイダとバージョン固定

```hcl
terraform {
  required_version = ">= 1.9"
  required_providers {
    google     = { source = "hashicorp/google",       version = "~> 6.0" }
    cloudflare = { source = "cloudflare/cloudflare",  version = "~> 5.0" }
    vercel     = { source = "vercel/vercel",          version = "~> 3.0" }
    supabase   = { source = "supabase/supabase",      version = "~> 1.0" }
  }
}
```

`.terraform.lock.hcl` はコミットする。バージョンは導入時点の最新安定版に合わせて調整する。

### GCP（`modules/gcp-api`）

| リソース | 主な設定 |
| --- | --- |
| `google_project_service` | `run`, `artifactregistry`, `secretmanager`, `iamcredentials`, `sts`, `forms` |
| `google_artifact_registry_repository` | Docker 形式。cleanup policy で直近 N 件のタグ以外を削除 |
| `google_service_account` (runtime) | Cloud Run 実行用。Secret Manager の `secretAccessor` のみ付与 |
| `google_secret_manager_secret` | `pg-dsn`, `resend-api-key`, `google-sa-json`, `swagger-password`。**値（version）は作らない** |
| `google_cloud_run_v2_service` | 下記参照 |
| `google_cloud_run_v2_job` (migrate) | 同イメージ、`command = ["/bin/migrate"]`、`PG_DSN` のみ注入、`max_retries = 0` |

Cloud Run Service の要点:

```hcl
resource "google_cloud_run_v2_service" "api" {
  name     = "forma-api"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.runtime.email
    timeout         = "3600s" # SSE 長時間接続のため

    scaling {
      min_instance_count = 0
      max_instance_count = 1  # インメモリ pubsub / rate limit のため固定（後述）
    }

    containers {
      image = var.initial_image # 初回のみ使用。以降は Actions が更新

      dynamic "env" {
        for_each = {
          APP_ENV           = "production"
          FRONTEND_BASE_URL = "https://forma.hiromichi.dev"
          ALLOWED_ORIGINS   = "https://forma.hiromichi.dev"
          COOKIE_DOMAIN     = var.cookie_domain
        }
        content {
          name  = env.key
          value = env.value
        }
      }

      env {
        name = "PG_DSN"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.pg_dsn.secret_id
            version = "latest"
          }
        }
      }
      # RESEND_API_KEY なども同様

      volume_mounts {
        name       = "google-sa"
        mount_path = "/run/secrets"
      }
    }

    volumes {
      name = "google-sa"
      secret {
        secret = google_secret_manager_secret.google_sa_json.secret_id
        items {
          version = "latest"
          path    = "google_sa.json"
        }
      }
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image, # deploy.yaml が更新する
      client, client_version,
    ]
  }
}
```

公開設定は `google_cloud_run_v2_service_iam_member`（`roles/run.invoker` を `allUsers`）で明示する。

### GitHub Actions 用 OIDC（`modules/gcp-github-oidc`）

- `google_iam_workload_identity_pool` / `_provider`
  - `attribute_condition = "assertion.repository == 'hiromichi-5/forma'"` でリポジトリを限定する（既存設定がブランチまで絞っていなければ `assertion.ref == 'refs/heads/main'` の追加も検討）
- デプロイ用 SA に付与するロール（最小権限）
  - `roles/artifactregistry.writer`（対象リポジトリに限定）
  - `roles/run.developer`（対象サービス・ジョブに限定）
  - `roles/iam.serviceAccountUser`（runtime SA に対してのみ）
- Terraform 実行用（plan/apply）には別の SA と Provider を用意し、apply は `main` ブランチ + GitHub Environment の承認に限定する

### Cloudflare（`modules/cloudflare-dns`）

| レコード | 内容 |
| --- | --- |
| `forma.hiromichi.dev` | Vercel 向け CNAME / A |
| `api.forma.hiromichi.dev` | Cloud Run 向け（ドメインマッピングなら `ghs.googlehosted.com` への CNAME） |
| Resend 用 | `send.` サブドメインの MX / SPF(TXT)、`resend._domainkey` の DKIM(TXT)、`_dmarc` |

ゾーン自体（`hiromichi.dev`）は他用途と共有している可能性が高いため、**ゾーンは data source で参照するだけにし、Forma 関連レコードのみ管理する**。

### Vercel（`modules/vercel-frontend`）

- `vercel_project`: `framework = "vite"`, `root_directory = "frontend"`, Git 連携
- `vercel_project_domain`: `forma.hiromichi.dev`
- `vercel_project_environment_variable`: `VITE_API_URL = https://api.forma.hiromichi.dev`（production）

### Supabase（`modules/supabase`）

- `supabase_project` を import して、リージョン・インスタンスサイズを記録する
- `supabase_settings` で API / Auth 設定のうち、Forma が使わない機能（Supabase Auth, Storage など）を明示的に無効化しておく
- DB の接続文字列（`PG_DSN`）は Supabase 側で発行されたものを手動で Secret Manager に投入する。Cloud Run からは **Supavisor（transaction mode, 6543）ではなく session mode か直接接続**を使う（pgx のプリペアドステートメントとの相性のため）。現状どちらを使っているかは要確認

### Resend

公式 Terraform プロバイダが無いため、Resend 側のドメイン登録は手作業のまま残し、**DNS レコードだけ Cloudflare モジュールで管理する**。API キーは Secret Manager に手動投入。

### シーケンス

#### 既存リソースの取り込み

```text
1. bootstrap を apply（tfstate 用バケット作成）
2. modules を書く（現状設定を gcloud / コンソールから写す）
3. imports.tf に import ブロックを書く
     import {
       to = module.api.google_cloud_run_v2_service.api
       id = "projects/<project>/locations/<region>/services/forma-api"
     }
4. terraform plan で差分を確認 → 差分がなくなるまでコード側を修正
5. terraform apply（実質 no-op）→ imports.tf を削除
```

`terraform plan -generate-config-out=generated.tf` で雛形を生成し、手で整形する方法も使える。

#### CI（`.github/workflows/infra.yaml` を新設）

```text
PR（infra/** 変更）       : fmt -check → validate → tflint → plan → PR にコメント
main マージ（infra/** 変更）: plan → GitHub Environment "prod-infra" の承認 → apply
```

既存の `deploy.yaml`（アプリデプロイ）とは独立させる。

## 検討した別の選択肢

- **Pulumi / CDKTF**: 型のある言語で書けるが、Go チームでも HCL のほうが情報量・レビューのしやすさで勝る。CDKTF は開発終了が発表されているため不採用
- **gcloud スクリプトのみ**: 宣言的でなく、ドリフト検出ができない
- **Cloud Run を Terraform で完全管理（イメージも Terraform で更新）**: デプロイのたびに Terraform apply が必要になり、マイグレーション → デプロイの順序制御が複雑になる。現行の Actions フローを活かすため不採用
- **Supabase を IaC 対象外にする**: プロバイダの成熟度は高くないが、プロジェクトのリージョン・サイズが記録されるだけでも価値があるため対象に含める

## 注意点・懸念事項

- **シークレット**: 値は Terraform に書かない。`google_secret_manager_secret_version` を使わず、`gcloud secrets versions add` で投入する。サービスアカウント鍵（`google-sa-json`）は長期鍵なので定期ローテーション手順を `docs/guides/deploy.md` に追記する
- **max_instance_count = 1 の意味**: 可用性はシングルポイント。Cloud Run の自動再起動はあるが、デプロイ時に一瞬 SSE が切断される（フロントは再接続する前提）。複数インスタンス化は `docs/design/iac-aws-ideal.md` の「アプリ側の前提変更」を参照
- **ドリフト**: コンソールでの手動変更は次回 plan で検出される。定期的に `terraform plan` を実行するスケジュールジョブ（週1）を CI に入れる
- **要確認事項**（取り込み前に実環境で確認する）
  - GCP リージョン、Cloud Run のサービス名・ジョブ名、現在の CPU/メモリ・`min/max instances`・タイムアウト
  - `api.forma.hiromichi.dev` が Cloud Run ドメインマッピングか、Cloudflare 経由のどちらで公開されているか
  - Supabase のリージョン（Cloud Run と同リージョンか）と接続方式（直接 / Supavisor）
  - `COOKIE_DOMAIN` の本番値
  - Swagger の Basic 認証を本番で有効にしているか
