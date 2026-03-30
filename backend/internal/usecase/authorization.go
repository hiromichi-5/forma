package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

func requireEditor(
	ctx context.Context,
	memberRepo repository.MemberRepository,
	formID, userID uuid.UUID,
) error {
	role, err := memberRepo.GetRole(ctx, userID, formID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeForbidden)
		}
		return err
	}
	if role != entity.RoleAdmin && role != entity.RoleEditor {
		return entity.NewError(entity.CodeForbidden)
	}
	return nil
}

func requireAdmin(
	ctx context.Context,
	memberRepo repository.MemberRepository,
	formID, userID uuid.UUID,
) error {
	role, err := memberRepo.GetRole(ctx, userID, formID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeForbidden)
		}
		return err
	}
	if role != entity.RoleAdmin {
		return entity.NewError(entity.CodeForbidden)
	}
	return nil
}
