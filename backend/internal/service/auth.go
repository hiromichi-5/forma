package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
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
