package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	q *db.Queries
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{q: db.New(pool)}
}

func (r *SessionRepository) GetByToken(ctx context.Context, token string) (entity.Session, error) {
	row, err := r.q.GetSessionByTokenHash(ctx, hashToken(token))
	if err != nil {
		return entity.Session{}, repoError(err)
	}
	return toSession(row, token), nil
}

func (r *SessionRepository) Create(
	ctx context.Context,
	session entity.Session,
) (entity.Session, error) {
	row, err := r.q.CreateSession(ctx, db.CreateSessionParams{
		ID:        toUUID(session.ID),
		UserID:    toUUID(session.UserID),
		TokenHash: hashToken(session.Token),
		ExpiresAt: toTimestamptz(session.ExpiresAt),
	})
	if err != nil {
		return entity.Session{}, repoError(err)
	}
	return toSession(row, session.Token), nil
}

func (r *SessionRepository) DeleteByToken(ctx context.Context, token string) error {
	n, err := r.q.DeleteSessionByTokenHash(ctx, hashToken(token))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *SessionRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.DeleteSessionsByUser(ctx, toUUID(userID)); err != nil {
		return repoError(err)
	}
	return nil
}

func (r *SessionRepository) DeleteByUserExcept(
	ctx context.Context,
	userID, sessionID uuid.UUID,
) error {
	err := r.q.DeleteSessionsByUserExcept(ctx, db.DeleteSessionsByUserExceptParams{
		UserID: toUUID(userID),
		ID:     toUUID(sessionID),
	})
	if err != nil {
		return repoError(err)
	}
	return nil
}

func toSession(row db.Session, token string) entity.Session {
	return entity.Session{
		ID:        fromUUID(row.ID),
		UserID:    fromUUID(row.UserID),
		Token:     token,
		ExpiresAt: fromTimestamptz(row.ExpiresAt),
		CreatedAt: fromTimestamptz(row.CreatedAt),
	}
}
