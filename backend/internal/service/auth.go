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

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

type AuthService struct{ q AuthStore }

func NewAuthService(q AuthStore) *AuthService { return &AuthService{q: q} }

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (string, error) {
	u, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	return uuid.UUID(u.ID.Bytes).String(), nil
}

func (s *AuthService) Signup(ctx context.Context, email, password, displayName string) (string, error) {
	if email == "" || password == "" || displayName == "" {
		return "", ErrValidation
	}
	if _, err := s.q.GetUserByEmail(ctx, email); err == nil {
		return "", ErrConflict
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	uid := uuid.New()
	u, err := s.q.CreateUser(ctx, db.CreateUserParams{
		ID:           dbUUID(uid),
		Email:        email,
		PasswordHash: string(hashed),
		DisplayName:  displayName,
	})
	if err != nil {
		return "", ErrConflict
	}
	return uuid.UUID(u.ID.Bytes).String(), nil
}

func dbUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }
