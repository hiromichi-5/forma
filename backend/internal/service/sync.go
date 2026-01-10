package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/forms/v1"

	"github.com/hiromichi-5/forma/backend/internal/db"
)

func (s *Service) SyncFormOnce(ctx context.Context, formID string, actor uuid.UUID) (synced int, newTickets int, last time.Time, err error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return 0, 0, time.Time{}, err
	}

	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return 0, 0, time.Time{}, ErrValidation
	}

	form, err := s.Q.GetFormByID(ctx, dbUUID(formUUID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, time.Time{}, ErrFormsNotFound
		}
		return 0, 0, time.Time{}, err
	}

	if err := s.refreshFormQuestions(ctx, form.ID, form.FormID); err != nil {
		return 0, 0, time.Time{}, err
	}

	filter := ""
	if form.SyncedAt.Valid {
		formatted := form.SyncedAt.Time.UTC().Format(time.RFC3339)
		if _, err := time.Parse(time.RFC3339, formatted); err != nil {
			return 0, 0, time.Time{}, ErrValidation
		}
		filter = "timestamp >= " + formatted
	}

	var all []*forms.FormResponse
	token := ""
	for {
		page, e := s.GF.ListResponses(ctx, form.FormID, filter, token)
		if e != nil {
			return 0, 0, time.Time{}, e
		}
		if page.Responses != nil {
			all = append(all, page.Responses...)
		}
		if page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}

	type pair struct {
		submitted time.Time
		r         *forms.FormResponse
	}
	ps := make([]pair, 0, len(all))
	for _, r := range all {
		if r == nil || (r.CreateTime == "" && r.LastSubmittedTime == "") {
			continue
		}
		ts := r.LastSubmittedTime
		if ts == "" {
			ts = r.CreateTime
		}
		tm, e := time.Parse(time.RFC3339, ts)
		if e != nil {
			continue
		}
		ps = append(ps, pair{submitted: tm, r: r})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].submitted.Before(ps[j].submitted) })

	defaultStatus, err := s.Q.GetDefaultFormStatus(ctx, form.ID)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	var maxSubmitted time.Time
	for _, p := range ps {
		if p.submitted.After(maxSubmitted) {
			maxSubmitted = p.submitted
		}

		rid := p.r.ResponseId
		if rid == "" {
			continue
		}

		answersBytes, err := json.Marshal(p.r.Answers)
		if err != nil {
			continue
		}

		respondent := pgtype.Text{Valid: false}
		if p.r.RespondentEmail != "" {
			respondent = pgtype.Text{String: p.r.RespondentEmail, Valid: true}
		}

		rowsAffected, e := s.Q.CreateTicket(ctx, db.CreateTicketParams{
			ID:              dbUUID(uuid.New()),
			FormID:          form.ID,
			ResponseID:      rid,
			RespondentEmail: respondent,
			Answers:         answersBytes,
			StatusID:        defaultStatus.ID,
			AssigneeID:      pgtype.UUID{Valid: false},
			Priority:        "medium",
			SubmittedAt:     pgtype.Timestamptz{Time: p.submitted, Valid: true},
		})
		if e != nil {
			continue
		}
		if rowsAffected > 0 {
			synced++
			newTickets++
		}
	}

	if !maxSubmitted.IsZero() {
		_ = s.Q.UpdateFormSyncedAt(ctx, db.UpdateFormSyncedAtParams{
			ID:       form.ID,
			SyncedAt: pgtype.Timestamptz{Time: maxSubmitted, Valid: true},
		})
	}

	return synced, newTickets, maxSubmitted, nil
}

func (s *Service) refreshFormQuestions(ctx context.Context, formID pgtype.UUID, googleFormID string) error {
	form, err := s.GF.GetForm(ctx, googleFormID)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "403") {
			return ErrFormsNotShared
		}
		if strings.Contains(lower, "404") {
			return ErrFormsNotFound
		}
		return err
	}
	if form == nil {
		return ErrFormsNotFound
	}

	for _, item := range form.Items {
		if item == nil {
			continue
		}
		questions := extractQuestions(item)
		for _, q := range questions {
			if q == nil {
				continue
			}

			qid := q.QuestionId
			if qid == "" {
				continue
			}

			title := item.Title
			if title == "" {
				title = qid
			}

			optBytes := marshalQuestionOptions(q)

			if err := s.Q.UpsertFormQuestion(ctx, db.UpsertFormQuestionParams{
				FormID:       formID,
				QuestionID:   qid,
				Title:        title,
				QuestionType: detectQuestionType(q),
				Options:      optBytes,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractQuestions(item *forms.Item) []*forms.Question {
	if item == nil {
		return nil
	}
	if item.QuestionItem != nil && item.QuestionItem.Question != nil {
		return []*forms.Question{item.QuestionItem.Question}
	}
	if item.QuestionGroupItem != nil && len(item.QuestionGroupItem.Questions) > 0 {
		return item.QuestionGroupItem.Questions
	}
	return nil
}

func marshalQuestionOptions(q *forms.Question) []byte {
	if q == nil || q.ChoiceQuestion == nil {
		return nil
	}
	choices := make([]string, 0, len(q.ChoiceQuestion.Options))
	for _, opt := range q.ChoiceQuestion.Options {
		if opt == nil {
			continue
		}
		if strings.TrimSpace(opt.Value) != "" {
			choices = append(choices, opt.Value)
		}
	}
	if len(choices) == 0 {
		return nil
	}
	payload := map[string]any{"choices": choices}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return b
}

func detectQuestionType(q *forms.Question) string {
	if q == nil {
		return "unknown"
	}
	if q.TextQuestion != nil {
		if q.TextQuestion.Paragraph {
			return "paragraph"
		}
		return "text"
	}
	if q.ChoiceQuestion != nil {
		switch strings.ToUpper(q.ChoiceQuestion.Type) {
		case "RADIO":
			return "radio"
		case "CHECKBOX":
			return "checkbox"
		case "DROP_DOWN":
			return "drop_down"
		default:
			return "choice"
		}
	}
	if q.ScaleQuestion != nil {
		return "scale"
	}
	if q.DateQuestion != nil {
		return "date"
	}
	if q.TimeQuestion != nil {
		return "time"
	}
	if q.FileUploadQuestion != nil {
		return "file"
	}
	if q.RowQuestion != nil {
		return "row"
	}
	if q.RatingQuestion != nil {
		return "rating"
	}
	return "unknown"
}
