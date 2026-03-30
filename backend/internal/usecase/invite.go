package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

const inviteTTL = 7 * 24 * time.Hour

type InviteUseCase struct {
	inviteRepo repository.InviteRepository
	memberRepo repository.MemberRepository
	userRepo   repository.UserRepository
	txm        repository.TxManager
	now        func() time.Time
}

func NewInviteUseCase(
	inviteRepo repository.InviteRepository,
	memberRepo repository.MemberRepository,
	userRepo repository.UserRepository,
	txm repository.TxManager,
) *InviteUseCase {
	return &InviteUseCase{
		inviteRepo: inviteRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		txm:        txm,
		now:        time.Now,
	}
}

func (uc *InviteUseCase) CreateInvite(
	ctx context.Context,
	formID, userID uuid.UUID,
	email, role string,
) (entity.Invite, error) {
	if role != entity.RoleAdmin && role != entity.RoleEditor {
		return entity.Invite{}, entity.NewError(entity.CodeValidation)
	}

	if err := uc.requireAdmin(ctx, formID, userID); err != nil {
		return entity.Invite{}, err
	}

	invite, err := uc.inviteRepo.Create(ctx, entity.Invite{
		ID:        uuid.New(),
		FormID:    formID,
		Email:     email,
		Role:      role,
		InvitedBy: userID,
		ExpiresAt: uc.now().Add(inviteTTL),
	})
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return entity.Invite{}, entity.NewError(entity.CodeValidation)
		}
		return entity.Invite{}, err
	}
	return invite, nil
}

func (uc *InviteUseCase) ListInvites(
	ctx context.Context,
	formID, userID uuid.UUID,
) ([]entity.Invite, error) {
	if err := uc.requireAdmin(ctx, formID, userID); err != nil {
		return nil, err
	}
	return uc.inviteRepo.ListActive(ctx, formID)
}

func (uc *InviteUseCase) DeleteInvite(
	ctx context.Context,
	formID, userID, inviteID uuid.UUID,
) error {
	if err := uc.requireAdmin(ctx, formID, userID); err != nil {
		return err
	}

	return uc.txm.Do(ctx, func(repos repository.Repos) error {
		invite, err := repos.Invite.GetForUpdate(ctx, inviteID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeInviteNotFound)
			}
			return err
		}

		if invite.FormID != formID {
			return entity.NewError(entity.CodeInviteNotFound)
		}

		return repos.Invite.Delete(ctx, inviteID)
	})
}

func (uc *InviteUseCase) AcceptInvite(ctx context.Context, inviteID, userID uuid.UUID) error {
	return uc.txm.Do(ctx, func(repos repository.Repos) error {
		invite, err := repos.Invite.GetForUpdate(ctx, inviteID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeInviteNotFound)
			}
			return err
		}

		if invite.AcceptedAt != nil {
			return entity.NewError(entity.CodeInviteNotFound)
		}
		if !invite.ExpiresAt.After(uc.now()) {
			return entity.NewError(entity.CodeInviteExpired)
		}

		user, err := repos.User.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeUserNotFound)
			}
			return err
		}

		if user.Email != invite.Email {
			return entity.NewError(entity.CodeForbidden)
		}

		if _, err := repos.Invite.Accept(ctx, inviteID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeInviteNotFound)
			}
			return err
		}

		return repos.Member.Upsert(ctx, userID, invite.FormID, invite.Role)
	})
}

func (uc *InviteUseCase) requireAdmin(ctx context.Context, formID, userID uuid.UUID) error {
	role, err := uc.memberRepo.GetRole(ctx, userID, formID)
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
