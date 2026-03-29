package entity

import (
	"time"

	"github.com/google/uuid"
)

type Form struct {
	ID                  uuid.UUID
	FormID              string
	Title               string
	Description         *string
	TitleQuestionID     *string
	EmailCollectionType *string
	SyncedAt            *time.Time
	CreatedAt           time.Time
}

type FormQuestion struct {
	QuestionID   string
	Title        string
	QuestionType string
	Options      []string
}

type FormStatus struct {
	ID           uuid.UUID
	FormID       uuid.UUID
	Name         string
	Color        *string
	DisplayOrder int32
	IsDefault    bool
}
