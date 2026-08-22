package entity

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID         uuid.UUID
	FormID     uuid.UUID
	Email      string
	Role       Role
	InvitedBy  uuid.UUID
	AcceptedAt *time.Time
	ExpiresAt  time.Time
	CreatedAt  time.Time
}
