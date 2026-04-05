# デプロイガイド

## アーキテクチャ

```text
                     ┌─────────────┐
                     │  Cloudflare │
                     └──────┬──────┘
                            │
              ┌─────────────┼─────────────┐
              │             │             │
    forma.hiromichi.dev     │   api.forma.hiromichi.dev
              │             │             │
       ┌──────┴──────┐     │   ┌─────────┴─────────┐
       │   Vercel    │     │   │    Cloud Run       │
       │  (Frontend) │     │   │    (Backend)       │
       └─────────────┘     │   └─────────┬─────────┘
                           │             │
                           │   ┌─────────┴─────────┐
                           │   │    Supabase        │
                           │   │   (PostgreSQL)     │
                           │   └───────────────────┘
                           │
                     ┌─────┴─────┐
                     │  AWS SES  │
                     └───────────┘
```

## 前提

- GCP プロジェクトが作成済み
- `gcloud` CLI がインストール・認証済み
- Supabase アカウントがある
- Cloudflare で `hiromichi.dev` を管理している
- GitHub リポジトリに push 権限がある

## 1. Supabase

### プロジェクト作成

1. [Supabase Dashboard](https://supabase.com/dashboard) でプロジェクトを作成
2. リージョン: `Northeast Asia (Tokyo)` を推奨
3. 作成後、Settings → Database → Connection string (URI) から接続文字列を取得

接続文字列の形式:

```text
postgresql://postgres.[project-ref]:[password]@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres
```

**注意**: Connection Pooling（Transaction mode, ポート 6543）を使うこと。

## 2. GCP リソース

以下のコマンドで変数を設定してから進める:

```bash
export PROJECT_ID=your-gcp-project-id
export REGION=asia-northeast1
export AR_REPO=forma
export SERVICE_NAME=forma-api
export MIGRATION_JOB=forma-migrate
export SA_NAME=forma-deploy
```

### Artifact Registry

```bash
gcloud artifacts repositories create $AR_REPO \
  --repository-format=docker \
  --location=$REGION \
  --project=$PROJECT_ID
```

### Secret Manager

以下のシークレットを登録する:

```bash
# PostgreSQL 接続文字列（Supabase の接続文字列）
echo -n "postgresql://..." | gcloud secrets create pg-dsn --data-file=- --project=$PROJECT_ID

# Google Service Account JSON（Forms API 用）
gcloud secrets create google-sa --data-file=./secrets/google_sa.json --project=$PROJECT_ID

# AWS SES 認証情報
echo -n "your-access-key-id" | gcloud secrets create aws-access-key-id --data-file=- --project=$PROJECT_ID
echo -n "your-secret-access-key" | gcloud secrets create aws-secret-access-key --data-file=- --project=$PROJECT_ID
```

### Cloud Run サービス

```bash
gcloud run deploy $SERVICE_NAME \
  --region $REGION \
  --project $PROJECT_ID \
  --image "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}/api:latest" \
  --port 8080 \
  --allow-unauthenticated \
  --set-env-vars "APP_ENV=production" \
  --set-env-vars "FRONTEND_BASE_URL=https://forma.hiromichi.dev" \
  --set-env-vars "ALLOWED_ORIGINS=https://forma.hiromichi.dev" \
  --set-env-vars "COOKIE_DOMAIN=.hiromichi.dev" \
  --set-env-vars "AWS_REGION=ap-northeast-1" \
  --set-env-vars "SES_FROM_ADDRESS=no-reply@forma.hiromichi.dev" \
  --set-env-vars "SES_REPLY_TO_ADDRESS=support@forma.hiromichi.dev" \
  --set-secrets "PG_DSN=pg-dsn:latest" \
  --set-secrets "AWS_ACCESS_KEY_ID=aws-access-key-id:latest" \
  --set-secrets "AWS_SECRET_ACCESS_KEY=aws-secret-access-key:latest" \
  --set-secrets "/run/secrets/google_sa.json=google-sa:latest"
```

### Cloud Run Job（マイグレーション）

```bash
gcloud run jobs create $MIGRATION_JOB \
  --region $REGION \
  --project $PROJECT_ID \
  --image "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}/api:latest" \
  --command "/bin/migrate" \
  --set-env-vars "MIGRATION_DIR=/migrations" \
  --set-secrets "PG_DSN=pg-dsn:latest"
```

### Workload Identity Federation（GitHub Actions → GCP 認証）

```bash
# サービスアカウント作成
gcloud iam service-accounts create $SA_NAME \
  --display-name="Forma Deploy (GitHub Actions)" \
  --project=$PROJECT_ID

SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# 必要な権限を付与
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"

# Workload Identity Pool 作成
gcloud iam workload-identity-pools create "github" \
  --location="global" \
  --display-name="GitHub Actions" \
  --project=$PROJECT_ID

# OIDC Provider 作成
gcloud iam workload-identity-pools providers create-oidc "github-actions" \
  --location="global" \
  --workload-identity-pool="github" \
  --display-name="GitHub Actions" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --project=$PROJECT_ID

# サービスアカウントにバインド（リポジトリを指定）
REPO="hiromichi-5/forma"  # GitHub リポジトリ
POOL_ID=$(gcloud iam workload-identity-pools describe "github" \
  --location="global" --project=$PROJECT_ID --format="value(name)")

gcloud iam service-accounts add-iam-policy-binding $SA_EMAIL \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/${POOL_ID}/attribute.repository/${REPO}" \
  --project=$PROJECT_ID
```

### GitHub リポジトリの Variables 設定

Settings → Secrets and variables → Actions → Variables に以下を設定:

| Variable | 値の例 |
|---|---|
| `GCP_PROJECT_ID` | `your-project-id` |
| `GCP_REGION` | `asia-northeast1` |
| `AR_REPOSITORY` | `forma` |
| `CLOUD_RUN_SERVICE` | `forma-api` |
| `CLOUD_RUN_MIGRATION_JOB` | `forma-migrate` |
| `WIF_PROVIDER` | `projects/123456/locations/global/workloadIdentityPools/github/providers/github-actions` |
| `WIF_SERVICE_ACCOUNT` | `forma-deploy@your-project-id.iam.gserviceaccount.com` |

WIF_PROVIDER の値は以下で取得:

```bash
gcloud iam workload-identity-pools providers describe "github-actions" \
  --location="global" \
  --workload-identity-pool="github" \
  --project=$PROJECT_ID \
  --format="value(name)"
```

## 3. Cloudflare DNS

1. Cloudflare Dashboard → DNS → Records
2. CNAME レコードを追加:
   - Name: `api.forma`
   - Target: Cloud Run サービスのデフォルト URL（例: `forma-api-xxxxx-an.a.run.app`）
   - Proxy status: **DNS only**（オレンジの雲を OFF）

3. Cloud Run 側でカスタムドメインをマッピング:

```bash
gcloud run domain-mappings create \
  --service $SERVICE_NAME \
  --domain api.forma.hiromichi.dev \
  --region $REGION \
  --project $PROJECT_ID
```

**注意**: Cloud Run のドメインマッピングで TLS 証明書が自動プロビジョニングされるため、Cloudflare の Proxy（オレンジの雲）は OFF にする。

## 4. Vercel

Vercel Dashboard → Project Settings → Environment Variables:

| Variable | Value |
|---|---|
| `VITE_API_URL` | `https://api.forma.hiromichi.dev` |

設定後、再デプロイが必要。

## 5. 初回デプロイ

全てのリソースが作成できたら、main ブランチに push することで自動デプロイが走る。

手動でデプロイする場合:

```bash
# イメージビルド & プッシュ
IMAGE="${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}/api:manual"
docker build -t $IMAGE .
docker push $IMAGE

# マイグレーション実行
gcloud run jobs update $MIGRATION_JOB --image $IMAGE --region $REGION --project $PROJECT_ID
gcloud run jobs execute $MIGRATION_JOB --region $REGION --project $PROJECT_ID --wait

# デプロイ
gcloud run services update $SERVICE_NAME --image $IMAGE --region $REGION --project $PROJECT_ID
```

## 確認

```bash
# ヘルスチェック
curl https://api.forma.hiromichi.dev/healthz

# ブラウザで https://forma.hiromichi.dev にアクセスし、ログインが正常に動作することを確認
```
