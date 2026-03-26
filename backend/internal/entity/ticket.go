package entity

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID              uuid.UUID
	FormID          uuid.UUID
	ResponseID      string
	RespondentEmail *string
	Answers         []byte
	StatusID        uuid.UUID
	AssigneeID      *uuid.UUID
	Priority        string
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
