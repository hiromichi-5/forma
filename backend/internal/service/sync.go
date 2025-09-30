package service

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/forms/v1"

	"github.com/hiromichi-5/forma/backend/internal/db"
)

func (s *Service) roleFor(ctx context.Context, formID string, actor uuid.UUID) (string, error) {
	r, err := s.Q.GetUserFormRole(ctx, db.GetUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: actor, Valid: true},
		FormID: formID,
	})
	if err != nil {
		return "", ErrForbidden
	}
	return r, nil
}

func (s *Service) SyncFormOnce(ctx context.Context, formID string, actor uuid.UUID) (synced int, newTickets int, last time.Time, err error) {
	role, err := s.roleFor(ctx, formID, actor)
	if err != nil || (role != "admin" && role != "editor") {
		return 0, 0, time.Time{}, ErrForbidden
	}

	if err := s.refreshFormQuestions(ctx, formID); err != nil {
		return 0, 0, time.Time{}, err
	}

	// カーソル決定 - 既存のsync_cursorを取得、なければ7日前を使用
	var cursor time.Time
	syncCursor, err := s.Q.GetFormSyncCursor(ctx, formID)
	if err != nil || !syncCursor.Valid {
		cursor = time.Now().Add(-7 * 24 * time.Hour)
	} else {
		cursor = syncCursor.Time
	}

	formattedCursor := cursor.UTC().Format(time.RFC3339)
	// Validate the formatted timestamp to ensure it matches RFC3339
	if _, err := time.Parse(time.RFC3339, formattedCursor); err != nil {
		return 0, 0, time.Time{}, ErrForbidden
	}
	filter := "timestamp >= " + formattedCursor
	var all []*forms.FormResponse
	token := ""
	for {
		page, e := s.GF.ListResponses(ctx, formID, filter, token)
		if e != nil {
			err = e
			return
		}
		if page.Responses != nil {
			all = append(all, page.Responses...)
		}
		if page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}

	// submittedTime昇順で処理
	type pair struct {
		submitted time.Time
		r         *forms.FormResponse
	}
	ps := make([]pair, 0, len(all))
	for _, r := range all {
		if r.CreateTime == "" && r.LastSubmittedTime == "" {
			continue
		}
		// submitted_atはLastSubmittedTime優先
		ts := r.LastSubmittedTime
		if ts == "" {
			ts = r.CreateTime
		}
		t, e := time.Parse(time.RFC3339, ts)
		if e != nil {
			continue
		}
		ps = append(ps, pair{submitted: t, r: r})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].submitted.Before(ps[j].submitted) })

	var maxSubmitted time.Time
	for _, p := range ps {
		if p.submitted.After(maxSubmitted) {
			maxSubmitted = p.submitted
		}

		// 代表キー: responseId
		rid := p.r.ResponseId
		if rid == "" {
			continue
		}

		// payloadはそのまま保持
		payloadData := map[string]any{"answers": p.r.Answers}
		payloadBytes, err := json.Marshal(payloadData)
		if err != nil {
			continue
		}

		// responses挿入（重複は無視）
		rowsAffected, e := s.Q.InsertResponse(ctx, db.InsertResponseParams{
			ResponseID:    rid,
			FormID:        formID,
			SubmittedAt:   pgtype.Timestamptz{Time: p.submitted, Valid: true},
			Payload:       payloadBytes,
			SchemaVersion: 1,
		})
		if e != nil {
			continue
		}

		// 新規挿入された場合のみ
		if rowsAffected > 0 {
			synced++
			// 新規のみチケット作成
			ticketID := uuid.New()
			_, e = s.Q.CreateTicket(ctx, db.CreateTicketParams{
				ID:         pgtype.UUID{Bytes: ticketID, Valid: true},
				FormID:     formID,
				ResponseID: rid,
			})
			if e == nil {
				newTickets++
			}
		}
	}

	if !maxSubmitted.IsZero() {
		_ = s.Q.UpdateSyncCursor(ctx, db.UpdateSyncCursorParams{
			FormID:     formID,
			SyncCursor: pgtype.Timestamptz{Time: maxSubmitted, Valid: true},
		})
	}

	return synced, newTickets, maxSubmitted, nil
}

func (s *Service) refreshFormQuestions(ctx context.Context, formID string) error {
	form, err := s.GF.GetForm(ctx, formID)
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

	var defaultCandidate string
	var firstQuestion string

	for _, item := range form.Items {
		if item == nil {
			continue
		}
		questions := extractQuestions(item)
		for _, q := range questions {
			if q.Question == nil {
				continue
			}

			qid := q.Question.QuestionId
			if qid == "" {
				continue
			}

			title := q.Title
			if title == "" {
				title = qid
			}

			optBytes := marshalQuestionOptions(q.Question)

			if err := s.Q.UpsertFormQuestion(ctx, db.UpsertFormQuestionParams{
				FormID:       formID,
				QuestionID:   qid,
				Title:        title,
				QuestionType: detectQuestionType(q.Question),
				Options:      optBytes,
			}); err != nil {
				return err
			}

			if firstQuestion == "" {
				firstQuestion = qid
			}
			if defaultCandidate == "" && isGoodTitleQuestion(q.Question) {
				defaultCandidate = qid
			}
		}
	}

	if defaultCandidate == "" {
		defaultCandidate = firstQuestion
	}

	if defaultCandidate == "" {
		return nil
	}

	current, err := s.Q.GetFormTitleQuestion(ctx, formID)
	if err != nil {
		return err
	}
	if current.Valid && current.String != "" {
		return nil
	}

	return s.Q.UpdateFormTitleQuestion(ctx, db.UpdateFormTitleQuestionParams{
		FormID: formID,
		TitleQuestionID: pgtype.Text{
			String: defaultCandidate,
			Valid:  true,
		},
	})
}

type questionWithTitle struct {
	Title    string
	Question *forms.Question
}

func extractQuestions(item *forms.Item) []questionWithTitle {
	var out []questionWithTitle
	if item.QuestionItem != nil && item.QuestionItem.Question != nil {
		title := item.Title
		if title == "" {
			title = item.QuestionItem.Question.QuestionId
		}
		out = append(out, questionWithTitle{Title: title, Question: item.QuestionItem.Question})
	}

	if item.QuestionGroupItem != nil {
		for _, q := range item.QuestionGroupItem.Questions {
			if q == nil {
				continue
			}
			title := item.Title
			if title == "" {
				title = q.QuestionId
			}
			out = append(out, questionWithTitle{Title: title, Question: q})
		}
	}

	return out
}

func detectQuestionType(q *forms.Question) string {
	if q == nil {
		return "unknown"
	}
	switch {
	case q.TextQuestion != nil:
		if q.TextQuestion.Paragraph {
			return "paragraph"
		}
		return "text"
	case q.ChoiceQuestion != nil:
		if q.ChoiceQuestion.Type != "" {
			return strings.ToLower(q.ChoiceQuestion.Type)
		}
		return "choice"
	case q.ScaleQuestion != nil:
		return "scale"
	case q.RatingQuestion != nil:
		return "rating"
	case q.FileUploadQuestion != nil:
		return "file_upload"
	case q.RowQuestion != nil:
		return "row"
	case q.TimeQuestion != nil:
		return "time"
	case q.DateQuestion != nil:
		return "date"
	default:
		return "unknown"
	}
}

func marshalQuestionOptions(q *forms.Question) []byte {
	if q == nil {
		return nil
	}
	payload := map[string]any{}
	if cq := q.ChoiceQuestion; cq != nil {
		values := make([]string, 0, len(cq.Options))
		for _, opt := range cq.Options {
			if opt != nil && opt.Value != "" {
				values = append(values, opt.Value)
			}
		}
		if len(values) > 0 {
			payload["choices"] = values
		}
		if cq.Shuffle {
			payload["shuffle"] = true
		}
	}
	if sq := q.ScaleQuestion; sq != nil {
		payload["low"] = sq.Low
		payload["high"] = sq.High
		if sq.LowLabel != "" {
			payload["low_label"] = sq.LowLabel
		}
		if sq.HighLabel != "" {
			payload["high_label"] = sq.HighLabel
		}
	}
	if len(payload) == 0 {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return b
}

func isGoodTitleQuestion(q *forms.Question) bool {
	if q == nil {
		return false
	}
	if q.TextQuestion != nil {
		return true
	}
	if q.ChoiceQuestion != nil {
		return true
	}
	return false
}
