package entity

import (
	"time"

	"github.com/google/uuid"
)

// EmailCollectionType は Google Forms のメールアドレス収集設定。
type EmailCollectionType string

const (
	EmailCollectionUnspecified    EmailCollectionType = "EMAIL_COLLECTION_TYPE_UNSPECIFIED"
	EmailCollectionDoNotCollect   EmailCollectionType = "DO_NOT_COLLECT"
	EmailCollectionVerified       EmailCollectionType = "VERIFIED"
	EmailCollectionResponderInput EmailCollectionType = "RESPONDER_INPUT"
)

type Form struct {
	ID                  uuid.UUID
	GoogleFormID        string
	Title               string
	Description         *string
	TitleQuestionID     *string
	EmailCollectionType *EmailCollectionType
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
