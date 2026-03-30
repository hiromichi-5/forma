package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	q *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: db.New(pool)}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	row, err := r.q.GetUserByID(ctx, toUUID(id))
	if err != nil {
		return entity.User{}, repoError(err)
	}
	return toUserFromGetByID(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return entity.User{}, repoError(err)
	}
	return toUserFromGetByEmail(row), nil
}

func (r *UserRepository) Create(ctx context.Context, user entity.User) (entity.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           toUUID(user.ID),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		VerifiedAt:   toTimestamptzPtr(user.VerifiedAt),
	})
	if err != nil {
		return entity.User{}, repoError(err)
	}
	return toUserFromCreate(row), nil
}

func (r *UserRepository) UpdateDisplayName(
	ctx context.Context,
	id uuid.UUID,
	displayName string,
) (entity.User, error) {
	row, err := r.q.UpdateUserDisplayName(ctx, db.UpdateUserDisplayNameParams{
		ID:          toUUID(id),
		DisplayName: displayName,
	})
	if err != nil {
		return entity.User{}, repoError(err)
	}
	return toUserFromUpdateDisplayName(row), nil
}

func (r *UserRepository) UpdatePasswordHash(
	ctx context.Context,
	id uuid.UUID,
	passwordHash string,
) error {
	n, err := r.q.UpdateUserPasswordHash(ctx, db.UpdateUserPasswordHashParams{
		ID:           toUUID(id),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.DeleteUser(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *UserRepository) SetVerifiedAt(
	ctx context.Context,
	id uuid.UUID,
	verifiedAt time.Time,
) error {
	n, err := r.q.SetUserVerifiedAt(ctx, db.SetUserVerifiedAtParams{
		ID:         toUUID(id),
		VerifiedAt: toTimestamptz(verifiedAt),
	})
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *UserRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (entity.Session, error) {
	row, err := r.q.GetSessionByID(ctx, toUUID(id))
	if err != nil {
		return entity.Session{}, repoError(err)
	}
	return toSession(row), nil
}

func (r *UserRepository) CreateSession(
	ctx context.Context,
	session entity.Session,
) (entity.Session, error) {
	row, err := r.q.CreateSession(ctx, db.CreateSessionParams{
		ID:     toUUID(session.ID),
		UserID: toUUID(session.UserID),
	})
	if err != nil {
		return entity.Session{}, repoError(err)
	}
	return toSession(row), nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.DeleteSession(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *UserRepository) CreateEmailVerificationToken(
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

func (r *UserRepository) GetEmailVerificationTokenByToken(
	ctx context.Context,
	token string,
) (entity.EmailVerificationToken, error) {
	row, err := r.q.GetEmailVerificationTokenByToken(ctx, token)
	if err != nil {
		return entity.EmailVerificationToken{}, repoError(err)
	}
	return toEmailVerificationToken(row), nil
}

func (r *UserRepository) UseEmailVerificationToken(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.UseEmailVerificationToken(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *UserRepository) DeleteEmailVerificationTokensByUser(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return r.q.DeleteEmailVerificationTokensByUser(ctx, toUUID(userID))
}

func (r *UserRepository) CreatePasswordResetToken(
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

func (r *UserRepository) GetPasswordResetTokenByToken(
	ctx context.Context,
	token string,
) (entity.PasswordResetToken, error) {
	row, err := r.q.GetPasswordResetTokenByToken(ctx, token)
	if err != nil {
		return entity.PasswordResetToken{}, repoError(err)
	}
	return toPasswordResetToken(row), nil
}

func (r *UserRepository) UsePasswordResetToken(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.UsePasswordResetToken(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *UserRepository) DeletePasswordResetTokensByUser(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return r.q.DeletePasswordResetTokensByUser(ctx, toUUID(userID))
}

func toUserFromGetByID(row db.GetUserByIDRow) entity.User {
	return entity.User{
		ID:           fromUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		VerifiedAt:   fromTimestamptzPtr(row.VerifiedAt),
		CreatedAt:    fromTimestamptz(row.CreatedAt),
	}
}

func toUserFromGetByEmail(row db.GetUserByEmailRow) entity.User {
	return entity.User{
		ID:           fromUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		VerifiedAt:   fromTimestamptzPtr(row.VerifiedAt),
		CreatedAt:    fromTimestamptz(row.CreatedAt),
	}
}

func toUserFromCreate(row db.CreateUserRow) entity.User {
	return entity.User{
		ID:           fromUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		VerifiedAt:   fromTimestamptzPtr(row.VerifiedAt),
		CreatedAt:    fromTimestamptz(row.CreatedAt),
	}
}

func toUserFromUpdateDisplayName(row db.UpdateUserDisplayNameRow) entity.User {
	return entity.User{
		ID:           fromUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		VerifiedAt:   fromTimestamptzPtr(row.VerifiedAt),
		CreatedAt:    fromTimestamptz(row.CreatedAt),
	}
}

func toSession(row db.Session) entity.Session {
	return entity.Session{
		ID:        fromUUID(row.ID),
		UserID:    fromUUID(row.UserID),
		CreatedAt: fromTimestamptz(row.CreatedAt),
	}
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
