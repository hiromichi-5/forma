package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ticketHistoryChange struct {
	FieldName string
	OldValue  *string
	NewValue  *string
}

func buildTicketHistoryChanges(statusChanged bool, oldStatusName, newStatusName string, assigneeChanged bool, oldAssigneeName, newAssigneeName *string, priorityChanged bool, oldPriority, newPriority string) []ticketHistoryChange {
	changes := make([]ticketHistoryChange, 0, 3)
	if statusChanged {
		oldValue := oldStatusName
		newValue := newStatusName
		changes = append(changes, ticketHistoryChange{
			FieldName: "status",
			OldValue:  &oldValue,
			NewValue:  &newValue,
		})
	}
	if assigneeChanged {
		changes = append(changes, ticketHistoryChange{
			FieldName: "assignee",
			OldValue:  oldAssigneeName,
			NewValue:  newAssigneeName,
		})
	}
	if priorityChanged {
		oldValue := oldPriority
		newValue := newPriority
		changes = append(changes, ticketHistoryChange{
			FieldName: "priority",
			OldValue:  &oldValue,
			NewValue:  &newValue,
		})
	}
	return changes
}

func equalUUID(a, b pgtype.UUID) bool {
	if a.Valid != b.Valid {
		return false
	}
	if !a.Valid {
		return true
	}
	return a.Bytes == b.Bytes
}

func textFromStringPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func (s *Service) getUserDisplayName(ctx context.Context, userID uuid.UUID) (string, error) {
	row, err := s.Users.GetUserByID(ctx, dbUUID(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return row.DisplayName, nil
}
