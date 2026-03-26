package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TicketHistoryView struct {
	ID            string    `json:"id"`
	TicketID      string    `json:"ticket_id"`
	ChangedBy     *string   `json:"changed_by"`
	ChangedByName string    `json:"changed_by_name"`
	FieldName     string    `json:"field_name"`
	OldValue      *string   `json:"old_value"`
	NewValue      *string   `json:"new_value"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *Service) ListTicketHistories(
	ctx context.Context,
	ticketID string,
	actor uuid.UUID,
) ([]TicketHistoryView, error) {
	if err := s.RequireFormAccessForTicket(ctx, ticketID, actor); err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, ErrValidation
	}

	rows, err := s.Q.ListTicketHistoriesByTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []TicketHistoryView{}, nil
		}
		return nil, err
	}

	out := make([]TicketHistoryView, 0, len(rows))
	for _, row := range rows {
		out = append(out, ticketHistoryFromRow(row))
	}
	return out, nil
}

func ticketHistoryFromRow(row db.TicketHistory) TicketHistoryView {
	return TicketHistoryView{
		ID:            uuid.UUID(row.ID.Bytes).String(),
		TicketID:      uuid.UUID(row.TicketID.Bytes).String(),
		ChangedBy:     uuidPtr(row.ChangedBy),
		ChangedByName: row.ChangedByName,
		FieldName:     row.FieldName,
		OldValue:      textPtr(row.OldValue),
		NewValue:      textPtr(row.NewValue),
		CreatedAt:     row.CreatedAt.Time,
	}
}

func uuidPtr(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	value := uuid.UUID(id.Bytes).String()
	return &value
}
