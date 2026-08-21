package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/logger"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

const inviteTTL = 7 * 24 * time.Hour

type InviteUseCase struct {
	inviteRepo      repository.InviteRepository
	memberRepo      repository.MemberRepository
	userRepo        repository.UserRepository
	uow             repository.UnitOfWork[repository.InviteRepos]
	emailSender     repository.EmailSender
	frontendBaseURL string
	now             func() time.Time
}

func NewInviteUseCase(
	inviteRepo repository.InviteRepository,
	memberRepo repository.MemberRepository,
	userRepo repository.UserRepository,
	uow repository.UnitOfWork[repository.InviteRepos],
	emailSender repository.EmailSender,
	frontendBaseURL string,
) *InviteUseCase {
	return &InviteUseCase{
		inviteRepo:      inviteRepo,
		memberRepo:      memberRepo,
		userRepo:        userRepo,
		uow:             uow,
		emailSender:     emailSender,
		frontendBaseURL: frontendBaseURL,
		now:             time.Now,
	}
}

func (uc *InviteUseCase) CreateInvite(
	ctx context.Context,
	formID, userID uuid.UUID,
	email string,
	role entity.Role,
) (entity.Invite, error) {
	if !role.Valid() {
		return entity.Invite{}, entity.NewError(entity.CodeValidation)
	}

	if err := requireAdmin(ctx, uc.memberRepo, formID, userID); err != nil {
		return entity.Invite{}, err
	}

	if err := uc.ensureEmailNotMember(ctx, formID, email); err != nil {
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
			return entity.Invite{}, entity.NewError(entity.CodeActiveInviteAlreadyExists)
		}
		return entity.Invite{}, err
	}

	acceptURL := uc.frontendBaseURL + "/invites/" + invite.ID.String() + "/accept"
	if err := uc.emailSender.SendEmail(ctx, repository.SendEmailInput{
		To:           []string{invite.Email},
		TemplateName: repository.TemplateInvite,
		TemplateData: map[string]string{
			"accept_url": acceptURL,
			"role":       string(invite.Role),
		},
	}); err != nil {
		_ = uc.inviteRepo.Delete(ctx, invite.ID)
		return entity.Invite{}, err
	}

	logger.From(ctx).Info("invite created",
		"invite_id", invite.ID.String(),
		"form_id", formID.String(),
	)

	return invite, nil
}

func (uc *InviteUseCase) ListInvites(
	ctx context.Context,
	formID, userID uuid.UUID,
) ([]entity.Invite, error) {
	if err := requireAdmin(ctx, uc.memberRepo, formID, userID); err != nil {
		return nil, err
	}
	return uc.inviteRepo.ListActive(ctx, formID)
}

func (uc *InviteUseCase) DeleteInvite(
	ctx context.Context,
	formID, userID, inviteID uuid.UUID,
) error {
	if err := requireAdmin(ctx, uc.memberRepo, formID, userID); err != nil {
		return err
	}

	return uc.uow.Do(ctx, func(repos repository.InviteRepos) error {
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

		if err := repos.Invite.Delete(ctx, inviteID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeInviteNotFound)
			}
			return err
		}
		return nil
	})
}

func (uc *InviteUseCase) AcceptInvite(ctx context.Context, inviteID, userID uuid.UUID) error {
	return uc.uow.Do(ctx, func(repos repository.InviteRepos) error {
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
			return entity.NewError(entity.CodeResourceHidden)
		}

		if err := ensureUserNotMember(ctx, repos.Member, invite.FormID, userID); err != nil {
			return err
		}

		if _, err := repos.Invite.Accept(ctx, inviteID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeInviteNotFound)
			}
			return err
		}

		if err := repos.Member.Upsert(ctx, invite.FormID, userID, invite.Role); err != nil {
			return err
		}

		logger.From(ctx).Info("invite accepted",
			"invite_id", inviteID.String(),
			"form_id", invite.FormID.String(),
		)

		return nil
	})
}

func (uc *InviteUseCase) ensureEmailNotMember(
	ctx context.Context,
	formID uuid.UUID,
	email string,
) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}

	return ensureUserNotMember(ctx, uc.memberRepo, formID, user.ID)
}

func ensureUserNotMember(
	ctx context.Context,
	memberRepo repository.MemberRepository,
	formID, userID uuid.UUID,
) error {
	_, err := memberRepo.GetRole(ctx, formID, userID)
	if err == nil {
		return entity.NewError(entity.CodeAlreadyMember)
	}
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	return err
}
