package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.StatusRepository = (*StatusRepository)(nil)

type StatusRepository struct {
	q *db.Queries
}

func NewStatusRepository(pool *pgxpool.Pool) *StatusRepository {
	return &StatusRepository{q: db.New(pool)}
}

func (r *StatusRepository) Create(
	ctx context.Context,
	status entity.FormStatus,
) (entity.FormStatus, error) {
	row, err := r.q.CreateFormStatus(ctx, db.CreateFormStatusParams{
		ID:           toUUID(status.ID),
		FormID:       toUUID(status.FormID),
		Name:         status.Name,
		Color:        toTextPtr(status.Color),
		DisplayOrder: status.DisplayOrder,
		IsDefault:    status.IsDefault,
	})
	if err != nil {
		return entity.FormStatus{}, repoError(err)
	}
	return toFormStatus(row), nil
}

func (r *StatusRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.FormStatus, error) {
	row, err := r.q.GetFormStatusByID(ctx, toUUID(id))
	if err != nil {
		return entity.FormStatus{}, repoError(err)
	}
	return toFormStatus(row), nil
}

func (r *StatusRepository) GetDefault(
	ctx context.Context,
	formID uuid.UUID,
) (entity.FormStatus, error) {
	row, err := r.q.GetDefaultFormStatus(ctx, toUUID(formID))
	if err != nil {
		return entity.FormStatus{}, repoError(err)
	}
	return toFormStatus(row), nil
}

func (r *StatusRepository) List(
	ctx context.Context,
	formID uuid.UUID,
) ([]entity.FormStatus, error) {
	rows, err := r.q.ListFormStatuses(ctx, toUUID(formID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.FormStatus, len(rows))
	for i, row := range rows {
		result[i] = toFormStatus(row)
	}
	return result, nil
}

func (r *StatusRepository) Update(
	ctx context.Context,
	status entity.FormStatus,
) (entity.FormStatus, error) {
	row, err := r.q.UpdateFormStatus(ctx, db.UpdateFormStatusParams{
		ID:           toUUID(status.ID),
		Name:         status.Name,
		Color:        toTextPtr(status.Color),
		DisplayOrder: status.DisplayOrder,
	})
	if err != nil {
		return entity.FormStatus{}, repoError(err)
	}
	return toFormStatus(row), nil
}

func (r *StatusRepository) SetDefault(
	ctx context.Context,
	formID uuid.UUID,
	statusID uuid.UUID,
) (entity.FormStatus, error) {
	row, err := r.q.SetDefaultFormStatus(ctx, db.SetDefaultFormStatusParams{
		FormID: toUUID(formID),
		ID:     toUUID(statusID),
	})
	if err != nil {
		return entity.FormStatus{}, repoError(err)
	}
	return toFormStatus(row), nil
}

func (r *StatusRepository) ClearDefault(ctx context.Context, formID uuid.UUID) error {
	return r.q.ClearDefaultFormStatus(ctx, toUUID(formID))
}

func (r *StatusRepository) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.DeleteFormStatus(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func toFormStatus(row db.FormStatus) entity.FormStatus {
	return entity.FormStatus{
		ID:           fromUUID(row.ID),
		FormID:       fromUUID(row.FormID),
		Name:         row.Name,
		Color:        fromTextPtr(row.Color),
		DisplayOrder: row.DisplayOrder,
		IsDefault:    row.IsDefault,
	}
}
