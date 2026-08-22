package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type SessionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.Session, error)
	Create(ctx context.Context, session entity.Session) (entity.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
