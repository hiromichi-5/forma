package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

type StatusUseCase struct {
	statusRepo repository.StatusRepository
	memberRepo repository.MemberRepository
	ticketRepo repository.TicketRepository
	txm        repository.TxManager
}

func NewStatusUseCase(
	statusRepo repository.StatusRepository,
	memberRepo repository.MemberRepository,
	ticketRepo repository.TicketRepository,
	txm repository.TxManager,
) *StatusUseCase {
	return &StatusUseCase{
		statusRepo: statusRepo,
		memberRepo: memberRepo,
		ticketRepo: ticketRepo,
		txm:        txm,
	}
}

func (uc *StatusUseCase) ListStatuses(
	ctx context.Context,
	formID, userID uuid.UUID,
) ([]entity.FormStatus, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return nil, err
	}
	return uc.statusRepo.List(ctx, formID)
}

func (uc *StatusUseCase) CreateStatus(
	ctx context.Context,
	formID, userID uuid.UUID,
	name string,
	color *string,
	displayOrder int32,
	isDefault bool,
) (entity.FormStatus, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return entity.FormStatus{}, err
	}

	name = strings.TrimSpace(name)
	if name == "" || displayOrder <= 0 {
		return entity.FormStatus{}, entity.NewError(entity.CodeValidation)
	}

	var trimmedColor *string
	if color != nil && strings.TrimSpace(*color) != "" {
		c := strings.TrimSpace(*color)
		trimmedColor = &c
	}

	statusID := uuid.New()

	if !isDefault {
		status, err := uc.statusRepo.Create(ctx, entity.FormStatus{
			ID:           statusID,
			FormID:       formID,
			Name:         name,
			Color:        trimmedColor,
			DisplayOrder: displayOrder,
			IsDefault:    false,
		})
		if err != nil {
			if errors.Is(err, repository.ErrConflict) {
				return entity.FormStatus{}, entity.NewError(entity.CodeValidation)
			}
			return entity.FormStatus{}, err
		}
		return status, nil
	}

	var status entity.FormStatus
	if err := uc.txm.Do(ctx, func(repos repository.Repos) error {
		var createErr error
		status, createErr = repos.Status.Create(ctx, entity.FormStatus{
			ID:           statusID,
			FormID:       formID,
			Name:         name,
			Color:        trimmedColor,
			DisplayOrder: displayOrder,
			IsDefault:    false,
		})
		if createErr != nil {
			return createErr
		}
		if err := repos.Status.ClearDefault(ctx, formID); err != nil {
			return err
		}
		var setErr error
		status, setErr = repos.Status.SetDefault(ctx, formID, statusID)
		return setErr
	}); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return entity.FormStatus{}, entity.NewError(entity.CodeValidation)
		}
		return entity.FormStatus{}, err
	}

	return status, nil
}

func (uc *StatusUseCase) UpdateStatus(
	ctx context.Context,
	formID, userID, statusID uuid.UUID,
	name, color *string,
	displayOrder *int32,
) (entity.FormStatus, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return entity.FormStatus{}, err
	}

	current, err := uc.statusRepo.GetByID(ctx, statusID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.FormStatus{}, entity.NewError(entity.CodeForbidden)
		}
		return entity.FormStatus{}, err
	}
	if current.FormID != formID {
		return entity.FormStatus{}, entity.NewError(entity.CodeForbidden)
	}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return entity.FormStatus{}, entity.NewError(entity.CodeValidation)
		}
		current.Name = trimmed
	}

	if color != nil {
		trimmed := strings.TrimSpace(*color)
		if trimmed == "" {
			current.Color = nil
		} else {
			current.Color = &trimmed
		}
	}

	if displayOrder != nil {
		if *displayOrder <= 0 {
			return entity.FormStatus{}, entity.NewError(entity.CodeValidation)
		}
		current.DisplayOrder = *displayOrder
	}

	updated, err := uc.statusRepo.Update(ctx, current)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.FormStatus{}, entity.NewError(entity.CodeForbidden)
		}
		if errors.Is(err, repository.ErrConflict) {
			return entity.FormStatus{}, entity.NewError(entity.CodeValidation)
		}
		return entity.FormStatus{}, err
	}
	return updated, nil
}

func (uc *StatusUseCase) SetDefaultStatus(
	ctx context.Context,
	formID, userID, statusID uuid.UUID,
) error {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return err
	}

	return uc.txm.Do(ctx, func(repos repository.Repos) error {
		if err := repos.Status.ClearDefault(ctx, formID); err != nil {
			return err
		}
		_, err := repos.Status.SetDefault(ctx, formID, statusID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeForbidden)
			}
			return err
		}
		return nil
	})
}

func (uc *StatusUseCase) DeleteStatus(
	ctx context.Context,
	formID, userID, statusID uuid.UUID,
) error {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return err
	}

	status, err := uc.statusRepo.GetByID(ctx, statusID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeForbidden)
		}
		return err
	}
	if status.FormID != formID {
		return entity.NewError(entity.CodeForbidden)
	}
	if status.IsDefault {
		return entity.NewError(entity.CodeValidation)
	}

	count, err := uc.ticketRepo.CountByStatus(ctx, statusID)
	if err != nil {
		return err
	}
	if count > 0 {
		return entity.NewError(entity.CodeValidation)
	}

	if err := uc.statusRepo.Delete(ctx, statusID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeForbidden)
		}
		return err
	}
	return nil
}
