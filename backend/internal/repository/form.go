package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type FormRepository interface {
	Create(ctx context.Context, form entity.Form) (entity.Form, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Form, error)
	UpdateTitleQuestion(ctx context.Context, id uuid.UUID, titleQuestionID *string) error
	UpdateSyncedAt(ctx context.Context, id uuid.UUID, syncedAt time.Time) error

	ListQuestions(ctx context.Context, formID uuid.UUID) ([]entity.FormQuestion, error)
	UpsertQuestion(ctx context.Context, formID uuid.UUID, question entity.FormQuestion) error
}
