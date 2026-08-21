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

type TicketHistory struct {
	ID            uuid.UUID
	TicketID      uuid.UUID
	ChangedBy     *uuid.UUID
	ChangedByName string
	FieldName     string
	OldValue      *string
	NewValue      *string
	CreatedAt     time.Time
}
