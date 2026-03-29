package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.FormRepository = (*FormRepository)(nil)

type FormRepository struct {
	q *db.Queries
}

func NewFormRepository(pool *pgxpool.Pool) *FormRepository {
	return &FormRepository{q: db.New(pool)}
}

func (r *FormRepository) Create(ctx context.Context, form entity.Form) (entity.Form, error) {
	row, err := r.q.CreateForm(ctx, db.CreateFormParams{
		ID:                  toUUID(form.ID),
		FormID:              form.FormID,
		Title:               form.Title,
		Description:         toTextPtr(form.Description),
		TitleQuestionID:     toTextPtr(form.TitleQuestionID),
		EmailCollectionType: toTextPtr(form.EmailCollectionType),
		SyncedAt:            toTimestamptzPtr(form.SyncedAt),
	})
	if err != nil {
		return entity.Form{}, repoError(err)
	}
	return toForm(row), nil
}

func (r *FormRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Form, error) {
	row, err := r.q.GetFormByID(ctx, toUUID(id))
	if err != nil {
		return entity.Form{}, repoError(err)
	}
	return toForm(row), nil
}

func (r *FormRepository) UpdateTitleQuestion(
	ctx context.Context,
	id uuid.UUID,
	titleQuestionID *string,
) error {
	return r.q.UpdateFormTitleQuestion(ctx, db.UpdateFormTitleQuestionParams{
		ID:              toUUID(id),
		TitleQuestionID: toTextPtr(titleQuestionID),
	})
}

func (r *FormRepository) UpdateSyncedAt(
	ctx context.Context,
	id uuid.UUID,
	syncedAt time.Time,
) error {
	return r.q.UpdateFormSyncedAt(ctx, db.UpdateFormSyncedAtParams{
		ID:       toUUID(id),
		SyncedAt: toTimestamptz(syncedAt),
	})
}

func (r *FormRepository) ListQuestions(
	ctx context.Context,
	formID uuid.UUID,
) ([]entity.FormQuestion, error) {
	rows, err := r.q.ListFormQuestions(ctx, toUUID(formID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.FormQuestion, len(rows))
	for i, row := range rows {
		result[i] = entity.FormQuestion{
			QuestionID:   row.QuestionID,
			Title:        row.Title,
			QuestionType: row.QuestionType,
			Options:      unmarshalOptions(row.Options),
		}
	}
	return result, nil
}

func (r *FormRepository) UpsertQuestion(
	ctx context.Context,
	formID uuid.UUID,
	question entity.FormQuestion,
) error {
	return r.q.UpsertFormQuestion(ctx, db.UpsertFormQuestionParams{
		FormID:       toUUID(formID),
		QuestionID:   question.QuestionID,
		Title:        question.Title,
		QuestionType: question.QuestionType,
		Options:      marshalOptions(question.Options),
	})
}

func toForm(row db.Form) entity.Form {
	return entity.Form{
		ID:                  fromUUID(row.ID),
		FormID:              row.FormID,
		Title:               row.Title,
		Description:         fromTextPtr(row.Description),
		TitleQuestionID:     fromTextPtr(row.TitleQuestionID),
		EmailCollectionType: fromTextPtr(row.EmailCollectionType),
		SyncedAt:            fromTimestamptzPtr(row.SyncedAt),
		CreatedAt:           fromTimestamptz(row.CreatedAt),
	}
}

func marshalOptions(opts []string) []byte {
	if len(opts) == 0 {
		return nil
	}
	b, err := json.Marshal(map[string]any{"choices": opts})
	if err != nil {
		return nil
	}
	return b
}

func unmarshalOptions(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	var v struct {
		Choices []string `json:"choices"`
	}
	if json.Unmarshal(data, &v) != nil {
		return nil
	}
	return v.Choices
}
