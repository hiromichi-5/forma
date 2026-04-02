# チケット API

## 概要

チケットの一覧・詳細・更新および変更履歴の取得を提供する。チケットは `POST /v1/forms/:form_id/sync` で Google Forms の回答から自動的に作成される。

---

## GET /v1/tickets

フォームのチケット一覧を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/tickets` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### クエリパラメータ

| パラメータ | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `form_id` | string (UUID) | Yes | フォーム ID |
| `status_id` | string (UUID) | No | ステータスで絞り込み |

### レスポンス

#### 200 OK

```json
{
  "tickets": [
    {
      "id": "550e8400-...",
      "form_id": "660e8400-...",
      "form_title": "お問い合わせフォーム",
      "response_id": "ABC123",
      "respondent_email": "respondent@example.com",
      "status": {
        "id": "770e8400-...",
        "name": "未対応",
        "color": "#E53935"
      },
      "priority": "medium",
      "title_question_id": "qid_001",
      "title": "商品について質問があります",
      "assignee": {
        "id": "880e8400-...",
        "email": "staff@example.com",
        "display_name": "田中花子"
      },
      "submitted_at": "2026-03-30T08:00:00Z",
      "created_at": "2026-03-30T12:00:00Z"
    }
  ]
}
```

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | form_id パラメータが不正な UUID |
| `RESOURCE_HIDDEN` | 404 | メンバーでない、status_id が不正 |

### 補足

- `formContext` を利用することで、フォーム・ステータス・メンバー・質問の参照データを1回のクエリで取得し、N+1 問題を回避している
- チケットのタイトルは `deriveTitle` で、設定された質問 → デフォルト質問 → 他の質問 → フォームタイトル → レスポンス ID の順にフォールバックする

---

## GET /v1/tickets/:ticket_id

チケットの詳細（回答データ含む）を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/tickets/:ticket_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

#### 200 OK

一覧のフィールドに加えて `answers` が含まれる。

```json
{
  "id": "550e8400-...",
  "form_id": "660e8400-...",
  "form_title": "お問い合わせフォーム",
  "response_id": "ABC123",
  "respondent_email": "respondent@example.com",
  "status": { "id": "...", "name": "未対応", "color": "#E53935" },
  "priority": "medium",
  "title_question_id": "qid_001",
  "title": "商品について質問があります",
  "assignee": null,
  "submitted_at": "2026-03-30T08:00:00Z",
  "created_at": "2026-03-30T12:00:00Z",
  "answers": [
    {
      "question_id": "qid_001",
      "question_title": "お名前",
      "question_type": "TEXT",
      "values": ["山田太郎"],
      "display_value": "山田太郎"
    },
    {
      "question_id": "qid_002",
      "question_title": "お問い合わせ内容",
      "question_type": "PARAGRAPH_TEXT",
      "values": ["商品Aの在庫について"],
      "display_value": "商品Aの在庫について"
    }
  ]
}
```

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `RESOURCE_HIDDEN` | 404 | チケットが存在しない、メンバーでない |

---

## PATCH /v1/tickets/:ticket_id

チケットのステータス・担当者・優先度を更新する。変更は履歴に記録される。

| 項目 | 値 |
| --- | --- |
| メソッド | `PATCH` |
| パス | `/v1/tickets/:ticket_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### リクエストボディ

すべてオプショナル。指定されたフィールドのみ更新される。

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `status_id` | string? | No | 新しいステータス UUID |
| `assignee_id` | string / null | No | 新しい担当者 UUID。`null` で担当者を解除 |
| `priority` | string? | No | 新しい優先度（`high`, `medium`, `low`） |

```json
{
  "status_id": "770e8400-...",
  "assignee_id": "880e8400-...",
  "priority": "high"
}
```

担当者を解除する場合:

```json
{
  "assignee_id": null
}
```

### レスポンス

**200 OK** — 更新後のチケット詳細（GET /v1/tickets/:ticket_id と同じ形式）

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | priority が不正な値、assignee_id と null を同時指定 |
| `RESOURCE_HIDDEN` | 404 | チケットが存在しない、メンバーでない、指定ステータスが不正、担当者がメンバーでない |
| `USER_NOT_FOUND` | 404 | 指定担当者のユーザーが存在しない |

### 補足

- `assignee_id` フィールドは JSON の `null`（担当者解除）と未指定（変更なし）を `nullableUUIDPayload` で判別する
- 変更履歴は `changeRecorder` が各フィールドの変更を蓄積し、トランザクション内で一括保存する
- 値が変わらない場合（同じステータスへの更新など）は変更履歴を記録しない

## GET /v1/tickets/:ticket_id/histories

チケットの変更履歴を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/tickets/:ticket_id/histories` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

#### 200 OK

```json
{
  "histories": [
    {
      "id": "990e8400-...",
      "ticket_id": "550e8400-...",
      "changed_by": "880e8400-...",
      "changed_by_name": "田中花子",
      "field_name": "status",
      "old_value": "未対応",
      "new_value": "対応中",
      "created_at": "2026-03-30T14:00:00Z"
    }
  ]
}
```

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string | 履歴エントリの UUID |
| `ticket_id` | string | チケット UUID |
| `changed_by` | string? | 変更者のユーザー UUID |
| `changed_by_name` | string | 変更者の表示名 |
| `field_name` | string | 変更されたフィールド（`status`, `assignee`, `priority`） |
| `old_value` | string? | 変更前の値 |
| `new_value` | string? | 変更後の値 |
| `created_at` | string | 変更日時（RFC3339） |

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `RESOURCE_HIDDEN` | 404 | チケットが存在しない、メンバーでない |
