package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var allowedTicketPriorities = map[string]struct{}{
	"high":   {},
	"medium": {},
	"low":    {},
}

func isValidTicketPriority(p string) bool {
	_, ok := allowedTicketPriorities[p]
	return ok
}

func (s *Service) ListTickets(
	ctx context.Context,
	formID, statusID string,
	actor uuid.UUID,
) ([]TicketSummary, error) {
	if formID == "" {
		return []TicketSummary{}, nil
	}
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return nil, err
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return nil, ErrValidation
	}
	var sid pgtype.UUID
	if statusID != "" {
		parsed, err := uuid.Parse(statusID)
		if err != nil {
			return nil, ErrValidation
		}
		sid = pgtype.UUID{Bytes: parsed, Valid: true}
	}

	rows, err := s.Q.ListTickets(ctx, db.ListTicketsParams{
		Column1: pgtype.UUID{Bytes: fid, Valid: true},
		Column2: sid,
	})
	if err != nil {
		return nil, err
	}

	cache := make(map[string]*formQuestionSet)
	summaries := make([]TicketSummary, 0, len(rows))
	for _, row := range rows {
		set, err := s.getFormQuestionSet(ctx, uuid.UUID(row.FormID.Bytes).String(), cache)
		if err != nil {
			return nil, err
		}
		answers, err := parseResponseAnswers(row.Answers)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return TicketDetail{}, ErrForbidden
		}
		return TicketDetail{}, err
	}
	set, err := s.getFormQuestionSet(ctx, uuid.UUID(row.FormID.Bytes).String(), nil)
	if err != nil {
		return TicketDetail{}, err
	}
	answers, err := parseResponseAnswers(row.Answers)
	if err != nil {
		return TicketDetail{}, err
	}
	return buildTicketDetail(row, answers, *set), nil
}

func (s *Service) UpdateTicket(
	ctx context.Context,
	id string,
	statusID *string,
	assignee *uuid.UUID,
	clearAssignee bool,
	priority *string,
	actor uuid.UUID,
) (TicketDetail, error) {
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
			return TicketDetail{}, ErrForbidden
		}
		return TicketDetail{}, err
	}

	changedByName, err := s.getUserDisplayName(ctx, actor)
	if err != nil {
		return TicketDetail{}, err
	}

	if clearAssignee && assignee != nil {
		return TicketDetail{}, ErrValidation
	}

	if assignee != nil {
		if err := s.validateAssignee(ctx, currentRow.FormID, *assignee); err != nil {
			return TicketDetail{}, err
		}
	}

	var statusChanged bool
	var oldStatusName string
	var newStatusName string
	if statusID != nil {
		st := strings.TrimSpace(*statusID)
		if st == "" {
			return TicketDetail{}, ErrValidation
		}
		statusUUID, err := uuid.Parse(st)
		if err != nil {
			return TicketDetail{}, ErrValidation
		}
		statusRow, err := s.Q.GetFormStatusByID(ctx, pgtype.UUID{Bytes: statusUUID, Valid: true})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return TicketDetail{}, ErrValidation
			}
			return TicketDetail{}, err
		}
		if statusRow.FormID != currentRow.FormID {
			return TicketDetail{}, ErrValidation
		}
		if !equalUUID(statusRow.ID, currentRow.StatusID) {
			oldStatusRow, err := s.Q.GetFormStatusByID(ctx, currentRow.StatusID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return TicketDetail{}, ErrValidation
				}
				return TicketDetail{}, err
			}
			oldStatusName = oldStatusRow.Name
			newStatusName = statusRow.Name
			statusChanged = true

			if _, err := s.Q.UpdateTicketStatus(ctx, db.UpdateTicketStatusParams{
				ID:       pgtype.UUID{Bytes: uid, Valid: true},
				StatusID: statusRow.ID,
			}); err != nil {
				return TicketDetail{}, err
			}
		}
	}

	var assigneeChanged bool
	var oldAssigneeName *string
	var newAssigneeName *string
	if clearAssignee || assignee != nil {
		param := pgtype.UUID{Valid: false}
		if assignee != nil {
			param = pgtype.UUID{Bytes: *assignee, Valid: true}
		}
		if !equalUUID(currentRow.AssigneeID, param) {
			assigneeChanged = true
			if currentRow.AssigneeID.Valid {
				name, err := s.getUserDisplayName(ctx, uuid.UUID(currentRow.AssigneeID.Bytes))
				if err != nil {
					return TicketDetail{}, err
				}
				oldAssigneeName = &name
			}
			if param.Valid {
				name, err := s.getUserDisplayName(ctx, uuid.UUID(param.Bytes))
				if err != nil {
					return TicketDetail{}, err
				}
				newAssigneeName = &name
			}
			if _, err := s.Q.UpdateTicketAssignee(ctx, db.UpdateTicketAssigneeParams{
				ID:         pgtype.UUID{Bytes: uid, Valid: true},
				AssigneeID: param,
			}); err != nil {
				return TicketDetail{}, err
			}
		}
	}

	var priorityChanged bool
	var oldPriority string
	var newPriority string
	if priority != nil {
		p := strings.TrimSpace(*priority)
		p = strings.ToLower(p)
		if !isValidTicketPriority(p) {
			return TicketDetail{}, ErrValidation
		}
		if p != currentRow.Priority {
			oldPriority = currentRow.Priority
			newPriority = p
			priorityChanged = true
			if _, err := s.Q.UpdateTicketPriority(ctx, db.UpdateTicketPriorityParams{
				ID:       pgtype.UUID{Bytes: uid, Valid: true},
				Priority: p,
			}); err != nil {
				return TicketDetail{}, err
			}
		}
	}

	histories := buildTicketHistoryChanges(
		statusChanged,
		oldStatusName,
		newStatusName,
		assigneeChanged,
		oldAssigneeName,
		newAssigneeName,
		priorityChanged,
		oldPriority,
		newPriority,
	)
	for _, history := range histories {
		if _, err := s.Q.CreateTicketHistory(ctx, db.CreateTicketHistoryParams{
			ID:            dbUUID(uuid.New()),
			TicketID:      dbUUID(uid),
			ChangedBy:     dbUUID(actor),
			ChangedByName: changedByName,
			FieldName:     history.FieldName,
			OldValue:      textFromStringPtr(history.OldValue),
			NewValue:      textFromStringPtr(history.NewValue),
		}); err != nil {
			return TicketDetail{}, err
		}
	}

	updatedRow, err := s.Q.GetTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		return TicketDetail{}, err
	}
	set, err := s.getFormQuestionSet(ctx, uuid.UUID(updatedRow.FormID.Bytes).String(), nil)
	if err != nil {
		return TicketDetail{}, err
	}
	answers, err := parseResponseAnswers(updatedRow.Answers)
	if err != nil {
		return TicketDetail{}, err
	}
	return buildTicketDetail(updatedRow, answers, *set), nil
}

func (s *Service) getFormQuestionSet(
	ctx context.Context,
	formID string,
	cache map[string]*formQuestionSet,
) (*formQuestionSet, error) {
	if cache != nil {
		if set, ok := cache[formID]; ok {
			return set, nil
		}
	}
	uid, err := uuid.Parse(formID)
	if err != nil {
		return nil, ErrValidation
	}
	rows, err := s.Q.ListFormQuestions(ctx, dbUUID(uid))
	if err != nil {
		return nil, err
	}
	set := newFormQuestionSet(rows)
	if cache != nil {
		cache[formID] = &set
	}
	return &set, nil
}

func (s *Service) validateAssignee(
	ctx context.Context,
	formID pgtype.UUID,
	assignee uuid.UUID,
) error {
	_, err := s.Roles.GetFormMemberRole(ctx, db.GetFormMemberRoleParams{
		UserID: dbUUID(assignee),
		FormID: formID,
	})
	if err != nil {
		return ErrValidation
	}
	return nil
}
