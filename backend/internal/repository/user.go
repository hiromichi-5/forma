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
}
