package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrFormsNotShared = errors.New("forms not shared")
	ErrFormsNotFound  = errors.New("forms not found")
)

var reFormID = regexp.MustCompile(`/forms/d/e/([a-zA-Z0-9_-]+)/`)

type FormQuestion struct {
	QuestionID   string   `json:"question_id"`
	Title        string   `json:"title"`
	QuestionType string   `json:"question_type"`
	Options      []string `json:"options,omitempty"`
}

type FormSummary struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	SyncedAt *time.Time `json:"synced_at"`
}

type FormDetail struct {
	ID                  string     `json:"id"`
	Title               string     `json:"title"`
	Description         *string    `json:"description"`
	TitleQuestionID     *string    `json:"title_question_id"`
	EmailCollectionType *string    `json:"email_collection_type"`
	SyncedAt            *time.Time `json:"synced_at"`
	CreatedAt           time.Time  `json:"created_at"`
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

func (s *Service) RegisterForm(ctx context.Context, formURL string, creator uuid.UUID) (uuid.UUID, error) {
	formID, err := extractFormID(formURL)
	if err != nil {
		return uuid.UUID{}, ErrValidation
	}

	f, err := s.GF.GetForm(ctx, formID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "403") {
			return uuid.UUID{}, ErrFormsNotShared
		}
		return uuid.UUID{}, ErrFormsNotFound
	}

	if f == nil || f.Info == nil {
		return uuid.UUID{}, ErrFormsNotFound
	}

	title := strings.TrimSpace(f.Info.Title)
	if title == "" {
		title = formID
	}

	var description pgtype.Text
	if f.Info.Description != "" {
		description = pgtype.Text{String: f.Info.Description, Valid: true}
	}

	newID := uuid.New()
	form, err := s.Q.CreateForm(ctx, db.CreateFormParams{
		ID:                  dbUUID(newID),
		FormID:              formID,
		Title:               title,
		Description:         description,
		TitleQuestionID:     pgtype.Text{Valid: false},
		EmailCollectionType: pgtype.Text{Valid: false},
		SyncedAt:            pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if err := s.Q.UpsertFormMember(ctx, db.UpsertFormMemberParams{
		UserID: dbUUID(creator),
		FormID: form.ID,
		Role:   "admin",
	}); err != nil {
		return uuid.UUID{}, err
	}

	if err := s.initFormStatuses(ctx, form.ID); err != nil {
		return uuid.UUID{}, err
	}

	return newID, nil
}

func (s *Service) initFormStatuses(ctx context.Context, formID pgtype.UUID) error {
	statuses := []struct {
		name      string
		order     int32
		isDefault bool
	}{
		{name: "未対応", order: 1, isDefault: true},
		{name: "対応中", order: 2, isDefault: false},
		{name: "対応完了", order: 3, isDefault: false},
	}
	for _, st := range statuses {
		if _, err := s.Q.CreateFormStatus(ctx, db.CreateFormStatusParams{
			ID:           dbUUID(uuid.New()),
			FormID:       formID,
			Name:         st.name,
			Color:        pgtype.Text{Valid: false},
			DisplayOrder: st.order,
			IsDefault:    st.isDefault,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListForms(ctx context.Context, actor uuid.UUID) ([]FormSummary, error) {
	fs, err := s.Q.ListUserAccessibleForms(ctx, dbUUID(actor))
	if err != nil {
		return nil, err
	}
	out := make([]FormSummary, 0, len(fs))
	for _, f := range fs {
		out = append(out, FormSummary{
			ID:       uuid.UUID(f.ID.Bytes).String(),
			Title:    f.Title,
			SyncedAt: timestamptzPtr(f.SyncedAt),
		})
	}
	return out, nil
}

func (s *Service) GetForm(ctx context.Context, formID string, actor uuid.UUID) (FormDetail, error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return FormDetail{}, err
	}
	uid, err := uuid.Parse(formID)
	if err != nil {
		return FormDetail{}, ErrValidation
	}
	f, err := s.Q.GetFormByID(ctx, dbUUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FormDetail{}, ErrFormsNotFound
		}
		return FormDetail{}, err
	}
	return FormDetail{
		ID:                  uuid.UUID(f.ID.Bytes).String(),
		Title:               f.Title,
		Description:         textPtr(f.Description),
		TitleQuestionID:     textPtr(f.TitleQuestionID),
		EmailCollectionType: textPtr(f.EmailCollectionType),
		SyncedAt:            timestamptzPtr(f.SyncedAt),
		CreatedAt:           f.CreatedAt.Time,
	}, nil
}

func (s *Service) UpdateFormTitleQuestion(ctx context.Context, formID string, titleQuestionID *string, actor uuid.UUID) error {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return err
	}
	uid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}

	var questionID pgtype.Text
	if titleQuestionID != nil && strings.TrimSpace(*titleQuestionID) != "" {
		questions, err := s.Q.ListFormQuestions(ctx, dbUUID(uid))
		if err != nil {
			return err
		}
		found := false
		for _, q := range questions {
			if q.QuestionID == *titleQuestionID {
				found = true
				break
			}
		}
		if !found {
			return ErrValidation
		}
		questionID = pgtype.Text{String: *titleQuestionID, Valid: true}
	}

	return s.Q.UpdateFormTitleQuestion(ctx, db.UpdateFormTitleQuestionParams{
		ID:              dbUUID(uid),
		TitleQuestionID: questionID,
	})
}

func (s *Service) ListFormQuestions(ctx context.Context, formID string, actor uuid.UUID) ([]FormQuestion, error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(formID)
	if err != nil {
		return nil, ErrValidation
	}
	rows, err := s.Q.ListFormQuestions(ctx, dbUUID(uid))
	if err != nil {
		return nil, err
	}
	out := make([]FormQuestion, 0, len(rows))
	for _, row := range rows {
		fq := FormQuestion{
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

func timestamptzPtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	v := ts.Time
	return &v
}

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	value := t.String
	return &value
}
