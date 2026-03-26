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
	ListActiveFormInvites(ctx context.Context, formID pgtype.UUID) ([]db.FormInvite, error)
	GetFormInviteForUpdate(ctx context.Context, id pgtype.UUID) (db.FormInvite, error)
	AcceptFormInvite(ctx context.Context, id pgtype.UUID) (db.FormInvite, error)
	DeleteFormInvite(ctx context.Context, id pgtype.UUID) error
}

type rolesStore interface {
	GetFormMemberRole(ctx context.Context, arg db.GetFormMemberRoleParams) (db.FormRole, error)
	UpsertFormMember(ctx context.Context, arg db.UpsertFormMemberParams) error
	DeleteFormMember(ctx context.Context, arg db.DeleteFormMemberParams) error
	ListFormMembers(ctx context.Context, formID pgtype.UUID) ([]db.ListFormMembersRow, error)
	ListUserAccessibleForms(
		ctx context.Context,
		userID pgtype.UUID,
	) ([]db.ListUserAccessibleFormsRow, error)
	CountFormAdmins(ctx context.Context, formID pgtype.UUID) (int64, error)
}

type usersStore interface {
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.GetUserByIDRow, error)
}

type Service struct {
	Q       *db.Queries
	GF      google.FormsClient
	Invites invitesStore
	Roles   rolesStore
	Users   usersStore

	now          func() time.Time
	generateCode func() (string, error)
}

func NewService(q *db.Queries, gf google.FormsClient) *Service {
	s := &Service{
		Q:            q,
		GF:           gf,
		Invites:      q,
		Roles:        q,
		Users:        q,
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
