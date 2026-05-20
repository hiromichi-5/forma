package usecase

import (
	"context"

	"github.com/google/uuid"
)

type TicketEvent struct {
	FormID   uuid.UUID
	TicketID uuid.UUID
}

type EventPublisher interface {
	PublishTicketUpdated(ctx context.Context, event TicketEvent) error
	Subscribe(formID uuid.UUID) (<-chan TicketEvent, func())
}
