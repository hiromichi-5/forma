package pubsub

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
)

type MemoryHub struct {
	mu   sync.RWMutex
	subs map[uuid.UUID][]chan usecase.TicketEvent
}

func NewMemoryHub() *MemoryHub {
	return &MemoryHub{
		subs: make(map[uuid.UUID][]chan usecase.TicketEvent),
	}
}

func (h *MemoryHub) Subscribe(formID uuid.UUID) (<-chan usecase.TicketEvent, func()) {
	ch := make(chan usecase.TicketEvent, 8)
	h.mu.Lock()
	h.subs[formID] = append(h.subs[formID], ch)
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.subs[formID]
		for i, c := range subs {
			if c == ch {
				h.subs[formID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}
	return ch, unsubscribe
}

func (h *MemoryHub) PublishTicketUpdated(_ context.Context, event usecase.TicketEvent) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.subs[event.FormID] {
		select {
		case ch <- event:
		default:
		}
	}
	return nil
}
