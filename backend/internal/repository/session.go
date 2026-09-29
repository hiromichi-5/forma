package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type SessionRepository interface {
	GetByToken(ctx context.Context, token string) (entity.Session, error)
	Create(ctx context.Context, session entity.Session) (entity.Session, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	DeleteByUserExcept(ctx context.Context, userID, sessionID uuid.UUID) error
}
