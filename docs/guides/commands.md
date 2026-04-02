# コマンドガイド

## 概要

このガイドでは、プロジェクトで使用される主要なコマンドとその使用方法について説明します。

## コマンド

このプロジェクトでは`/makefile`に主要なコマンドが定義されています。  

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
