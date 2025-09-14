package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hiromichi-5/forma/backend/internal/db"
)

func (s *Service) ListTickets(ctx context.Context, formID, status string, actor uuid.UUID) ([]db.Ticket, error) {
	if formID != "" {
		if err := s.RequireEditor(ctx, formID, actor); err != nil {
			return nil, err
		}
		return s.Q.ListTickets(ctx, db.ListTicketsParams{
			Column1: formID,
			Column2: status,
		})
	}

	return []db.Ticket{}, nil
}

func (s *Service) GetTicket(ctx context.Context, id string, actor uuid.UUID) (db.Ticket, error) {
	if err := s.RequireFormAccessForTicket(ctx, id, actor); err != nil {
		return db.Ticket{}, err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return db.Ticket{}, ErrValidation
	}
	return s.Q.GetTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
}

func (s *Service) UpdateTicket(ctx context.Context, id string, status *string, assignee *uuid.UUID, actor uuid.UUID) (db.Ticket, error) {
	if err := s.RequireFormAccessForTicket(ctx, id, actor); err != nil {
		return db.Ticket{}, err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return db.Ticket{}, ErrValidation
	}
	var st string
	if status != nil {
		st = *status
	}
	var aid pgtype.UUID
	if assignee != nil {
		aid = pgtype.UUID{Bytes: *assignee, Valid: true}
	}
	return s.Q.UpdateTicket(ctx, db.UpdateTicketParams{
		ID:         pgtype.UUID{Bytes: uid, Valid: true},
		Status:     st,
		AssigneeID: aid,
	})
}
