package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type ProfileUseCase struct {
	userRepo repository.UserRepository
}

func NewProfileUseCase(userRepo repository.UserRepository) *ProfileUseCase {
	return &ProfileUseCase{userRepo: userRepo}
}

func (uc *ProfileUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.User{}, entity.NewError(entity.CodeUserNotFound)
		}
		return entity.User{}, err
	}
	return user, nil
}

func (uc *ProfileUseCase) UpdateDisplayName(
	ctx context.Context,
	userID uuid.UUID,
	displayName string,
) (entity.User, error) {
	if displayName == "" {
		return entity.User{}, entity.NewError(entity.CodeValidation)
	}

	user, err := uc.userRepo.UpdateDisplayName(ctx, userID, displayName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.User{}, entity.NewError(entity.CodeUserNotFound)
		}
		return entity.User{}, err
	}
	return user, nil
}

func (uc *ProfileUseCase) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	err := uc.userRepo.Delete(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeUserNotFound)
		}
		return err
	}
	return nil
}

func (uc *ProfileUseCase) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	currentPassword, newPassword string,
) error {
	if currentPassword == "" || newPassword == "" {
		return entity.NewError(entity.CodeValidation)
	}
	if len(newPassword) < 8 {
		return entity.NewError(entity.CodeValidation)
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeUserNotFound)
		}
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)) != nil {
		return entity.NewError(entity.CodeIncorrectPassword)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePasswordHash(ctx, userID, string(hashed)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeUserNotFound)
		}
		return err
	}
	return nil
}
