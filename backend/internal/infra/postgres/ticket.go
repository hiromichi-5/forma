package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.TicketRepository = (*TicketRepository)(nil)

type TicketRepository struct {
	q *db.Queries
}

func NewTicketRepository(pool *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{q: db.New(pool)}
}

func (r *TicketRepository) Create(ctx context.Context, ticket entity.Ticket) (bool, error) {
	n, err := r.q.CreateTicket(ctx, db.CreateTicketParams{
		ID:              toUUID(ticket.ID),
		FormID:          toUUID(ticket.FormID),
		ResponseID:      ticket.ResponseID,
		RespondentEmail: toTextPtr(ticket.RespondentEmail),
		Answers:         ticket.Answers,
		StatusID:        toUUID(ticket.StatusID),
		AssigneeID:      toNullUUID(ticket.AssigneeID),
		Priority:        ticket.Priority,
		SubmittedAt:     toTimestamptz(ticket.SubmittedAt),
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error) {
	row, err := r.q.GetTicket(ctx, toUUID(id))
	if err != nil {
		return entity.Ticket{}, repoError(err)
	}
	return toTicket(row), nil
}

func (r *TicketRepository) List(
	ctx context.Context,
	formID uuid.UUID,
	statusID *uuid.UUID,
) ([]entity.Ticket, error) {
	rows, err := r.q.ListTickets(ctx, db.ListTicketsParams{
		FormID:  toUUID(formID),
		Column2: toNullUUID(statusID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]entity.Ticket, len(rows))
	for i, row := range rows {
		result[i] = toTicket(row)
	}
	return result, nil
}

func (r *TicketRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	statusID uuid.UUID,
) error {
	n, err := r.q.UpdateTicketStatus(ctx, db.UpdateTicketStatusParams{
		ID:       toUUID(id),
		StatusID: toUUID(statusID),
	})
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *TicketRepository) UpdateAssignee(
	ctx context.Context,
	id uuid.UUID,
	assigneeID *uuid.UUID,
) error {
	n, err := r.q.UpdateTicketAssignee(ctx, db.UpdateTicketAssigneeParams{
		ID:         toUUID(id),
		AssigneeID: toNullUUID(assigneeID),
	})
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *TicketRepository) UpdatePriority(
	ctx context.Context,
	id uuid.UUID,
	priority string,
) error {
	n, err := r.q.UpdateTicketPriority(ctx, db.UpdateTicketPriorityParams{
		ID:       toUUID(id),
		Priority: priority,
	})
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *TicketRepository) CreateHistory(
	ctx context.Context,
	history entity.TicketHistory,
) (entity.TicketHistory, error) {
	row, err := r.q.CreateTicketHistory(ctx, db.CreateTicketHistoryParams{
		ID:            toUUID(history.ID),
		TicketID:      toUUID(history.TicketID),
		ChangedBy:     toNullUUID(history.ChangedBy),
		ChangedByName: history.ChangedByName,
		FieldName:     history.FieldName,
		OldValue:      toTextPtr(history.OldValue),
		NewValue:      toTextPtr(history.NewValue),
	})
	if err != nil {
		return entity.TicketHistory{}, err
	}
	return toTicketHistory(row), nil
}

func (r *TicketRepository) ListHistories(
	ctx context.Context,
	ticketID uuid.UUID,
) ([]entity.TicketHistory, error) {
	rows, err := r.q.ListTicketHistoriesByTicket(ctx, toUUID(ticketID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.TicketHistory, len(rows))
	for i, row := range rows {
		result[i] = toTicketHistory(row)
	}
	return result, nil
}

func (r *TicketRepository) CountByStatus(ctx context.Context, statusID uuid.UUID) (int64, error) {
	return r.q.CountTicketsByStatus(ctx, toUUID(statusID))
}

func toTicket(row db.Ticket) entity.Ticket {
	return entity.Ticket{
		ID:              fromUUID(row.ID),
		FormID:          fromUUID(row.FormID),
		ResponseID:      row.ResponseID,
		RespondentEmail: fromTextPtr(row.RespondentEmail),
		Answers:         row.Answers,
		StatusID:        fromUUID(row.StatusID),
		AssigneeID:      fromNullUUID(row.AssigneeID),
		Priority:        row.Priority,
		SubmittedAt:     fromTimestamptz(row.SubmittedAt),
		CreatedAt:       fromTimestamptz(row.CreatedAt),
	}
}

func toTicketHistory(row db.TicketHistory) entity.TicketHistory {
	return entity.TicketHistory{
		ID:            fromUUID(row.ID),
		TicketID:      fromUUID(row.TicketID),
		ChangedBy:     fromNullUUID(row.ChangedBy),
		ChangedByName: row.ChangedByName,
		FieldName:     row.FieldName,
		OldValue:      fromTextPtr(row.OldValue),
		NewValue:      fromTextPtr(row.NewValue),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
	}
}
