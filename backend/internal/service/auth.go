package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error)
	CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error)
	DeleteSession(ctx context.Context, id pgtype.UUID) (int64, error)
	CreateEmailVerificationToken(ctx context.Context, arg db.CreateEmailVerificationTokenParams) (db.EmailVerificationToken, error)
	GetEmailVerificationTokenByToken(ctx context.Context, token string) (db.EmailVerificationToken, error)
	UseEmailVerificationToken(ctx context.Context, id pgtype.UUID) (int64, error)
	DeleteEmailVerificationTokensByUser(ctx context.Context, userID pgtype.UUID) error
	SetUserVerifiedAt(ctx context.Context, arg db.SetUserVerifiedAtParams) error
	CreatePasswordResetToken(ctx context.Context, arg db.CreatePasswordResetTokenParams) (db.PasswordResetToken, error)
	GetPasswordResetTokenByToken(ctx context.Context, token string) (db.PasswordResetToken, error)
	UsePasswordResetToken(ctx context.Context, id pgtype.UUID) (int64, error)
	DeletePasswordResetTokensByUser(ctx context.Context, userID pgtype.UUID) error
	UpdateUserPasswordHash(ctx context.Context, arg db.UpdateUserPasswordHashParams) error
}

type AuthService struct {
	q             AuthStore
	now           func() time.Time
	generateToken func() (string, error)
}

const tokenTTL = 24 * time.Hour

func NewAuthService(q AuthStore) *AuthService {
	return &AuthService{
		q:             q,
		now:           time.Now,
		generateToken: defaultToken,
	}
}

func defaultToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	return strings.ToLower(token), nil
}

func (s *AuthService) nowTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *AuthService) newToken() (string, error) {
	if s.generateToken != nil {
		return s.generateToken()
	}
	return defaultToken()
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (string, error) {
	if email == "" || password == "" {
		return "", ErrValidation
	}
	u, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	if !u.VerifiedAt.Valid {
		return "", ErrEmailNotVerified
	}

	sid := uuid.New()
	_, err = s.q.CreateSession(ctx, db.CreateSessionParams{
		ID:     dbUUID(sid),
		UserID: u.ID,
	})
	if err != nil {
		return "", err
	}
	return sid.String(), nil
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
		VerifiedAt:   pgtype.Timestamptz{},
	})
	if err != nil {
		return "", ErrConflict
	}
	if err := s.issueEmailVerificationToken(ctx, u.ID); err != nil {
		return "", err
	}
	return uuid.UUID(u.ID.Bytes).String(), nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return ErrValidation
	}
	rows, err := s.q.DeleteSession(ctx, dbUUID(sid))
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvalidCredentials
	}
	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return ErrValidation
	}
	t, err := s.q.GetEmailVerificationTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTokenNotFound
		}
		return err
	}
	used, err := s.q.UseEmailVerificationToken(ctx, t.ID)
	if err != nil {
		return err
	}
	if used == 0 {
		return ErrTokenNotFound
	}
	return s.q.SetUserVerifiedAt(ctx, db.SetUserVerifiedAtParams{
		ID:         t.UserID,
		VerifiedAt: pgtype.Timestamptz{Time: s.nowTime(), Valid: true},
	})
}

func (s *AuthService) ResendEmailVerification(ctx context.Context, email string) error {
	if email == "" {
		return ErrValidation
	}
	u, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if u.VerifiedAt.Valid {
		return nil
	}
	if err := s.q.DeleteEmailVerificationTokensByUser(ctx, u.ID); err != nil {
		return err
	}
	return s.issueEmailVerificationToken(ctx, u.ID)
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	if email == "" {
		return ErrValidation
	}
	u, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if err := s.q.DeletePasswordResetTokensByUser(ctx, u.ID); err != nil {
		return err
	}
	return s.issuePasswordResetToken(ctx, u.ID)
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if token == "" || newPassword == "" {
		return ErrValidation
	}
	if len(newPassword) < 8 {
		return ErrValidation
	}
	t, err := s.q.GetPasswordResetTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTokenNotFound
		}
		return err
	}
	used, err := s.q.UsePasswordResetToken(ctx, t.ID)
	if err != nil {
		return err
	}
	if used == 0 {
		return ErrTokenNotFound
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.q.UpdateUserPasswordHash(ctx, db.UpdateUserPasswordHashParams{
		ID:           t.UserID,
		PasswordHash: string(hashed),
	}); err != nil {
		return err
	}
	return s.q.DeletePasswordResetTokensByUser(ctx, t.UserID)
}

func (s *AuthService) issueEmailVerificationToken(ctx context.Context, userID pgtype.UUID) error {
	token, err := s.newToken()
	if err != nil {
		return err
	}
	_, err = s.q.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
		ID:        dbUUID(uuid.New()),
		UserID:    userID,
		Token:     token,
		ExpiresAt: pgtype.Timestamptz{Time: s.nowTime().Add(tokenTTL), Valid: true},
	})
	return err
}

func (s *AuthService) issuePasswordResetToken(ctx context.Context, userID pgtype.UUID) error {
	token, err := s.newToken()
	if err != nil {
		return err
	}
	_, err = s.q.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		ID:        dbUUID(uuid.New()),
		UserID:    userID,
		Token:     token,
		ExpiresAt: pgtype.Timestamptz{Time: s.nowTime().Add(tokenTTL), Valid: true},
	})
	return err
}

func dbUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }
