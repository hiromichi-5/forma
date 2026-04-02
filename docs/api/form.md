# フォーム API

## 概要

Google Forms の登録・一覧・詳細取得・設定変更、質問一覧取得、フォーム同期を提供する。

## POST /v1/forms

Google Forms をシステムに登録する。登録者は Admin メンバーとなる。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/forms` |
| 認証 | 必要（SessionMiddleware） |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- | --- |
| `url` | string | Yes | Google Forms の URL またはフォーム ID |

```json
{
  "url": "https://docs.google.com/forms/d/e/1FAIpQL.../viewform"
}
```

### レスポンス

#### 201 Created

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string | 登録されたフォームの UUID |

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | URL 形式が不正 |
| `FORM_NOT_FOUND` | 404 | Google Forms API でフォームが見つからない |
| `FORM_NOT_SHARED` | 404 | フォームがサービスアカウントに共有されていない |
| `FORM_ALREADY_REGISTERED` | 409 | 同じフォームが既に登録されている |

### 補足

- URL からのフォーム ID 抽出は正規表現 `/forms/d/e/([a-zA-Z0-9_-]+)/` を使用
- 20文字以上でスラッシュを含まない文字列はフォーム ID として直接受け付ける
- フォームタイトルが空の場合、Google Form ID をタイトルとして使用する

## GET /v1/forms

自分がメンバーとして参加しているフォームの一覧を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/forms` |
| 認証 | 必要（SessionMiddleware） |

### レスポンス

#### 200 OK

```json
{
  "forms": [
    {
      "id": "...",
      "form_id": "1FAIpQL...",
      "title": "お問い合わせフォーム"
    }
  ]
}
```

---

## GET /v1/forms/:form_id

フォームの詳細を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/forms/:form_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

#### 200 OK

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string | フォーム UUID |
| `form_id` | string | Google Forms のフォーム ID |
| `title` | string | フォームタイトル |
| `description` | string? | フォームの説明 |
| `created_at` | string | 作成日時（RFC3339） |

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `RESOURCE_HIDDEN` | 404 | メンバーでない |
| `FORM_NOT_FOUND` | 404 | フォームが存在しない |

---

## PATCH /v1/forms/:form_id

フォームのタイトル質問 ID を設定する。設定された質問の回答がチケットのタイトルとして使用される。

| 項目 | 値 |
| --- | --- |
| メソッド | `PATCH` |
| パス | `/v1/forms/:form_id` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `title_question_id` | string? | No | タイトルに使う質問 ID。未指定なら変更なし、`null` で解除 |

### レスポンス

#### 204 No Content

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | 指定された質問 ID がフォームに存在しない、または空文字 |
| `RESOURCE_HIDDEN` | 404 | メンバーでない |
| `FORM_NOT_FOUND` | 404 | フォームが存在しない |

### 補足

- 空の `{}` の場合、既存設定は維持される
- `{"title_question_id": null}` のみ解除として扱う

---

## GET /v1/forms/:form_id/questions

フォームの質問一覧を取得する。

| 項目 | 値 |
| --- | --- |
| メソッド | `GET` |
| パス | `/v1/forms/:form_id/questions` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

#### 200 OK

```json
{
  "questions": [
    {
      "question_id": "abc123",
      "title": "お名前",
      "question_type": "TEXT",
      "options": []
    }
  ]
}
```

---

## POST /v1/forms/:form_id/sync

Google Forms から新しい回答を同期し、チケットとして取り込む。

| 項目 | 値 |
| --- | --- |
| メソッド | `POST` |
| パス | `/v1/forms/:form_id/sync` |
| 認証 | 必要（SessionMiddleware） |
| 権限 | Editor 以上 |

### レスポンス

#### 200 OK

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `synced` | bool | 常に `true` |
| `new_tickets` | int | 新規に作成されたチケット数 |
| `last` | string | 最後の同期日時（RFC3339）。新しい回答がない場合は空文字 |

```json
{
  "synced": true,
  "new_tickets": 5,
  "last": "2026-03-30T12:00:00Z"
}
```

### エラー

| コード | HTTP | 条件 |
| --- | --- | --- |
| `RESOURCE_HIDDEN` | 404 | メンバーでない |
| `FORM_NOT_FOUND` | 404 | フォームが存在しない |
| `FORM_NOT_SHARED` | 404 | フォームがサービスアカウントに共有されていない |

### 補足

- 命名がリソース指向でないが、フォーム単位での同期はリソースとして切り出しにくいため、リソース指向を逸脱することを許容している
- ただし、 `ticketRepo.Create` は `ON CONFLICT DO NOTHING` で実装されており、既存の ResponseID と重複した場合は挿入をスキップするため、何度呼んでも同じ最終状態になる冪等な操作を実現している
- 新しいチケットのデフォルトステータスはフォームのデフォルトステータスが適用される
- 優先度は `medium` で初期化される
