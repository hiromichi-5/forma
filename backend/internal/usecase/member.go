package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

type MemberUseCase struct {
	memberRepo repository.MemberRepository
	userRepo   repository.UserRepository
}

func NewMemberUseCase(
	memberRepo repository.MemberRepository,
	userRepo repository.UserRepository,
) *MemberUseCase {
	return &MemberUseCase{
		memberRepo: memberRepo,
		userRepo:   userRepo,
	}
}

func (uc *MemberUseCase) AddMember(
	ctx context.Context,
	formID, userID uuid.UUID,
	email string,
	role entity.Role,
) error {
	if !role.Valid() {
		return entity.NewError(entity.CodeValidation)
	}

	if err := requireAdmin(ctx, uc.memberRepo, formID, userID); err != nil {
		return err
	}

	target, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeUserNotFound)
		}
		return err
	}

	if _, err := uc.memberRepo.GetRole(ctx, formID, target.ID); err == nil {
		return entity.NewError(entity.CodeAlreadyMember)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	return uc.memberRepo.Upsert(ctx, formID, target.ID, role)
}

func (uc *MemberUseCase) ChangeRole(
	ctx context.Context,
	formID, userID, targetUserID uuid.UUID,
	role entity.Role,
) error {
	if !role.Valid() {
		return entity.NewError(entity.CodeValidation)
	}

	if err := requireAdmin(ctx, uc.memberRepo, formID, userID); err != nil {
		return err
	}

	currentRole, err := uc.memberRepo.GetRole(ctx, formID, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeResourceHidden)
		}
		return err
	}

	if currentRole == role {
		return nil
	}

	if !role.CanAdmin() {
		if err := uc.ensureFormKeepsAdmin(ctx, formID, targetUserID); err != nil {
			return err
		}
	}

	return uc.memberRepo.Upsert(ctx, formID, targetUserID, role)
}

func (uc *MemberUseCase) RemoveMember(
	ctx context.Context,
	formID, userID, targetUserID uuid.UUID,
) error {
	if err := requireAdmin(ctx, uc.memberRepo, formID, userID); err != nil {
		return err
	}

	if err := uc.ensureFormKeepsAdmin(ctx, formID, targetUserID); err != nil {
		return err
	}

	if err := uc.memberRepo.Delete(ctx, formID, targetUserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (uc *MemberUseCase) ListMembers(
	ctx context.Context,
	formID, userID uuid.UUID,
) ([]entity.Member, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return nil, err
	}
	return uc.memberRepo.List(ctx, formID)
}

func (uc *MemberUseCase) ensureFormKeepsAdmin(
	ctx context.Context,
	formID, targetUserID uuid.UUID,
) error {
	role, err := uc.memberRepo.GetRole(ctx, formID, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}
	if !role.CanAdmin() {
		return nil
	}
	count, err := uc.memberRepo.CountAdmins(ctx, formID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return entity.NewError(entity.CodeLastAdmin)
	}
	return nil
}
