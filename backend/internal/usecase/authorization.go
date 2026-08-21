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
	role, err := resolveRole(ctx, memberRepo, formID, userID)
	if err != nil {
		return err
	}
	if !role.CanEdit() {
		return entity.NewError(entity.CodeResourceHidden)
	}
	return nil
}

func requireAdmin(
	ctx context.Context,
	memberRepo repository.MemberRepository,
	formID, userID uuid.UUID,
) error {
	role, err := resolveRole(ctx, memberRepo, formID, userID)
	if err != nil {
		return err
	}
	if !role.CanAdmin() {
		return entity.NewError(entity.CodeForbidden)
	}
	return nil
}

func resolveRole(
	ctx context.Context,
	memberRepo repository.MemberRepository,
	formID, userID uuid.UUID,
) (entity.Role, error) {
	role, err := memberRepo.GetRole(ctx, formID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", entity.NewError(entity.CodeResourceHidden)
		}
		return "", err
	}
	return role, nil
}
