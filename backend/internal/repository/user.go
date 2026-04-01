package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	Create(ctx context.Context, user entity.User) (entity.User, error)
	UpdateDisplayName(ctx context.Context, id uuid.UUID, displayName string) (entity.User, error)
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetVerifiedAt(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error

	GetSessionByID(ctx context.Context, id uuid.UUID) (entity.Session, error)
	CreateSession(ctx context.Context, session entity.Session) (entity.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error

	CreateEmailVerificationToken(
		ctx context.Context,
		token entity.EmailVerificationToken,
	) (entity.EmailVerificationToken, error)
	GetEmailVerificationTokenByToken(
		ctx context.Context,
		token string,
	) (entity.EmailVerificationToken, error)
	UseEmailVerificationToken(ctx context.Context, id uuid.UUID) error
	DeleteEmailVerificationTokensByUser(ctx context.Context, userID uuid.UUID) error

	CreatePasswordResetToken(
		ctx context.Context,
		token entity.PasswordResetToken,
	) (entity.PasswordResetToken, error)
	GetPasswordResetTokenByToken(
		ctx context.Context,
		token string,
	) (entity.PasswordResetToken, error)
	UsePasswordResetToken(ctx context.Context, id uuid.UUID) error
	DeletePasswordResetTokensByUser(ctx context.Context, userID uuid.UUID) error
}
