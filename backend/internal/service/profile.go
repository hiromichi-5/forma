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
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.GetUserByIDRow, error)
	UpdateUserDisplayName(ctx context.Context, arg db.UpdateUserDisplayNameParams) (db.UpdateUserDisplayNameRow, error)
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

	row, err := s.q.GetUserByID(ctx, dbUUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, err
	}

	return userFromGetUserByIDRow(row), nil
}

func (s *ProfileService) UpdateDisplayName(ctx context.Context, userID, displayName string) (db.User, error) {
	if displayName == "" {
		return db.User{}, ErrValidation
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.User{}, ErrValidation
	}

	row, err := s.q.UpdateUserDisplayName(ctx, db.UpdateUserDisplayNameParams{
		ID:          dbUUID(uid),
		DisplayName: displayName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, err
	}

	return userFromUpdateUserDisplayNameRow(row), nil
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

	row, err := s.q.GetUserByID(ctx, dbUUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(currentPassword)) != nil {
		return ErrIncorrectPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.q.UpdateUserPasswordHash(ctx, db.UpdateUserPasswordHashParams{
		ID:           row.ID,
		PasswordHash: string(hashed),
	}); err != nil {
		return err
	}

	return nil
}

func userFromGetUserByIDRow(row db.GetUserByIDRow) db.User {
	return db.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		VerifiedAt:   row.VerifiedAt,
		CreatedAt:    row.CreatedAt,
	}
}

func userFromUpdateUserDisplayNameRow(row db.UpdateUserDisplayNameRow) db.User {
	return db.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		VerifiedAt:   row.VerifiedAt,
		CreatedAt:    row.CreatedAt,
	}
}
