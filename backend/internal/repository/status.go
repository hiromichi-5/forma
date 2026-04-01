package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type StatusRepository interface {
	Create(ctx context.Context, status entity.FormStatus) (entity.FormStatus, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.FormStatus, error)
	GetDefault(ctx context.Context, formID uuid.UUID) (entity.FormStatus, error)
	List(ctx context.Context, formID uuid.UUID) ([]entity.FormStatus, error)
	Update(ctx context.Context, status entity.FormStatus) (entity.FormStatus, error)
	SetDefault(ctx context.Context, formID uuid.UUID, statusID uuid.UUID) (entity.FormStatus, error)
	ClearDefault(ctx context.Context, formID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
