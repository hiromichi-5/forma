## セットアップ手順
### 1. リポジトリをクローン
```bash
git clone [リポジトリのURL]
cd forma
```

### 2. 開発ツールをインストール
以下のコマンドラインツールを利用します。

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 3. 環境変数の設定
ローカル開発用の環境変数ファイルを作成してください。

```bash
cp .env.example .env.local
```

### 4. データベースの起動
Docker ComposeでPostgreSQLを起動します。

```bash
docker compose up -d db
```

### 5. データベースマイグレーション
```bash
# direnv を利用していない場合は、最初に環境変数を読み込む
# export $(grep -v '^#' .env.local | xargs)

make migrate
```

### 6. 開発サーバーの起動
```bash
make dev
```

### 7. 動作確認
別ターミナルで以下を実行します。

```bash
curl http://localhost:8080/healthz
```

`OK`が返れば成功です。

---

## 開発フロー

このプロジェクトでは**Makefile**に主要なコマンドが定義されています。  
以下の手順に従って開発を進めてください。

### API 仕様の変更（oapi-codegen）
1. `openapi/openapi.yaml` を編集  
2. コード再生成
   ```bash
   make openapi
   ```

### データベース構造の変更（goose）
1. 新しいマイグレーションファイルを作成
   ```bash
   goose -dir backend/migrations create add_new_feature sql
   ```
2. ファイルを編集  
3. マイグレーションを適用
   ```bash
   make migrate
   ```

### データベースクエリの変更（sqlc）
1. `backend/internal/db/queries/`内の`.sql`ファイルを編集  
2. コード再生成
   ```bash
   make sqlc
   ```

### テストの実行
```bash
make test
```
