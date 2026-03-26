package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	viper.AutomaticEnv()

	pgDSN := viper.GetString("PG_DSN")
	if pgDSN == "" {
		log.Fatal("PG_DSN が必要です")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("pgxpoolの初期化に失敗しました: %v", err)
	}
	defer pool.Close()

	now := time.Now()
	seedUsers := []seedUser{
		{
			Email:       "a@example.com",
			Password:    "aexample",
			DisplayName: "シードユーザa",
			Verified:    true,
		},
		{
			Email:       "b@example.com",
			Password:    "bexample",
			DisplayName: "シードユーザb",
			Verified:    true,
		},
		{
			Email:       "c@example.com",
			Password:    "cexample",
			DisplayName: "シードユーザc",
			Verified:    false,
		},
	}

	for _, u := range seedUsers {
		uid, err := upsertUser(ctx, pool, u, now)
		if err != nil {
			log.Fatalf("ユーザ作成に失敗しました: %v", err)
		}
		fmt.Printf("ユーザ作成・更新: %s (id=%s)\n", u.Email, uid.String())

		if !u.Verified {
			if err := resetEmailVerificationToken(
				ctx,
				pool,
				uid,
				now.Add(24*time.Hour),
			); err != nil {
				log.Fatalf("メール認証トークンの作成に失敗しました: %v", err)
			}
		}
	}
}

type seedUser struct {
	Email       string
	Password    string
	DisplayName string
	Verified    bool
}

func upsertUser(
	ctx context.Context,
	pool *pgxpool.Pool,
	user seedUser,
	now time.Time,
) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("bcryptに失敗しました: %w", err)
	}

	verifiedAt := pgtype.Timestamptz{Valid: false}
	if user.Verified {
		verifiedAt = pgtype.Timestamptz{Time: now, Valid: true}
	}

	userID := uuid.New()
	row := pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, verified_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
		    display_name = EXCLUDED.display_name,
		    verified_at = EXCLUDED.verified_at
		RETURNING id
	`, userID, user.Email, string(hash), user.DisplayName, verifiedAt)

	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return uuid.UUID{}, fmt.Errorf("ユーザのUPSERTに失敗しました: %w", err)
	}
	return id, nil
}

func resetEmailVerificationToken(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	expiresAt time.Time,
) error {
	_, err := pool.Exec(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("既存トークンの削除に失敗しました: %w", err)
	}

	tokenID := uuid.New()
	token := uuid.NewString()
	_, err = pool.Exec(ctx, `
		INSERT INTO email_verification_tokens (id, user_id, token, expires_at, used_at)
		VALUES ($1, $2, $3, $4, NULL)
	`, tokenID, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("トークンの挿入に失敗しました: %w", err)
	}
	return nil
}
