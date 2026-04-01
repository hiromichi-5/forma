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
	uow        repository.UnitOfWork[repository.StatusRepos]
}

func NewStatusUseCase(
	statusRepo repository.StatusRepository,
	memberRepo repository.MemberRepository,
	ticketRepo repository.TicketRepository,
	uow repository.UnitOfWork[repository.StatusRepos],
) *StatusUseCase {
	return &StatusUseCase{
		statusRepo: statusRepo,
		memberRepo: memberRepo,
		ticketRepo: ticketRepo,
		uow:        uow,
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
				return entity.FormStatus{}, entity.NewError(entity.CodeStatusConflict)
			}
			return entity.FormStatus{}, err
		}
		return status, nil
	}

	var status entity.FormStatus
	if err := uc.uow.Do(ctx, func(repos repository.StatusRepos) error {
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
			return entity.FormStatus{}, entity.NewError(entity.CodeStatusConflict)
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
	isDefault *bool,
) (entity.FormStatus, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return entity.FormStatus{}, err
	}

	if isDefault != nil && !*isDefault {
		return entity.FormStatus{}, entity.NewError(entity.CodeValidation)
	}

	current, err := uc.statusRepo.GetByID(ctx, statusID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.FormStatus{}, entity.NewError(entity.CodeResourceHidden)
		}
		return entity.FormStatus{}, err
	}
	if current.FormID != formID {
		return entity.FormStatus{}, entity.NewError(entity.CodeResourceHidden)
	}

	next, err := applyStatusUpdate(current, name, color, displayOrder)
	if err != nil {
		return entity.FormStatus{}, err
	}

	if isDefault == nil {
		updated, updateErr := uc.statusRepo.Update(ctx, next)
		if updateErr != nil {
			if errors.Is(updateErr, repository.ErrNotFound) {
				return entity.FormStatus{}, entity.NewError(entity.CodeResourceHidden)
			}
			if errors.Is(updateErr, repository.ErrConflict) {
				return entity.FormStatus{}, entity.NewError(entity.CodeStatusConflict)
			}
			return entity.FormStatus{}, updateErr
		}
		return updated, nil
	}

	var updated entity.FormStatus
	if err := uc.uow.Do(ctx, func(repos repository.StatusRepos) error {
		var updateErr error
		updated, updateErr = repos.Status.Update(ctx, next)
		if updateErr != nil {
			return updateErr
		}
		if err := repos.Status.ClearDefault(ctx, formID); err != nil {
			return err
		}
		updated, updateErr = repos.Status.SetDefault(ctx, formID, statusID)
		if updateErr != nil {
			return updateErr
		}
		return nil
	}); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.FormStatus{}, entity.NewError(entity.CodeResourceHidden)
		}
		if errors.Is(err, repository.ErrConflict) {
			return entity.FormStatus{}, entity.NewError(entity.CodeStatusConflict)
		}
		return entity.FormStatus{}, err
	}

	return updated, nil
}

func applyStatusUpdate(
	current entity.FormStatus,
	name, color *string,
	displayOrder *int32,
) (entity.FormStatus, error) {
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

	return current, nil
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
			return entity.NewError(entity.CodeResourceHidden)
		}
		return err
	}
	if status.FormID != formID {
		return entity.NewError(entity.CodeResourceHidden)
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
			return entity.NewError(entity.CodeResourceHidden)
		}
		return err
	}
	return nil
}
