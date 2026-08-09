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

一覧のフィールドに加えて `answers` と `notifications` が含まれる。

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
  ],
  "notifications": [
    {
      "notification_type": "status_change",
      "last_sent_at": "2026-03-30T15:00:00Z"
    },
    {
      "notification_type": "assignee_assigned",
      "last_sent_at": null
    }
  ]
}
```

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `notifications` | array | 通知種別ごとの最終送信日時。常にすべての種別が含まれる |
| `notifications[].notification_type` | string | 通知種別（`status_change`, `assignee_assigned`） |
| `notifications[].last_sent_at` | string? | 最終送信日時（RFC3339）。未送信なら `null` |

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `RESOURCE_HIDDEN` | 404 | チケットが存在しない、メンバーでない |

### 補足

- `notifications` は再送ボタンの状態表示（「3分前に通知済み」、レートリミット中の非活性化）のために返す。自動送信・手動送信のいずれも含む

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

**200 OK** — 更新後のチケット詳細（GET /v1/tickets/:ticket_id と同じ形式）に `notifications` を加えたもの。

```json
{
  "id": "550e8400-...",
  "...": "（GET /v1/tickets/:ticket_id と同じフィールド）",
  "notifications": [
    { "notification_type": "status_change", "result": "sent" }
  ]
}
```

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `notifications` | array | この更新で自動送信を試みた通知の結果。試行がなければ空配列 |
| `notifications[].notification_type` | string | 通知種別（`status_change`, `assignee_assigned`） |
| `notifications[].result` | string | `sent`（送信成功）または `failed`（送信失敗） |

自動送信されるのは通知設定が `always` の種別のみ。`confirm` の種別はここでは送信されず、`POST /v1/tickets/:ticket_id/notifications` を別途呼ぶ必要がある。

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | priority が不正な値、assignee_id と null を同時指定 |
| `RESOURCE_HIDDEN` | 404 | チケットが存在しない、メンバーでない、指定ステータスが不正、担当者がメンバーでない |
| `USER_NOT_FOUND` | 404 | 指定担当者のユーザーが存在しない |

通知メールの送信失敗はエラーにしない。チケットの更新は成功として `200 OK` を返し、`notifications` の `result` で伝える。

### 補足

- `assignee_id` フィールドは JSON の `null`（担当者解除）と未指定（変更なし）を `nullableUUIDPayload` で判別する
- 変更履歴は `changeRecorder` が各フィールドの変更を蓄積し、トランザクション内で一括保存する
- 値が変わらない場合（同じステータスへの更新など）は変更履歴を記録しない
- 通知メールはトランザクションのコミット後、レスポンスを返す前に同期的に送信する。値が変わらず変更履歴が記録されない場合は通知も送信しない
- `tickets.respondent_email` が `null` のチケットは、設定に関わらず送信しない
- 担当者の解除（`assignee_id: null`）は通知の対象外
- 通知設定については `docs/api/notification.md`、設計の背景は `docs/design/respondent-notification.md` を参照

---

## POST /v1/tickets/:ticket_id/notifications

回答者へ通知メールを手動で送信する。`confirm` モードでの送信と、届かなかった場合の再送に使う。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/tickets/:ticket_id/notifications` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `notification_type` | string | Yes | 通知種別（`status_change`, `assignee_assigned`） |

```json
{
  "notification_type": "status_change"
}
```

### レスポンス

#### 200 OK

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `notification_type` | string | 送信した通知種別 |
| `sent_at` | string | 送信日時（RFC3339） |

```json
{
  "notification_type": "status_change",
  "sent_at": "2026-03-30T15:00:00Z"
}
```

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | notification_type が不正な値 |
| `RESOURCE_HIDDEN` | 404 | チケットが存在しない、メンバーでない |
| `NOTIFICATION_DISABLED` | 409 | 該当種別の通知設定が `off` |
| `RESPONDENT_EMAIL_MISSING` | 409 | チケットに回答者のメールアドレスがない |
| `NOTIFICATION_RATE_LIMITED` | 429 | 同一チケット・同一種別で5分以内に送信済み |

### 補足

- 送信内容は**送信時点のチケットの状態**（現在のステータス名・担当者名）。過去の特定の変更に紐づけては送らない
- `assignee_assigned` は現在の担当者が `null` の場合 `VALIDATION_ERROR` となる
- レートリミットの判定には `PATCH` による自動送信の記録も含まれる。直前に `always` の自動通知が送られていれば、手動送信も5分間は制限される
- 送信に成功した場合のみ `ticket_notifications` に記録する。送信に失敗した場合は記録されないため、すぐに再試行できる

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
