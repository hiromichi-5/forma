package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeInvitesStore struct {
	invites map[uuid.UUID]db.FormInvite
}

func (f *fakeInvitesStore) CreateFormInvite(
	_ context.Context,
	arg db.CreateFormInviteParams,
) (db.FormInvite, error) {
	if f.invites == nil {
		f.invites = map[uuid.UUID]db.FormInvite{}
	}
	inv := db.FormInvite{
		ID:        arg.ID,
		FormID:    arg.FormID,
		Email:     arg.Email,
		Role:      arg.Role,
		InvitedBy: arg.InvitedBy,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	f.invites[arg.ID.Bytes] = inv
	return inv, nil
}

func (f *fakeInvitesStore) ListActiveFormInvites(
	_ context.Context,
	formID pgtype.UUID,
) ([]db.FormInvite, error) {
	out := []db.FormInvite{}
	for _, inv := range f.invites {
		if inv.FormID == formID && !inv.AcceptedAt.Valid && inv.ExpiresAt.Valid &&
			inv.ExpiresAt.Time.After(time.Now()) {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (f *fakeInvitesStore) GetFormInviteForUpdate(
	_ context.Context,
	id pgtype.UUID,
) (db.FormInvite, error) {
	inv, ok := f.invites[id.Bytes]
	if !ok {
		return db.FormInvite{}, pgx.ErrNoRows
	}
	return inv, nil
}

func (f *fakeInvitesStore) AcceptFormInvite(
	_ context.Context,
	id pgtype.UUID,
) (db.FormInvite, error) {
	inv, ok := f.invites[id.Bytes]
	if !ok {
		return db.FormInvite{}, pgx.ErrNoRows
	}
	inv.AcceptedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	f.invites[id.Bytes] = inv
	return inv, nil
}

func (f *fakeInvitesStore) DeleteFormInvite(_ context.Context, id pgtype.UUID) error {
	delete(f.invites, id.Bytes)
	return nil
}

type fakeRolesStore struct {
	roles map[uuid.UUID]db.FormRole
}

func (f *fakeRolesStore) GetFormMemberRole(
	_ context.Context,
	arg db.GetFormMemberRoleParams,
) (db.FormRole, error) {
	role, ok := f.roles[arg.UserID.Bytes]
	if !ok {
		return "", pgx.ErrNoRows
	}
	return role, nil
}

func (f *fakeRolesStore) UpsertFormMember(_ context.Context, _ db.UpsertFormMemberParams) error {
	return nil
}

func (f *fakeRolesStore) DeleteFormMember(_ context.Context, _ db.DeleteFormMemberParams) error {
	return nil
}

func (f *fakeRolesStore) ListFormMembers(
	_ context.Context,
	_ pgtype.UUID,
) ([]db.ListFormMembersRow, error) {
	return nil, nil
}

func (f *fakeRolesStore) ListUserAccessibleForms(
	_ context.Context,
	_ pgtype.UUID,
) ([]db.ListUserAccessibleFormsRow, error) {
	return nil, nil
}

func (f *fakeRolesStore) CountFormAdmins(_ context.Context, _ pgtype.UUID) (int64, error) {
	return 1, nil
}

type fakeUsersStore struct {
	users map[uuid.UUID]db.GetUserByIDRow
}

func (f *fakeUsersStore) GetUserByID(_ context.Context, id pgtype.UUID) (db.GetUserByIDRow, error) {
	u, ok := f.users[id.Bytes]
	if !ok {
		return db.GetUserByIDRow{}, pgx.ErrNoRows
	}
	return u, nil
}

func TestCreateInvite_Success(t *testing.T) {
	actor := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	invites := &fakeInvitesStore{}
	roles := &fakeRolesStore{roles: map[uuid.UUID]db.FormRole{actor: db.FormRoleAdmin}}
	svc := &Service{Invites: invites, Roles: roles}

	inv, err := svc.CreateInvite(
		context.Background(),
		formID.String(),
		"a@example.com",
		"editor",
		actor,
	)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if inv.Email != "a@example.com" {
		t.Fatalf("unexpected email: %s", inv.Email)
	}
}

func TestListInvites_Forbidden(t *testing.T) {
	actor := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	invites := &fakeInvitesStore{}
	roles := &fakeRolesStore{roles: map[uuid.UUID]db.FormRole{actor: db.FormRoleEditor}}
	svc := &Service{Invites: invites, Roles: roles}

	_, err := svc.ListInvites(context.Background(), formID.String(), actor)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}

func TestDeleteInvite_NotFound(t *testing.T) {
	actor := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	invites := &fakeInvitesStore{}
	roles := &fakeRolesStore{roles: map[uuid.UUID]db.FormRole{actor: db.FormRoleAdmin}}
	svc := &Service{Invites: invites, Roles: roles}

	err := svc.DeleteInvite(context.Background(), formID.String(), uuid.New().String(), actor)
	if !errors.Is(err, ErrInviteNotFound) {
		t.Fatalf("want ErrInviteNotFound, got %v", err)
	}
}

func TestAcceptInvite_EmailMismatch(t *testing.T) {
	invID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000020")
	actor := uuid.MustParse("00000000-0000-0000-0000-000000000030")
	invites := &fakeInvitesStore{invites: map[uuid.UUID]db.FormInvite{
		invID: {
			ID:        pgtype.UUID{Bytes: invID, Valid: true},
			FormID:    pgtype.UUID{Bytes: formID, Valid: true},
			Email:     "a@example.com",
			Role:      "editor",
			InvitedBy: pgtype.UUID{Bytes: actor, Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		},
	}}
	roles := &fakeRolesStore{roles: map[uuid.UUID]db.FormRole{actor: db.FormRoleAdmin}}
	users := &fakeUsersStore{users: map[uuid.UUID]db.GetUserByIDRow{
		actor: {Email: "other@example.com"},
	}}
	svc := &Service{Invites: invites, Roles: roles, Users: users}

	if err := svc.AcceptInvite(
		context.Background(),
		invID.String(),
		actor,
	); !errors.Is(
		err,
		ErrForbidden,
	) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}
