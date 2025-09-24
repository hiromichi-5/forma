package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProfileStore interface {
	GetUser(ctx context.Context, id pgtype.UUID) (db.User, error)
	UpdateUserDisplayName(ctx context.Context, arg db.UpdateUserDisplayNameParams) (db.User, error)
	DeleteUser(ctx context.Context, id pgtype.UUID) error
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

	err = s.q.DeleteUser(ctx, dbUUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
