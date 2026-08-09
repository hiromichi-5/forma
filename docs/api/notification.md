# 通知設定 API

## 概要

フォームごとの回答者向けメール通知設定の取得・変更を提供する。通知は種別ごとに「常時通知・毎回確認・通知しない」と通知内容の粒度を設定できる。設計の詳細は `docs/design/respondent-notification.md` を参照。

### 通知種別

| 値 | トリガー |
| --- | --- |
| `status_change` | チケットのステータスが変更されたとき |
| `assignee_assigned` | 担当者が新規に割り当てられたとき（解除は対象外） |

### モード

| 値 | 挙動 |
| --- | --- |
| `always` | 変更のたびに自動送信する |
| `confirm` | 変更のたびに操作者へ確認し、許可された場合のみ送信する |
| `off` | 送信しない |

---

## GET /v1/forms/:form_id/notification-settings

フォームの通知設定を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/forms/:form_id/notification-settings` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

#### 200 OK

```json
{
  "email_collection_type": "DO_NOT_COLLECT",
  "settings": [
    {
      "notification_type": "status_change",
      "mode": "confirm",
      "include_detail": true
    },
    {
      "notification_type": "assignee_assigned",
      "mode": "off",
      "include_detail": false
    }
  ]
}
```

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `email_collection_type` | string? | フォームのメールアドレス収集設定。未同期なら `null` |
| `settings[].notification_type` | string | 通知種別（`status_change`, `assignee_assigned`） |
| `settings[].mode` | string | モード（`always`, `confirm`, `off`） |
| `settings[].include_detail` | bool | `true` なら変更後のステータス名・担当者名をメールに含める |

`email_collection_type` は Google Forms から取得した値をそのまま返す。

| 値 | 意味 |
| --- | --- |
| `DO_NOT_COLLECT` | メールアドレスを収集しない。通知は届かない |
| `VERIFIED` | サインイン中のアカウントから自動収集 |
| `RESPONDER_INPUT` | 回答者が入力する項目で収集 |
| `EMAIL_COLLECTION_TYPE_UNSPECIFIED` / `null` | 不明（同期がまだ行われていない場合を含む） |

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | form_id が不正な UUID |
| `RESOURCE_HIDDEN` | 404 | メンバーでない |
| `FORM_NOT_FOUND` | 404 | フォームが存在しない |

### 補足

- 変更は Admin のみだが、取得は Editor 以上に許可している。フロントエンドが `confirm` モードの確認ダイアログを表示するか判断するために設定を知る必要があるため
- 未設定の種別も `mode: "off"`, `include_detail: false` として補完して返す。常にすべての種別が含まれる
- `email_collection_type` が `DO_NOT_COLLECT` の場合、設定画面で「このフォームは回答者のメールアドレスを収集していないため、通知は送信されません」と警告する。設定の保存自体は許可する

---

## PATCH /v1/forms/:form_id/notification-settings

フォームの通知設定を変更する。

| 項目 | 値 |
| --- | --- |
| メソッド | `PATCH` |
| パス | `/v1/forms/:form_id/notification-settings` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Admin のみ |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `settings` | array | Yes | 更新する設定の配列。含まれない種別は変更されない |
| `settings[].notification_type` | string | Yes | 通知種別 |
| `settings[].mode` | string | Yes | モード |
| `settings[].include_detail` | bool | Yes | 詳細を含めるか |

```json
{
  "settings": [
    {
      "notification_type": "status_change",
      "mode": "always",
      "include_detail": true
    }
  ]
}
```

### レスポンス

**200 OK** — 更新後の全設定（`GET` と同じ形式）

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | notification_type / mode が不正な値、同一種別の重複指定 |
| `RESOURCE_HIDDEN` | 404 | 操作者がメンバーでない |
| `FORBIDDEN` | 403 | 操作者が Admin でない |
| `FORM_NOT_FOUND` | 404 | フォームが存在しない |

### 補足

- 設定は `form_notification_settings` に UPSERT される
- フォーム登録時に設定行は作成されない。これにより既存フォームで意図せず通知が始まることがない
