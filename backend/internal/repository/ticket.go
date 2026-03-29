package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type TicketRepository interface {
	// Create はチケットを作成する。ON CONFLICT DO NOTHING により、
	// 既存の (form_id, response_id) と重複した場合は false を返す。
	Create(ctx context.Context, ticket entity.Ticket) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error)
	List(ctx context.Context, formID uuid.UUID, statusID *uuid.UUID) ([]entity.Ticket, error)

	UpdateStatus(ctx context.Context, id uuid.UUID, statusID uuid.UUID) error
	UpdateAssignee(ctx context.Context, id uuid.UUID, assigneeID *uuid.UUID) error
	UpdatePriority(ctx context.Context, id uuid.UUID, priority string) error

	CreateHistory(ctx context.Context, history entity.TicketHistory) (entity.TicketHistory, error)
	ListHistories(ctx context.Context, ticketID uuid.UUID) ([]entity.TicketHistory, error)

	CountByStatus(ctx context.Context, statusID uuid.UUID) (int64, error)
}
