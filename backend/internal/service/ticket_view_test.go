package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestParseResponseAnswers(t *testing.T) {
	wrapped := map[string]any{
		"answers": map[string]any{
			"q1": map[string]any{
				"questionId": "q1",
				"textAnswers": map[string]any{
					"answers": []map[string]any{{"value": "A"}},
				},
			},
		},
	}
	payload, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}

	answers, err := parseResponseAnswers(payload)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if got := answers["q1"]; len(got) != 1 || got[0] != "A" {
		t.Fatalf("unexpected answers: %#v", answers)
	}

	raw := map[string]any{
		"q1": map[string]any{
			"questionId": "q1",
			"textAnswers": map[string]any{
				"answers": []map[string]any{{"value": "B"}},
			},
		},
	}
	rawPayload, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	answers, err = parseResponseAnswers(rawPayload)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if got := answers["q1"]; len(got) != 1 || got[0] != "B" {
		t.Fatalf("unexpected answers: %#v", answers)
	}

	answers, err = parseResponseAnswers(nil)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if len(answers) != 0 {
		t.Fatalf("want empty answers, got %#v", answers)
	}
}

func TestBuildTicketSummary_TitleResolution(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	ticketID := uuid.MustParse("00000000-0000-0000-0000-000000000020")
	statusID := uuid.MustParse("00000000-0000-0000-0000-000000000030")
	questions := newFormQuestionSet([]db.FormQuestion{{
		QuestionID:   "q1",
		Title:        "Title",
		QuestionType: "text",
	}})

	row := db.ListTicketsRow{
		ID:              pgtype.UUID{Bytes: ticketID, Valid: true},
		FormID:          pgtype.UUID{Bytes: formID, Valid: true},
		ResponseID:      "resp-1",
		StatusID:        pgtype.UUID{Bytes: statusID, Valid: true},
		Priority:        "medium",
		SubmittedAt:     pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		FormTitle:       "Form Title",
		TitleQuestionID: pgtype.Text{String: "q1", Valid: true},
		StatusName:      "未対応",
	}

	answers := map[string][]string{"q1": {"Answer Title"}}
	summary := buildTicketSummary(row, answers, questions)
	if summary.Title != "Answer Title" {
		t.Fatalf("want title from answer, got %s", summary.Title)
	}
	if summary.TitleQuestionID == nil || *summary.TitleQuestionID != "q1" {
		t.Fatalf("unexpected title_question_id: %#v", summary.TitleQuestionID)
	}

	row.TitleQuestionID = pgtype.Text{}
	answers = map[string][]string{}
	summary = buildTicketSummary(row, answers, questions)
	if summary.Title != "Form Title" {
		t.Fatalf("want title from form title, got %s", summary.Title)
	}
	if summary.TitleQuestionID == nil || *summary.TitleQuestionID != "q1" {
		t.Fatalf("unexpected title_question_id: %#v", summary.TitleQuestionID)
	}
}
