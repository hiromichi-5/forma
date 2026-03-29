package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type MemberRepository interface {
	Upsert(ctx context.Context, userID uuid.UUID, formID uuid.UUID, role string) error
	Delete(ctx context.Context, userID uuid.UUID, formID uuid.UUID) error
	GetRole(ctx context.Context, userID uuid.UUID, formID uuid.UUID) (string, error)
	List(ctx context.Context, formID uuid.UUID) ([]entity.Member, error)
	CountAdmins(ctx context.Context, formID uuid.UUID) (int64, error)
	ListAccessibleForms(ctx context.Context, userID uuid.UUID) ([]entity.Form, error)
}
