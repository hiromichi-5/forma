# SSE によるチケットリアルタイム配信

## 目的

チケットの変更（ステータス・担当者・優先度）を、同じフォームを閲覧している他のユーザーへリアルタイムに反映する。
Server-Sent Events（SSE）を使ってサーバーからクライアントへ変更イベントをプッシュし、手動リロードを不要にする。

### Non Goals

- チケット作成イベントの配信
- 複数サーバーインスタンスへの対応


## 概要

PATCH `/v1/tickets/:ticket_id` でチケットが更新されると、usecase 層がインメモリの `EventHub` にイベントを発行する。`EventHub` は同じ `form_id` を購読しているすべての SSE 接続にイベントをブロードキャストする。

SSE クライアントはイベントを受け取るとチケット一覧のキャッシュを更新し、UI に即時反映する。

```
[User A] PATCH /v1/tickets/:id
         │
         ▼
    TicketUseCase.UpdateTicket()
         │  uow.Do() 成功後
         ▼
    EventHub.Publish(TicketEvent{form_id, ticket_id})
         │
         ├──▶ chan (User B の SSE 接続)
         └──▶ chan (User C の SSE 接続)
                    │
                    ▼
              SSE ストリーム送信
                    │
                    ▼
              フロントエンドがキャッシュ更新
```

インメモリ実装は Go の `sync.RWMutex` + `map[formID][]chan` で構成する。将来的に複数インスタンスが必要になった場合は `EventPublisher` インターフェースの実装を差し替えることで対応できる。

---

## 設計詳細

### データモデル・型定義

```go
// backend/internal/usecase/event.go

type TicketEvent struct {
    FormID   uuid.UUID
    TicketID uuid.UUID
}

type EventPublisher interface {
    PublishTicketUpdated(ctx context.Context, event TicketEvent) error
    Subscribe(formID uuid.UUID) (ch <-chan TicketEvent, unsubscribe func())
}
```

```go
// backend/internal/infra/pubsub/memory.go

type MemoryHub struct {
    mu   sync.RWMutex
    subs map[uuid.UUID][]chan TicketEvent
}
```

### SSE エンドポイント

```
GET /v1/forms/:form_id/stream
```

| 項目 | 内容 |
|------|------|
| レスポンス | `Content-Type: text/event-stream` |

### イベント種別

| event 名 | トリガー | data の内容 |
|---------|---------|------------|
| `ticket_updated` | PATCH /v1/tickets/:id | `TicketDetailResp`（既存レスポンスと同形） |
| `ping` | 30 秒ごと | `{}` |

`ticket_updated` のペイロード:

```jsonc
{
  "ticket_id": "uuid",
  "form_id": "uuid",
  "ticket": {
    // TicketDetailResp と同形（既存の GET /v1/tickets/:id のレスポンスと同形）
  }
}
```

TicketEvent は ID のみ持ち、SSE ハンドラが `ticketUC.GetTicket()` でフル DTO を取得して送信する。
これにより `TicketDetail`（usecase 型）を infra 層のイベント構造体に乗せるレイヤー逆転を避ける。

### シーケンス

#### チケット更新 → SSE 配信

```
User A (browser)          Backend                           User B (browser)
     │                       │                                    │
     │  PATCH /tickets/:id   │                                    │
     │──────────────────────▶│                                    │
     │                       │  uow.Do()                          │
     │                       │  ├─ UpdateStatus/Assignee/Priority │
     │                       │  └─ CreateHistory                  │
     │                       │  (commit)                          │
     │                       │                                    │
     │                       │  hub.Publish(TicketEvent)          │
     │                       │──────────────────────────────────▶ │
     │                       │                    SSE: ticket_updated
     │  200 OK               │                                    │
     │◀──────────────────────│                                    │
```

#### SSE 接続確立

```
User B (browser)          Backend
     │                        │
     │  GET /v1/forms/:id/stream
     │──────────────────────▶ │
     │                        │  SessionMiddleware（認証）
     │                        │  requireEditor（認可）
     │                        │  hub.Subscribe(formID)
     │                        │
     │  200 text/event-stream │
     │ ◀──────────────────────│
     │                        │
     │   (接続維持)            │
     │ ◀──── ping (30s) ──────│
     │ ◀──── ping (30s) ──────│
```

### MemoryHub の実装方針

```go
func (h *MemoryHub) Subscribe(formID uuid.UUID) (<-chan TicketEvent, func()) {
    ch := make(chan TicketEvent, 8) // バッファあり
    h.mu.Lock()
    h.subs[formID] = append(h.subs[formID], ch)
    h.mu.Unlock()

    unsubscribe := func() {
        h.mu.Lock()
        defer h.mu.Unlock()
        // スライスから ch を削除して close
    }
    return ch, unsubscribe
}

func (h *MemoryHub) PublishTicketUpdated(_ context.Context, event TicketEvent) error {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for _, ch := range h.subs[event.FormID] {
        select {
        case ch <- event:
        default:
            // 遅いクライアントのチャネルが満杯なら drop（ブロックしない）
        }
    }
    return nil
}
```

バッファサイズを8として、一時的にイベントが溜まることを許容する。チャネルが埋まっている場合は遅いクライアントへのイベントを drop する。クリティカルなデータではなく UIのリアルタイム更新のため許容する。


### フロントエンド

```typescript
// frontend/src/hooks/use-ticket-stream.ts

export function useTicketStream(formId: string) {
    const queryClient = useQueryClient()

    useEffect(() => {
        const es = new EventSource(
            `/v1/forms/${formId}/stream`,
            { withCredentials: true }
        )
        es.addEventListener('ticket_updated', (e: MessageEvent) => {
            const { ticket } = JSON.parse(e.data)
            queryClient.setQueryData(['tickets', formId], (prev) =>
                prev?.map((t) => t.id === ticket.id ? ticket : t)
            )
        })
        return () => es.close()
    }, [formId, queryClient])
}
```

`use-form-responses.ts` でこのフックを呼び出すことで、SSE 受信時に既存のキャッシュをパッチ更新する。

---

## 検討した別の選択肢

- **PostgreSQL LISTEN/NOTIFY**: Cloud Run の複数インスタンスでも動作するが、現状は `--max-instances=1` で運用するため不要で、PostgreSQLに依存することも避けられるため、採用しない。

- **TicketEvent に完全なデータを乗せる**: SSE ハンドラ内の DB 往復をなくせるが、usecase 層の型（`TicketDetail`）を infra 層のイベント構造体に持たせることになりレイヤー依存が逆転するため採用しない。また、現状はDBへのアクセスを低減する必要はないと判断した。

- **WebSocket**: 双方向通信が不要なため採用しない。
