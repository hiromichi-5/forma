package entity

import (
	"time"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

func (p Priority) Valid() bool {
	switch p {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return true
	default:
		return false
	}
}

type Ticket struct {
	ID              uuid.UUID
	FormID          uuid.UUID
	ResponseID      string
	RespondentEmail *string
	Answers         []byte
	StatusID        uuid.UUID
	AssigneeID      *uuid.UUID
	Priority        Priority
	SubmittedAt     time.Time
	CreatedAt       time.Time
}

// AssigneeChange は担当者の「変更しない・解除する・指定する」の3状態を表す。
// ゼロ値は「変更しない」。
type AssigneeChange struct {
	specified bool
	userID    *uuid.UUID
}

func KeepAssignee() AssigneeChange {
	return AssigneeChange{}
}

func ClearAssignee() AssigneeChange {
	return AssigneeChange{specified: true}
}

func SetAssignee(userID uuid.UUID) AssigneeChange {
	return AssigneeChange{specified: true, userID: &userID}
}

func (c AssigneeChange) UserID() *uuid.UUID {
	return c.userID
}

func (t *Ticket) ChangeStatus(status *FormStatus) (old uuid.UUID, changed bool) {
	old = t.StatusID
	if status == nil || old == status.ID {
		return old, false
	}
	t.StatusID = status.ID
	return old, true
}

func (t *Ticket) ChangeAssignee(c AssigneeChange) (old *uuid.UUID, changed bool) {
	old = t.AssigneeID
	if !c.specified || uuidPtrEqual(old, c.userID) {
		return old, false
	}
	t.AssigneeID = c.userID
	return old, true
}

func (t *Ticket) ChangePriority(priority *Priority) (old Priority, changed bool) {
	old = t.Priority
	if priority == nil || old == *priority {
		return old, false
	}
	t.Priority = *priority
	return old, true
}

func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

type TicketField string

const (
	FieldStatus   TicketField = "status"
	FieldAssignee TicketField = "assignee"
	FieldPriority TicketField = "priority"
)

type TicketHistory struct {
	ID            uuid.UUID
	TicketID      uuid.UUID
	ChangedBy     *uuid.UUID
	ChangedByName string
	FieldName     TicketField
	OldValue      *string
	NewValue      *string
	CreatedAt     time.Time
}
