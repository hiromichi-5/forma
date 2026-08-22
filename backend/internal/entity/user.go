package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRef struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	DisplayName  string
	VerifiedAt   *time.Time
	CreatedAt    time.Time
}

type EmailVerificationToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
}

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
}
