# セットアップ手順

## 概要

このガイドでは、プロジェクトのローカル開発環境をセットアップするための手順を説明します。

## 前提条件

## 手順

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

### （5. データベースのマイグレーション）

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

`ok`が返れば成功です。
