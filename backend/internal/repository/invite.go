package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type InviteRepository interface {
	Create(ctx context.Context, invite entity.Invite) (entity.Invite, error)
	GetForUpdate(ctx context.Context, id uuid.UUID) (entity.Invite, error)
	Accept(ctx context.Context, id uuid.UUID) (entity.Invite, error)
	ListActive(ctx context.Context, formID uuid.UUID) ([]entity.Invite, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
