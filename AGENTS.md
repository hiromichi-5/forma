# Forma Backend API

Google Forms の回答を管理する Web アプリケーションです。

## タスクの実行ガイド

docs/* に詳細なドキュメントがある。タスクを実行する際は以下の手順を遵守する。

1. タスクに関連するドキュメントを読む
2. ドキュメントと照らし合わせながら現在の実装を理解する
3. タスクを実行する
4. ドキュメントの更新が必要な場合は更新する

- タスクを実行する前に、既存ドキュメントと既存実装が異なる場合は、ユーザに差異を説明して確認を取る。
- タスクを実行する際は、関連部分の実装の質だけでなく、**コードの一貫性を保つ**。既存コードとスタイルを変更する場合、ユーザに理由を説明する。

### ドキュメント構成

```text
docs
├── adr
│   ├── 001-clean-architecture.md
│   ├── 002-transaction-management.md
│   ├── 003-error-code-design.md
│   ├── 004-invite-email-failure-handling.md
│   └── TEMPLATE.md
├── api 
│   ├── auth.md
│   ├── design.md
│   ├── form.md
│   ├── invite.md
│   ├── member.md
│   ├── notification.md
│   ├── profile.md
│   ├── status.md
│   ├── TEMPLATE.md
│   └── ticket.md
├── architecture
│   ├── authorization.md
│   ├── backend-logging.md
│   ├── backend-testing.md
│   ├── error-handling.md
│   ├── overview.md
│   └── transaction.md
├── design
│   ├── respondent-notification.md
│   └── TEMPLATE.md
├── guides
│   ├── commands.md
│   ├── deploy.md
│   ├── getting-started.md
│   └── TEMPLATE.md
└── rules 
    └── TEMPLATE.md
```
