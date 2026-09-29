package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type TicketFilter struct {
	FormID     uuid.UUID
	StatusIDs  []uuid.UUID
	EmailQuery *string
}

type TicketCursor struct {
	SubmittedAt time.Time
	ID          uuid.UUID
}

type TicketStats struct {
	Count             int64
	LatestSubmittedAt time.Time
}

type TicketRepository interface {
	// Create はチケットを作成する。ON CONFLICT DO NOTHING により、
	// 既存の (form_id, response_id) と重複した場合は false を返す。
	Create(ctx context.Context, ticket entity.Ticket) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error)
	List(
		ctx context.Context,
		filter TicketFilter,
		after *TicketCursor,
		limit int,
	) ([]entity.Ticket, error)
	CountGroupByStatus(
		ctx context.Context,
		formID uuid.UUID,
		emailQuery *string,
	) (map[uuid.UUID]int64, error)
	SummarizeByForms(ctx context.Context, formIDs []uuid.UUID) (map[uuid.UUID]TicketStats, error)

	Save(ctx context.Context, ticket entity.Ticket) error

	CreateHistory(ctx context.Context, history entity.TicketHistory) (entity.TicketHistory, error)
	ListHistories(ctx context.Context, ticketID uuid.UUID) ([]entity.TicketHistory, error)

	CountByStatus(ctx context.Context, statusID uuid.UUID) (int64, error)
}
