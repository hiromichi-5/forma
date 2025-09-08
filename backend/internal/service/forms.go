package service

import (
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

func (s *Service) ListForms(ctx context.Context) ([]dbFormLite, error) {
	fs, err := s.Q.ListForms(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dbFormLite, 0, len(fs))
	for _, f := range fs {
		out = append(out, dbFormLite{FormID: f.FormID, Title: f.Title})
	}
	return out, nil
}

type dbFormLite struct{ FormID, Title string }

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
