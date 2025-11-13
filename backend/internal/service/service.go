package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"strings"
	"time"

	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/google"
	"github.com/jackc/pgx/v5/pgtype"
)

type invitesStore interface {
	CreateFormInvite(ctx context.Context, arg db.CreateFormInviteParams) (db.FormInvite, error)
	ListActiveFormInvites(ctx context.Context, arg db.ListActiveFormInvitesParams) ([]db.FormInvite, error)
	GetFormInviteForUpdate(ctx context.Context, code string) (db.FormInvite, error)
	RevokeFormInvite(ctx context.Context, arg db.RevokeFormInviteParams) (db.FormInvite, error)
}

type rolesStore interface {
	GetUserFormRole(ctx context.Context, arg db.GetUserFormRoleParams) (string, error)
	UpsertUserFormRole(ctx context.Context, arg db.UpsertUserFormRoleParams) error
	DeleteUserFormRole(ctx context.Context, arg db.DeleteUserFormRoleParams) error
	ListFormMembers(ctx context.Context, formID string) ([]db.ListFormMembersRow, error)
	ListUserAccessibleForms(ctx context.Context, userID pgtype.UUID) ([]db.ListUserAccessibleFormsRow, error)
	CountFormAdmins(ctx context.Context, formID string) (int64, error)
}

type Service struct {
	Q       *db.Queries
	GF      google.FormsClient
	Invites invitesStore
	Roles   rolesStore

	now          func() time.Time
	generateCode func() (string, error)
}

func NewService(q *db.Queries, gf google.FormsClient) *Service {
	s := &Service{
		Q:            q,
		GF:           gf,
		Invites:      q,
		Roles:        q,
		now:          time.Now,
		generateCode: defaultInviteCode,
	}
	return s
}

func defaultInviteCode() (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	code = strings.ToLower(code)
	if len(code) > 10 {
		code = code[:10]
	}
	return code, nil
}

func (s *Service) nowTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *Service) newInviteCode() (string, error) {
	if s.generateCode != nil {
		return s.generateCode()
	}
	return defaultInviteCode()
}
