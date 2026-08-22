package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type EmailVerificationTokenRepository interface {
	Create(
		ctx context.Context,
		token entity.EmailVerificationToken,
	) (entity.EmailVerificationToken, error)
	// GetByToken は未使用かつ有効期限内のトークンのみを返す。
	GetByToken(ctx context.Context, token string) (entity.EmailVerificationToken, error)
	Use(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
}

type PasswordResetTokenRepository interface {
	Create(
		ctx context.Context,
		token entity.PasswordResetToken,
	) (entity.PasswordResetToken, error)
	// GetByToken は未使用かつ有効期限内のトークンのみを返す。
	GetByToken(ctx context.Context, token string) (entity.PasswordResetToken, error)
	Use(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
}
