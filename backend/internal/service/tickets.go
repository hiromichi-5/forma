package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hiromichi-5/forma/backend/internal/db"
)

func (s *Service) ListTickets(ctx context.Context, formID, status string, actor uuid.UUID) ([]TicketSummary, error) {
	if formID == "" {
		return []TicketSummary{}, nil
	}
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return nil, err
	}

	rows, err := s.Q.ListTickets(ctx, db.ListTicketsParams{
		Column1: formID,
		Column2: status,
	})
	if err != nil {
		return nil, err
	}

	cache := make(map[string]*formQuestionSet)
	summaries := make([]TicketSummary, 0, len(rows))
	for _, row := range rows {
		set, err := s.getFormQuestionSet(ctx, row.FormID, cache)
		if err != nil {
			return nil, err
		}
		answers, err := parseResponseAnswers(row.Payload)
		if err != nil {
			return nil, err
		}
		summary := buildTicketSummary(row, answers, *set)
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (s *Service) GetTicket(ctx context.Context, id string, actor uuid.UUID) (TicketDetail, error) {
	if err := s.RequireFormAccessForTicket(ctx, id, actor); err != nil {
		return TicketDetail{}, err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return TicketDetail{}, ErrValidation
	}

	row, err := s.Q.GetTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		return TicketDetail{}, err
	}
	set, err := s.getFormQuestionSet(ctx, row.FormID, nil)
	if err != nil {
		return TicketDetail{}, err
	}
	answers, err := parseResponseAnswers(row.Payload)
	if err != nil {
		return TicketDetail{}, err
	}
	return buildTicketDetail(row, answers, *set), nil
}

func (s *Service) UpdateTicket(ctx context.Context, id string, status *string, assignee *uuid.UUID, priority *int32, actor uuid.UUID) (TicketDetail, error) {
	if err := s.RequireFormAccessForTicket(ctx, id, actor); err != nil {
		return TicketDetail{}, err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return TicketDetail{}, ErrValidation
	}

	currentRow, err := s.Q.GetTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TicketDetail{}, err
		}
		return TicketDetail{}, err
	}

	if assignee != nil {
		if err := s.validateAssignee(ctx, currentRow.FormID, *assignee); err != nil {
			return TicketDetail{}, err
		}
	}

	params := db.UpdateTicketParams{
		ID:         pgtype.UUID{Bytes: uid, Valid: true},
		Status:     nil,
		AssigneeID: pgtype.UUID{},
		Priority:   pgtype.Int4{},
	}

	if status != nil {
		st := strings.TrimSpace(*status)
		if st == "" {
			return TicketDetail{}, ErrValidation
		}
		params.Status = st
	}
	if assignee != nil {
		params.AssigneeID = pgtype.UUID{Bytes: *assignee, Valid: true}
	}
	if priority != nil {
		if *priority < 1 || *priority > 5 {
			return TicketDetail{}, ErrValidation
		}
		params.Priority = pgtype.Int4{Int32: *priority, Valid: true}
	}

	if _, err := s.Q.UpdateTicket(ctx, params); err != nil {
		return TicketDetail{}, err
	}

	updatedRow, err := s.Q.GetTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		return TicketDetail{}, err
	}
	set, err := s.getFormQuestionSet(ctx, updatedRow.FormID, nil)
	if err != nil {
		return TicketDetail{}, err
	}
	answers, err := parseResponseAnswers(updatedRow.Payload)
	if err != nil {
		return TicketDetail{}, err
	}
	return buildTicketDetail(updatedRow, answers, *set), nil
}

func (s *Service) getFormQuestionSet(ctx context.Context, formID string, cache map[string]*formQuestionSet) (*formQuestionSet, error) {
	if cache != nil {
		if set, ok := cache[formID]; ok {
			return set, nil
		}
	}
	rows, err := s.Q.ListFormQuestions(ctx, formID)
	if err != nil {
		return nil, err
	}
	set := newFormQuestionSet(rows)
	if cache != nil {
		cache[formID] = &set
	}
	return &set, nil
}

func (s *Service) validateAssignee(ctx context.Context, formID string, assignee uuid.UUID) error {
	_, err := s.Q.GetUserFormRole(ctx, db.GetUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: assignee, Valid: true},
		FormID: formID,
	})
	if err != nil {
		return ErrValidation
	}
	return nil
}
