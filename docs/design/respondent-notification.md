# 回答者へのメール通知

## 目的

チケットのステータスが変更されたとき、および担当者が割り当てられたときに、Google Forms の回答者へメールで通知する。フォームの管理者が通知種別ごとに「常時通知・毎回確認・通知しない」と、通知内容の粒度を選べるようにする。あわせて、届かなかった場合などに操作者が手動で通知を再送できるようにする。

### Non Goals

- チャット機能に関する通知（将来的に通知種別を追加する余地は残すが、本設計では実装しない）
- 同期（`POST /v1/forms/:form_id/sync`）で新規作成されたチケットの通知
- 担当者を解除（未割当に戻す）したときの通知
- 優先度変更の通知
- 回答者向けのチケット状況確認ページ、および通知メールからそこへ誘導するリンク
- 回答者による配信停止（opt-out）手段
- フォームのメンバー（Admin / Editor）への通知。本設計の通知先は回答者のみ

## 背景

- チケットは `POST /v1/forms/:form_id/sync` で Google Forms の回答から自動作成される
- チケットのステータス・担当者・優先度の変更は `PATCH /v1/tickets/:ticket_id` に集約されており、変更は `ticket_histories` に記録される
- メール送信基盤は `repository.EmailSender` と `infra/resend` に既にある。テンプレートは `infra/resend/templates/<name>/{subject.txt,body.html,body.txt}` を `replaceVars` で単純置換する方式で、条件分岐を持たない
- 回答者のメールアドレスは `tickets.respondent_email` に保存されるが、Google Forms がメールを収集していない場合は `NULL` になる（`usecase/sync.go`）
- `forms.email_collection_type` カラムは存在するが、現時点ではどこからも書き込まれておらず常に `NULL` である。本設計で Google Forms API から取得して埋める

## 概要

フォームごと・通知種別ごとに通知設定を持つ `form_notification_settings` テーブルを新設する。設定は「モード」（`always` / `confirm` / `off`）と「内容の粒度」（`include_detail`）の2軸を持つ。

通知の送信経路は2つある。`always`（常時通知）は管理者が定めたポリシーであり、クライアントの挙動に依存してはならないため、`PATCH /v1/tickets/:ticket_id` の処理内でサーバーが自動送信する。`confirm`（毎回確認）の送信と手動の再送は `POST /v1/tickets/:ticket_id/notifications` に統一する。

送信の記録は `ticket_notifications` テーブルに残す。この履歴がレートリミットの判定材料になり、同時に「最終通知日時」の表示にも使える。手動送信は同一チケット・同一種別につき5分に1通までに制限し、回答者の受信箱が短時間に埋まることを防ぐ。

回答者のメールアドレスが存在しないチケットは、設定に関わらず送信しない。またフォームがそもそもメールアドレスを収集していない場合は、`forms.email_collection_type` を根拠に設定画面で警告を表示する。

## 設計詳細

### 通知種別と設定

| 通知種別 | トリガー | `include_detail = true` | `include_detail = false` |
| --- | --- | --- | --- |
| `status_change` | チケットのステータスが実際に変更されたとき（全遷移が対象） | 新しいステータス名を含める | ステータスが更新されたことのみ伝える |
| `assignee_assigned` | 担当者が新規に割り当てられたとき（解除は対象外） | 担当者名を含める | 担当者が割り当てられたことのみ伝える |

モードは3種類。

| モード | 自動送信 | 手動送信・再送 |
| --- | --- | --- |
| `always` | する | できる |
| `confirm` | しない（操作者へ確認し、許可されれば手動送信の経路で送る） | できる |
| `off` | しない | できない |

`off` は「この種別の通知を送らない」という管理者のポリシーとして扱い、手動送信も拒否する。Editor がポリシーを迂回できないようにするため。

ステータスごとの通知有無（例:「対応完了」だけ通知する）は設定できない。全遷移が一律に対象となる。

### email_collection_type の取得と利用

`google.golang.org/api v0.291.0` の `forms.FormSettings.EmailCollectionType` から取得する。

| 値 | 意味 | 通知の可否 |
| --- | --- | --- |
| `DO_NOT_COLLECT` | メールアドレスを収集しない | 送信できない（警告表示の対象） |
| `VERIFIED` | サインイン中のアカウントから自動収集 | 送信できる |
| `RESPONDER_INPUT` | 回答者が入力する項目で収集 | 送信できる |
| `EMAIL_COLLECTION_TYPE_UNSPECIFIED` / `NULL` | 不明（未同期のフォームを含む） | 不明として扱い、警告しない |

書き込み箇所は2つ。

1. **フォーム登録時** — `usecase/form.go` の `Form.Create` に含める
2. **同期時** — `usecase/sync.go` が同期のたびに `GetForm` を呼んでいるため、ここで更新する。質問以外も取り込むようになったため、この関数は `refreshFormQuestions` から `refreshFormMetadata` に改名した

2 により、既存の登録済みフォーム（現在すべて `NULL`）は次回の同期で自動的に埋まる。バックフィル用のマイグレーションやバッチは不要。

利用箇所は設定画面の警告表示のみとする。`DO_NOT_COLLECT` の場合に「このフォームは回答者のメールアドレスを収集していないため、通知は送信されません」と表示するが、**設定の保存自体は許可する**。Google Forms 側で収集を有効にして同期すれば、設定を変えずにそのまま動作するため。

実際に送信するかどうかの判定には使わない。送信可否はチケットごとの `respondent_email` の有無で判定する。フォーム設定が `VERIFIED` でも、収集を有効化する前に同期されたチケットは `respondent_email` が `NULL` のままであり、フォーム単位の設定だけでは判定できないため。

### データモデル

```sql
CREATE TYPE notification_type AS ENUM ('status_change', 'assignee_assigned');
CREATE TYPE notification_mode AS ENUM ('always', 'confirm', 'off');

CREATE TABLE form_notification_settings (
  form_id UUID NOT NULL,
  notification_type notification_type NOT NULL,
  mode notification_mode NOT NULL DEFAULT 'off',
  include_detail BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (form_id, notification_type),

  CONSTRAINT form_notification_settings_form_fk
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
);

CREATE TABLE ticket_notifications (
  id UUID PRIMARY KEY,
  ticket_id UUID NOT NULL,
  notification_type notification_type NOT NULL,
  sent_by UUID,
  sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT ticket_notifications_ticket_fk
    FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
  CONSTRAINT ticket_notifications_user_fk
    FOREIGN KEY (sent_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX ticket_notifications_ticket_type_sent_at
  ON ticket_notifications(ticket_id, notification_type, sent_at DESC);
```

- `form_notification_settings` に行が存在しない種別は `mode = off`, `include_detail = false` として扱う。フォーム登録時に行を作らないため、**既存フォームで意図せず通知が始まることがない**
- 通知種別を ENUM の行として持つことで、将来チャット機能の通知種別を追加する際にテーブル定義の変更なしで拡張できる（ENUM への値追加のみ）
- `include_detail` のデフォルトは `false`（詳細を含めない）。回答者へ渡す情報量が少ない側を初期値とする
- `ticket_notifications` には**送信に成功したものだけ**を記録する。失敗を記録するとレートリミットに数えられ、届いていないのに再送できない状態になるため。失敗はサーバーログにのみ残す
- `sent_by` は自動送信の場合、その変更を行った操作者を指す
- インデックスはレートリミット判定（直近の送信を1件引く）のために張る

### 送信条件

自動送信・手動送信に共通する条件は以下。

1. `tickets.respondent_email` が `NULL` でないこと
2. 該当種別の `mode` が `off` でないこと

自動送信（`always`）はこれに加えて、

- 対象フィールドが**実際に変更された**こと。`usecase.changeRecorder` が変更を記録した場合と同じ条件で判定する（同じステータスへの更新など、値が変わらない場合は送信しない）
- `assignee_assigned` は変更後の担当者が `NULL` でないこと（解除は対象外）

手動送信はこれに加えて、

- レートリミットに抵触しないこと（後述）
- `assignee_assigned` は現在の担当者が `NULL` でないこと

### 送信内容は常に現在の状態

手動送信・再送では、**送信時点のチケットの状態**を通知する。「どの変更に対応する通知か」を履歴に紐付けることはしない。

回答者にとって意味があるのは現在どうなっているかであり、過去の特定の遷移を後から通知しても混乱を招くため。この仕様により、変更から送信までの間にチケットが再度更新されても不整合が起きない。

### レートリミット

手動送信（`POST /v1/tickets/:ticket_id/notifications`）にのみ適用する。

**同一チケット・同一通知種別につき5分に1通**。`ticket_notifications` から直近の送信を1件引き、5分以内であれば `NOTIFICATION_RATE_LIMITED`（HTTP 429）を返す。

判定には自動送信の記録も含める。直前に `always` の自動通知が飛んでいれば、手動送信も5分間は制限される。保護対象は回答者の受信箱であり、送信の由来は問わないため。

一方、`always` の自動送信自体はレートリミットの対象外とする。管理者が設定したポリシーであり、かつ必ず実際の変更に随伴するため。ただしこれは「編集者がステータスを往復させると都度メールが飛ぶ」ことを意味する。運用上の問題が出た場合は自動送信側にも制限を入れることを再検討する。

### シーケンス

`mode = confirm` の場合。

```text
[フロントエンド]                          [バックエンド]
GET /v1/forms/:form_id/notification-settings
  （画面表示時に取得しキャッシュ）
        │
   ユーザーがステータスを変更
        │
   PATCH /v1/tickets/:id {status_id} ──→ ステータス更新（トランザクション）
        │                                mode = always なら自動送信
        │←──────────────────── 200 OK {..., notification_results: []}
        │
   mode = confirm かつ
   ticket.respondent_email != null
        │
   確認ダイアログ表示
   「この変更を回答者に通知しますか？」
        │
        ├─ 通知する ──→ POST /v1/tickets/:id/notifications
        │                 {notification_type: "status_change"}
        │                        │
        │                   レートリミット判定
        │                   メール送信
        │                   ticket_notifications へ記録
        │                        │
        │←──────────────────── 200 OK {sent_at}
        │
        └─ 通知しない ──→ （リクエストを送らない）
```

`mode = always` の場合は確認ダイアログを挟まず、`PATCH` のレスポンスに含まれる `notification_results` の結果だけを見る。

なお `PATCH` のレスポンスには、詳細取得と同じ `notifications`（種別ごとの最終送信日時）と、この更新の送信結果を表す `notification_results` の両方が含まれる。前者は状態、後者は結果であり別物のため、フィールドを分けている。

再送は、チケット詳細画面の「回答者に通知」ボタンから同じ `POST /v1/tickets/:id/notifications` を呼ぶ。

### 送信タイミングと失敗時の扱い

`always` の自動送信は、トランザクションの**外側**、コミット完了後に行う。ADR-004 で定めた「トランザクション内に外部 I/O を含めない」原則に従う。

送信に失敗してもチケットの変更はロールバックしない。招待メール（ADR-004）と異なり、チケットの変更自体は操作者の意図した結果であり、通知の失敗を理由に取り消すべきではないため。

失敗は `PATCH` のレスポンスの `notifications` に含めて返し、フロントエンドが `sonner` のトーストで「ステータスを変更しましたが、回答者への通知メールの送信に失敗しました」と警告する。あわせてサーバーログにも記録する。

この設計上、`mode = always` のフォームでは `PATCH /v1/tickets/:ticket_id` のレスポンスが Resend API のレイテンシ分だけ遅くなる。カンバンのドラッグ&ドロップによるステータス変更も同様に待たされる。送信結果を操作者に伝えることを優先し、このレイテンシを許容する。

手動送信の失敗は `POST /v1/tickets/:ticket_id/notifications` のエラーレスポンスとして返す。送信に失敗した場合は `ticket_notifications` に記録しないため、すぐに再試行できる。

### メールテンプレート

種別2 × 粒度2 の4パターンを用意する。`renderTemplate` が条件分岐を持たないため、テンプレートディレクトリを分ける。

| テンプレート名 | 用途 | テンプレート変数 |
| --- | --- | --- |
| `ticket-status-changed` | ステータス変更・詳細なし | `form_title` |
| `ticket-status-changed-detailed` | ステータス変更・詳細あり | `form_title`, `status_name` |
| `ticket-assigned` | 担当者アサイン・詳細なし | `form_title` |
| `ticket-assigned-detailed` | 担当者アサイン・詳細あり | `form_title`, `assignee_name` |

回答者向けの確認リンクは含めないため、`accept_url` に相当する変数は持たない。

### API

新規に3つのエンドポイントを追加し、既存の `PATCH`/`GET /v1/tickets/:ticket_id` のレスポンスを拡張する。詳細は `docs/api/notification.md` および `docs/api/ticket.md` を参照。

| メソッド | パス | 権限 | 用途 |
| --- | --- | --- | --- |
| `GET` | `/v1/forms/:form_id/notification-settings` | Editor 以上 | 設定の取得 |
| `PATCH` | `/v1/forms/:form_id/notification-settings` | Admin のみ | 設定の変更 |
| `POST` | `/v1/tickets/:ticket_id/notifications` | Editor 以上 | 手動送信・再送 |

設定の `GET` を Editor 以上にしているのは、フロントエンドが確認ダイアログを出すか判断するために設定を知る必要があるため。変更は管理者に限る。

本プロジェクトは `openapi/openapi.yaml` から `oapi-codegen` で `internal/api/gen.go` を生成しているため、いずれも OpenAPI 仕様を先に更新して再生成する。

### 追加するエラーコード

| コード | HTTP | 意味 |
| --- | --- | --- |
| `NOTIFICATION_RATE_LIMITED` | 429 | 手動送信のレートリミットに抵触した |
| `NOTIFICATION_DISABLED` | 409 | 該当種別の通知設定が `off` である |
| `RESPONDENT_EMAIL_MISSING` | 409 | チケットに回答者のメールアドレスがない |

`NOTIFICATION_RATE_LIMITED` は本プロジェクト初の HTTP 429 となる。`handler/error.go` の `errorDefs` に追加する。

### 設定画面

フォーム管理画面に、既存の `statuses-dialog.tsx` / `members-dialog.tsx` と同じパターンのダイアログを追加する。Admin 以外には変更操作を見せない。

`email_collection_type` が `DO_NOT_COLLECT` の場合、前述の警告を表示する。

## 検討した別の選択肢

### PATCH の `notify` フィールドで送信する

`PATCH /v1/tickets/:ticket_id` のリクエストボディに `notify` を持たせ、確認ダイアログを変更前に表示して1回のリクエストで変更と通知を完了させる案。

- メリット: リクエストが1往復で済む。送信が必ず実際の変更に随伴するため、悪用経路が構造的に存在せずレートリミットが不要
- デメリット: 変更に随伴する送信しかできないため、**通知の再送ができない**。「メールが届かなかったのでもう一度送りたい」という運用要求を満たせない

**不採用理由:** 再送を要件に含めた結果、どのみち単独で送信できる経路とレートリミットが必要になり、`notify` を残すと送信経路が2本になって履歴記録とレートリミットを二重に実装することになるため。

### 承認待ちキューを設ける

`confirm` を「保留中の通知」一覧として永続化し、管理者が後からまとめて承認・却下する案。

- メリット: 操作した編集者のその場の判断に依存せず、レビューを挟める
- デメリット: 新しいテーブル・一覧 UI・承認 API が必要で実装コストが大きい

**不採用理由:** 現時点の要件に対して実装コストが見合わないため。

### 送信を変更履歴（`ticket_histories`）に紐付ける

どの変更に対する通知かを履歴 ID で特定し、同じ履歴に対しては一度しか送れないようにする案。

- メリット: 同一の変更について重複して通知することがない
- デメリット: 再送ができなくなる。また通知内容が過去の状態に固定されるため、その後チケットが更新されていると実態と食い違う通知が飛ぶ

**不採用理由:** 再送を要件に含めたため。重複送信の抑制はレートリミットで代替する。

### ステータスごとに通知有無を設定する

`form_statuses` に「回答者へ通知するか」のフラグを持たせ、「対応完了」だけ通知するといった制御を可能にする案。

- メリット: 回答者にとって意味のある遷移だけを通知でき、メールの量を抑えられる
- デメリット: 設定項目と UI が複雑になる。ステータスの追加・削除のたびに通知設定の考慮が必要になる

**不採用理由:** 初期スコープを絞るため。必要になった時点で `form_statuses` への列追加で拡張できる。

### メール送信を非同期にする

既存の `PublishTicketUpdated` と同様に fire-and-forget で送信する案。

- メリット: `PATCH` のレスポンスが Resend API のレイテンシに影響されない
- デメリット: 送信結果をレスポンスに含められないため、送信失敗を操作者に伝えられない

**不採用理由:** 送信失敗を操作者へ警告表示することを優先したため。

### `email_collection_type` で送信可否を判定する

フォーム単位の設定値を送信条件に含める案。

- メリット: チケットを見なくてもフォーム単位で通知の可否が分かる
- デメリット: 収集を有効化する前に同期されたチケットは `respondent_email` が `NULL` のままであり、フォーム設定が `VERIFIED` でも送信できない。逆にフォーム設定を後から `DO_NOT_COLLECT` に変えても、既存チケットのアドレスは有効なまま

**不採用理由:** フォーム単位の設定とチケット単位の実態が一致しないため。判定は `respondent_email` の有無で行い、`email_collection_type` は設定画面の警告表示にのみ使う。

## 注意点・懸念事項

- **回答者は配信停止できない**。回答者はサービスのユーザーではないため、通知を受け取りたくない場合の手段がない。通知が意図せず多量に送られないよう、デフォルトを `off` とし、`always` を選ぶかは管理者の判断に委ねる
- **`always` の自動送信はレートリミットの対象外**。編集者がステータスを往復させると都度メールが飛ぶ。実際の変更に随伴する送信であり管理者のポリシーでもあるため許容するが、運用次第では制限の追加を検討する
- **`mode = always` では `PATCH /v1/tickets/:ticket_id` のレイテンシが増加する**。同期送信を選択したトレードオフ。特にカンバンのドラッグ&ドロップは操作が連続しやすく、体感への影響が出る可能性がある
- **`confirm` モードではカンバンのドラッグごとに確認ダイアログが出る**。`response-kanban-view.tsx` の `handleDragEnd` はドロップ即座に変更を実行するため、ダイアログが操作のテンポを損なう。ドラッグ操作を多用するフォームでは `always` または `off` の選択が現実的である旨をヘルプ等で示す
- **ステータス名・担当者名がメール本文に載る**（`include_detail = true` の場合）。内部的な名称をそのまま外部の回答者へ送ることになるため、初期値を `false` とし、管理者が意識的に有効化する設計とする
- **フロントエンドは現在チケット更新のエラーを握りつぶしている**。`use-form-responses.ts` の `updateResponseStatus` などは `console.error` のみでユーザーに何も表示していない。通知失敗の警告表示にあわせて、この箇所のエラーハンドリングも見直す
- **`email_collection_type` は同期するまで埋まらない**。既存フォームは次回の同期まで `NULL` のままで、この間は警告を表示できない。`NULL` を「不明」として扱い警告しない仕様のため、収集していないフォームで一時的に警告が出ないことになる
