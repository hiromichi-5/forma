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
