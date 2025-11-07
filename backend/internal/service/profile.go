package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type ProfileStore interface {
	GetUser(ctx context.Context, id pgtype.UUID) (db.User, error)
	UpdateUserDisplayName(ctx context.Context, arg db.UpdateUserDisplayNameParams) (db.User, error)
	DeleteUser(ctx context.Context, id pgtype.UUID) (int64, error)
	UpdateUserPasswordHash(ctx context.Context, arg db.UpdateUserPasswordHashParams) error
}

type ProfileService struct {
	q ProfileStore
}

func NewProfileService(q ProfileStore) *ProfileService {
	return &ProfileService{q: q}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID string) (db.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.User{}, ErrValidation
	}

	user, err := s.q.GetUser(ctx, dbUUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, err
	}

	return user, nil
}

func (s *ProfileService) UpdateDisplayName(ctx context.Context, userID, displayName string) (db.User, error) {
	if displayName == "" {
		return db.User{}, ErrValidation
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.User{}, ErrValidation
	}

	user, err := s.q.UpdateUserDisplayName(ctx, db.UpdateUserDisplayNameParams{
		ID:          dbUUID(uid),
		DisplayName: displayName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, err
	}

	return user, nil
}

func (s *ProfileService) DeleteProfile(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrValidation
	}

	rows, err := s.q.DeleteUser(ctx, dbUUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *ProfileService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if currentPassword == "" || newPassword == "" {
		return ErrValidation
	}

	if len(newPassword) < 8 {
		return ErrValidation
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrValidation
	}

	user, err := s.q.GetUser(ctx, dbUUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)) != nil {
		return ErrIncorrectPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.q.UpdateUserPasswordHash(ctx, db.UpdateUserPasswordHashParams{
		ID:           user.ID,
		PasswordHash: string(hashed),
	}); err != nil {
		return err
	}

	return nil
}
