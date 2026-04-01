package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func CreateVerifiedUser(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	email, password, displayName string,
) uuid.UUID {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	userID := uuid.New()
	now := time.Now()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, verified_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, email, string(hashed), displayName, now, now)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return userID
}

func GetEmailVerificationToken(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
) string {
	t.Helper()

	var token string
	err := pool.QueryRow(ctx, `
		SELECT token FROM email_verification_tokens
		WHERE user_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&token)
	if err != nil {
		t.Fatalf("get email verification token: %v", err)
	}
	return token
}

func GetPasswordResetToken(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
) string {
	t.Helper()

	var token string
	err := pool.QueryRow(ctx, `
		SELECT token FROM password_reset_tokens
		WHERE user_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&token)
	if err != nil {
		t.Fatalf("get password reset token: %v", err)
	}
	return token
}
