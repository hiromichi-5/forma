package repository

import (
	"context"
	"time"
)

type GoogleForm struct {
	FormID      string
	Title       string
	Description string
	Items       []GoogleFormItem
}

type GoogleFormItem struct {
	Title     string
	Questions []GoogleFormQuestion
}

type GoogleFormQuestion struct {
	QuestionID   string
	QuestionType string
	Choices      []string
}

type GoogleFormResponse struct {
	ResponseID      string
	RespondentEmail string
	SubmittedAt     time.Time
	AnswersJSON     []byte
}

type GoogleFormResponsePage struct {
	Responses     []GoogleFormResponse
	NextPageToken string
}

type FormFetcher interface {
	GetForm(ctx context.Context, formID string) (*GoogleForm, error)
	ListResponses(
		ctx context.Context,
		formID, filter, pageToken string,
	) (*GoogleFormResponsePage, error)
}
