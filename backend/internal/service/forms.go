package service

import (
	"encoding/json"
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrFormsNotShared = errors.New("forms not shared")
	ErrFormsNotFound  = errors.New("forms not found")
)

var reFormID = regexp.MustCompile(`/forms/d/e/([a-zA-Z0-9_-]+)/`)

type FormQuestion struct {
	FormID       string   `json:"form_id"`
	QuestionID   string   `json:"question_id"`
	Title        string   `json:"title"`
	QuestionType string   `json:"question_type"`
	Options      []string `json:"options,omitempty"`
}

func extractFormID(u string) (string, error) {
	// フルURLから/forms/d/e/{ID}/を抜く
	if m := reFormID.FindStringSubmatch(u); len(m) == 2 {
		return m[1], nil
	}
	// IDだけが来た場合は受け入れる
	if len(u) >= 20 && !strings.Contains(u, "/") {
		return u, nil
	}
	// URLとして妥当でないものは弾く
	if _, err := url.ParseRequestURI(u); err != nil {
		return "", ErrValidation
	}
	return "", ErrValidation
}

type UserID = uuid.UUID

func (s *Service) RegisterForm(ctx context.Context, formURL string, pollingSec int, creator UserID) (string, error) {
	formID, err := extractFormID(formURL)
	if err != nil {
		return "", ErrValidation
	}

	f, err := s.GF.GetForm(ctx, formID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "403") {
			return "", ErrFormsNotShared
		}
		return "", ErrFormsNotFound
	}

	title := ""
	if f != nil && f.Info != nil && f.Info.Title != "" {
		title = f.Info.Title
	}

	if err := s.Q.UpsertForm(ctx, db.UpsertFormParams{
		FormID:      formID,
		Title:       title,
		Description: pgtype.Text{Valid: false},
		PollingSec:  pgtype.Int4{Int32: int32(pollingSec), Valid: true},
	}); err != nil {
		return "", err
	}

	if err := s.Q.UpsertUserFormRole(ctx, db.UpsertUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: creator, Valid: true},
		FormID: formID,
		Role:   "admin",
	}); err != nil {
		return "", err
	}
	return formID, nil
}

func (s *Service) ListForms(ctx context.Context, actor UserID) ([]dbFormLite, error) {
	fs, err := s.Q.ListUserAccessibleForms(ctx, pgtype.UUID{Bytes: actor, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]dbFormLite, 0, len(fs))
	for _, f := range fs {
		out = append(out, dbFormLite{FormId: f.FormID, Title: f.Title})
	}
	return out, nil
}

type dbFormLite struct {
	FormId string `json:"form_id"`
	Title  string `json:"title"`
}

func (s *Service) Health(ctx context.Context, formID string, actor UserID) (map[string]any, error) {
	role, err := s.Q.GetUserFormRole(ctx, db.GetUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: actor, Valid: true},
		FormID: formID,
	})
	if err != nil {
		return nil, ErrForbidden
	}
	if role != "admin" && role != "editor" {
		return nil, ErrForbidden
	}

	f, err := s.GF.GetForm(ctx, formID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "403") {
			return nil, ErrFormsNotShared
		}
		return nil, ErrFormsNotFound
	}

	title := ""
	if f != nil && f.Info != nil {
		title = f.Info.Title
	}

	return map[string]any{"form_id": formID, "title": title}, nil
}

func (s *Service) ListFormQuestions(ctx context.Context, formID string, actor uuid.UUID) ([]FormQuestion, error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return nil, err
	}
	rows, err := s.Q.ListFormQuestions(ctx, formID)
	if err != nil {
		return nil, err
	}
	out := make([]FormQuestion, 0, len(rows))
	for _, row := range rows {
		fq := FormQuestion{
			FormID:       row.FormID,
			QuestionID:   row.QuestionID,
			Title:        row.Title,
			QuestionType: row.QuestionType,
		}
		if len(row.Options) > 0 {
			var options map[string]any
			if err := json.Unmarshal(row.Options, &options); err == nil {
				if rawChoices, ok := options["choices"].([]any); ok {
					choices := make([]string, 0, len(rawChoices))
					for _, v := range rawChoices {
						if str, ok := v.(string); ok {
							choices = append(choices, str)
						}
					}
					fq.Options = choices
				}
			}
		}
		out = append(out, fq)
	}
	return out, nil
}

func (s *Service) SetFormTitleQuestion(ctx context.Context, formID string, questionID *string, actor uuid.UUID) error {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return err
	}
	var value pgtype.Text
	if questionID != nil && *questionID != "" {
		questions, err := s.Q.ListFormQuestions(ctx, formID)
		if err != nil {
			return err
		}
		found := false
		for _, q := range questions {
			if q.QuestionID == *questionID {
				found = true
				break
			}
		}
		if !found {
			return ErrValidation
		}
		value = pgtype.Text{String: *questionID, Valid: true}
	} else {
		value = pgtype.Text{Valid: false}
	}
	return s.Q.UpdateFormTitleQuestion(ctx, db.UpdateFormTitleQuestionParams{
		FormID:          formID,
		TitleQuestionID: value,
	})
}
