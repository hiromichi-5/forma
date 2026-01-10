package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type rolesStoreStub struct {
	role        db.FormRole
	adminCount  int64
	getErr      error
	countErr    error
	upsertArgs  []db.UpsertFormMemberParams
	deleteArgs  []db.DeleteFormMemberParams
	countCalled int
}

func (s *rolesStoreStub) GetFormMemberRole(ctx context.Context, arg db.GetFormMemberRoleParams) (db.FormRole, error) {
	if s.getErr != nil {
		return "", s.getErr
	}
	return s.role, nil
}

func (s *rolesStoreStub) UpsertFormMember(ctx context.Context, arg db.UpsertFormMemberParams) error {
	s.upsertArgs = append(s.upsertArgs, arg)
	return nil
}

func (s *rolesStoreStub) DeleteFormMember(ctx context.Context, arg db.DeleteFormMemberParams) error {
	s.deleteArgs = append(s.deleteArgs, arg)
	return nil
}

func (s *rolesStoreStub) ListFormMembers(ctx context.Context, formID pgtype.UUID) ([]db.ListFormMembersRow, error) {
	return nil, nil
}

func (s *rolesStoreStub) ListUserAccessibleForms(ctx context.Context, userID pgtype.UUID) ([]db.ListUserAccessibleFormsRow, error) {
	return nil, nil
}

func (s *rolesStoreStub) CountFormAdmins(ctx context.Context, formID pgtype.UUID) (int64, error) {
	s.countCalled++
	if s.countErr != nil {
		return 0, s.countErr
	}
	return s.adminCount, nil
}

func TestChangeRole_PreventsDemotingLastAdmin(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	stub := &rolesStoreStub{role: db.FormRoleAdmin, adminCount: 1}
	svc := &Service{Roles: stub}

	err := svc.ChangeRole(context.Background(), formID.String(), uid.String(), "editor")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want ErrValidation, got %v", err)
	}
	if len(stub.upsertArgs) != 0 {
		t.Fatalf("unexpected upsert call: %+v", stub.upsertArgs)
	}
}

func TestChangeRole_AllowsDemoteWhenAnotherAdminExists(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	stub := &rolesStoreStub{role: db.FormRoleAdmin, adminCount: 2}
	svc := &Service{Roles: stub}

	if err := svc.ChangeRole(context.Background(), formID.String(), uid.String(), "editor"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.upsertArgs) != 1 {
		t.Fatalf("expected single upsert call, got %d", len(stub.upsertArgs))
	}
	if stub.upsertArgs[0].Role != db.FormRoleEditor {
		t.Fatalf("unexpected role written: %s", stub.upsertArgs[0].Role)
	}
}

func TestRemoveMember_PreventsDeletingLastAdmin(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	stub := &rolesStoreStub{role: db.FormRoleAdmin, adminCount: 1}
	svc := &Service{Roles: stub}

	err := svc.RemoveMember(context.Background(), formID.String(), uid.String())
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want ErrValidation, got %v", err)
	}
	if len(stub.deleteArgs) != 0 {
		t.Fatalf("unexpected delete call: %+v", stub.deleteArgs)
	}
}

func TestRemoveMember_AllowsNonAdminRemoval(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	formID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	stub := &rolesStoreStub{role: db.FormRoleEditor, adminCount: 1}
	svc := &Service{Roles: stub}

	if err := svc.RemoveMember(context.Background(), formID.String(), uid.String()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.deleteArgs) != 1 {
		t.Fatalf("expected delete call, got %d", len(stub.deleteArgs))
	}
	if stub.countCalled != 0 {
		t.Fatalf("count should not be called for non-admin, got %d", stub.countCalled)
	}
}
