package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ repository.EmailVerificationTokenRepository = (*EmailVerificationTokenRepository)(nil)
	_ repository.PasswordResetTokenRepository     = (*PasswordResetTokenRepository)(nil)
)

type EmailVerificationTokenRepository struct {
	q *db.Queries
}

func NewEmailVerificationTokenRepository(pool *pgxpool.Pool) *EmailVerificationTokenRepository {
	return &EmailVerificationTokenRepository{q: db.New(pool)}
}

func (r *EmailVerificationTokenRepository) Create(
	ctx context.Context,
	token entity.EmailVerificationToken,
) (entity.EmailVerificationToken, error) {
	row, err := r.q.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
		ID:        toUUID(token.ID),
		UserID:    toUUID(token.UserID),
		Token:     token.Token,
		ExpiresAt: toTimestamptz(token.ExpiresAt),
	})
	if err != nil {
		return entity.EmailVerificationToken{}, repoError(err)
	}
	return toEmailVerificationToken(row), nil
}

func (r *EmailVerificationTokenRepository) GetByToken(
	ctx context.Context,
	token string,
) (entity.EmailVerificationToken, error) {
	row, err := r.q.GetEmailVerificationTokenByToken(ctx, token)
	if err != nil {
		return entity.EmailVerificationToken{}, repoError(err)
	}
	return toEmailVerificationToken(row), nil
}

func (r *EmailVerificationTokenRepository) Use(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.UseEmailVerificationToken(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *EmailVerificationTokenRepository) DeleteByUser(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return r.q.DeleteEmailVerificationTokensByUser(ctx, toUUID(userID))
}

type PasswordResetTokenRepository struct {
	q *db.Queries
}

func NewPasswordResetTokenRepository(pool *pgxpool.Pool) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{q: db.New(pool)}
}

func (r *PasswordResetTokenRepository) Create(
	ctx context.Context,
	token entity.PasswordResetToken,
) (entity.PasswordResetToken, error) {
	row, err := r.q.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		ID:        toUUID(token.ID),
		UserID:    toUUID(token.UserID),
		Token:     token.Token,
		ExpiresAt: toTimestamptz(token.ExpiresAt),
	})
	if err != nil {
		return entity.PasswordResetToken{}, repoError(err)
	}
	return toPasswordResetToken(row), nil
}

func (r *PasswordResetTokenRepository) GetByToken(
	ctx context.Context,
	token string,
) (entity.PasswordResetToken, error) {
	row, err := r.q.GetPasswordResetTokenByToken(ctx, token)
	if err != nil {
		return entity.PasswordResetToken{}, repoError(err)
	}
	return toPasswordResetToken(row), nil
}

func (r *PasswordResetTokenRepository) Use(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.UsePasswordResetToken(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *PasswordResetTokenRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeletePasswordResetTokensByUser(ctx, toUUID(userID))
}

func toEmailVerificationToken(row db.EmailVerificationToken) entity.EmailVerificationToken {
	return entity.EmailVerificationToken{
		ID:        fromUUID(row.ID),
		UserID:    fromUUID(row.UserID),
		Token:     row.Token,
		ExpiresAt: fromTimestamptz(row.ExpiresAt),
	}
}

func toPasswordResetToken(row db.PasswordResetToken) entity.PasswordResetToken {
	return entity.PasswordResetToken{
		ID:        fromUUID(row.ID),
		UserID:    fromUUID(row.UserID),
		Token:     row.Token,
		ExpiresAt: fromTimestamptz(row.ExpiresAt),
	}
}
