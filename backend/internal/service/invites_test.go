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

type stubInvitesStore struct {
	createFn func(ctx context.Context, arg db.CreateFormInviteParams) (db.FormInvite, error)
	listFn   func(ctx context.Context, arg db.ListActiveFormInvitesParams) ([]db.FormInvite, error)
	getFn    func(ctx context.Context, code string) (db.FormInvite, error)
	revokeFn func(ctx context.Context, arg db.RevokeFormInviteParams) (db.FormInvite, error)
}

func (s *stubInvitesStore) CreateFormInvite(ctx context.Context, arg db.CreateFormInviteParams) (db.FormInvite, error) {
	if s.createFn == nil {
		return db.FormInvite{}, nil
	}
	return s.createFn(ctx, arg)
}
func (s *stubInvitesStore) ListActiveFormInvites(ctx context.Context, arg db.ListActiveFormInvitesParams) ([]db.FormInvite, error) {
	if s.listFn == nil {
		return nil, nil
	}
	return s.listFn(ctx, arg)
}
func (s *stubInvitesStore) GetFormInviteForUpdate(ctx context.Context, code string) (db.FormInvite, error) {
	if s.getFn == nil {
		return db.FormInvite{}, nil
	}
	return s.getFn(ctx, code)
}
func (s *stubInvitesStore) RevokeFormInvite(ctx context.Context, arg db.RevokeFormInviteParams) (db.FormInvite, error) {
	if s.revokeFn == nil {
		return db.FormInvite{}, nil
	}
	return s.revokeFn(ctx, arg)
}

type stubRolesStore struct {
	getFn       func(ctx context.Context, arg db.GetUserFormRoleParams) (string, error)
	upsertFn    func(ctx context.Context, arg db.UpsertUserFormRoleParams) error
	deleteFn    func(ctx context.Context, arg db.DeleteUserFormRoleParams) error
	listMembers func(ctx context.Context, formID string) ([]db.ListFormMembersRow, error)
	listFormsFn func(ctx context.Context, userID pgtype.UUID) ([]db.ListUserAccessibleFormsRow, error)
}

func (s *stubRolesStore) GetUserFormRole(ctx context.Context, arg db.GetUserFormRoleParams) (string, error) {
	if s.getFn == nil {
		return "", nil
	}
	return s.getFn(ctx, arg)
}
func (s *stubRolesStore) UpsertUserFormRole(ctx context.Context, arg db.UpsertUserFormRoleParams) error {
	if s.upsertFn == nil {
		return nil
	}
	return s.upsertFn(ctx, arg)
}
func (s *stubRolesStore) DeleteUserFormRole(ctx context.Context, arg db.DeleteUserFormRoleParams) error {
	if s.deleteFn == nil {
		return nil
	}
	return s.deleteFn(ctx, arg)
}
func (s *stubRolesStore) ListFormMembers(ctx context.Context, formID string) ([]db.ListFormMembersRow, error) {
	if s.listMembers == nil {
		return nil, nil
	}
	return s.listMembers(ctx, formID)
}
func (s *stubRolesStore) ListUserAccessibleForms(ctx context.Context, userID pgtype.UUID) ([]db.ListUserAccessibleFormsRow, error) {
	if s.listFormsFn == nil {
		return nil, nil
	}
	return s.listFormsFn(ctx, userID)
}

func defaultStubRoles() *stubRolesStore {
	return &stubRolesStore{
		getFn: func(ctx context.Context, arg db.GetUserFormRoleParams) (string, error) {
			return "admin", nil
		},
	}
}

func stubInvite(formID, code string, expires time.Time, revoked bool) db.FormInvite {
	return db.FormInvite{
		Code:      code,
		FormID:    formID,
		Role:      "editor",
		ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
		CreatedBy: pgtype.UUID{Bytes: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Valid: true},
		Revoked:   revoked,
	}
}

// admin権限のユーザーが招待コードを発行でき、期限計算とコード生成が正しく反映される
func TestCreateInvite_Success(t *testing.T) {
	now := time.Unix(1, 0)
	svc := &Service{
		now:          nowFunc(now),
		generateCode: func() (string, error) { return "abc123", nil },
	}
	actor := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	var captured db.CreateFormInviteParams
	svc.Invites = &stubInvitesStore{
		createFn: func(ctx context.Context, arg db.CreateFormInviteParams) (db.FormInvite, error) {
			captured = arg
			return stubInvite(arg.FormID, arg.Code, arg.ExpiresAt.Time, false), nil
		},
	}
	svc.Roles = &stubRolesStore{
		getFn: func(ctx context.Context, arg db.GetUserFormRoleParams) (string, error) { return "admin", nil },
	}

	invite, err := svc.CreateInvite(context.Background(), "formA", actor)
	if err != nil {
		t.Fatalf("CreateInvite returned error: %v", err)
	}
	if invite.Code != "abc123" {
		t.Fatalf("want code abc123, got %s", invite.Code)
	}
	if captured.FormID != "formA" {
		t.Fatalf("unexpected formID %s", captured.FormID)
	}
	expectedExpiry := now.Add(inviteTTL)
	if !captured.ExpiresAt.Time.Equal(expectedExpiry) {
		t.Fatalf("unexpected expires_at: got %v want %v", captured.ExpiresAt.Time, expectedExpiry)
	}
}

// admin以外のユーザーが発行を試みたときErrForbiddenが返る
func TestCreateInvite_Forbidden(t *testing.T) {
	svc := &Service{
		now:          time.Now,
		generateCode: defaultInviteCode,
	}
	svc.Roles = &stubRolesStore{
		getFn: func(ctx context.Context, arg db.GetUserFormRoleParams) (string, error) { return "editor", nil },
	}

	_, err := svc.CreateInvite(context.Background(), "formA", uuid.Nil)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}

// 有効な招待の一覧をフォームIDと現在時刻をフィルタ条件として渡す
func TestListInvites_PassesFilters(t *testing.T) {
	now := time.Unix(1, 0)
	svc := &Service{
		now: nowFunc(now),
	}
	actor := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	svc.Invites = &stubInvitesStore{
		listFn: func(ctx context.Context, arg db.ListActiveFormInvitesParams) ([]db.FormInvite, error) {
			if arg.FormID != "formA" {
				t.Fatalf("unexpected formID %s", arg.FormID)
			}
			if !arg.ExpiresAt.Time.Equal(now) {
				t.Fatalf("unexpected cutoff time %v", arg.ExpiresAt.Time)
			}
			return []db.FormInvite{stubInvite("formA", "abc", now.Add(time.Hour), false)}, nil
		},
	}
	svc.Roles = defaultStubRoles()

	invites, err := svc.ListInvites(context.Background(), "formA", actor)
	if err != nil {
		t.Fatalf("ListInvites returned error: %v", err)
	}
	if len(invites) != 1 || invites[0].Code != "abc" {
		t.Fatalf("unexpected invites result %#v", invites)
	}
}

// 指定したコードが存在しない場合ErrInviteNotFoundを返す
func TestRevokeInvite_NotFound(t *testing.T) {
	svc := &Service{now: nowFunc(time.Now())}
	svc.Roles = defaultStubRoles()
	svc.Invites = &stubInvitesStore{
		revokeFn: func(ctx context.Context, arg db.RevokeFormInviteParams) (db.FormInvite, error) {
			return db.FormInvite{}, pgx.ErrNoRows
		},
	}

	_, err := svc.RevokeInvite(context.Background(), "formA", "code1", uuid.Nil)
	if !errors.Is(err, ErrInviteNotFound) {
		t.Fatalf("want ErrInviteNotFound, got %v", err)
	}
}

// 有効な招待コードでeditor権限付与と招待失効が実行される
func TestAcceptInvite_Success(t *testing.T) {
	now := time.Unix(10, 0)
	actor := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	svc := &Service{now: nowFunc(now)}

	invite := stubInvite("formA", "code123", now.Add(time.Hour), false)
	var upsertCalled bool
	var revokeCalled bool

	svc.Invites = &stubInvitesStore{
		getFn: func(ctx context.Context, code string) (db.FormInvite, error) {
			if code != "code123" {
				t.Fatalf("unexpected code %s", code)
			}
			return invite, nil
		},
		revokeFn: func(ctx context.Context, arg db.RevokeFormInviteParams) (db.FormInvite, error) {
			revokeCalled = true
			if arg.Code != "code123" {
				t.Fatalf("unexpected revoke code %s", arg.Code)
			}
			return invite, nil
		},
	}
	svc.Roles = &stubRolesStore{
		getFn: func(ctx context.Context, arg db.GetUserFormRoleParams) (string, error) { return "", pgx.ErrNoRows },
		upsertFn: func(ctx context.Context, arg db.UpsertUserFormRoleParams) error {
			upsertCalled = true
			if arg.Role != "editor" {
				t.Fatalf("unexpected role %s", arg.Role)
			}
			return nil
		},
	}

	if err := svc.AcceptInvite(context.Background(), "code123", actor); err != nil {
		t.Fatalf("AcceptInvite returned error: %v", err)
	}
	if !upsertCalled {
		t.Fatalf("expected upsertUserFormRole to be called")
	}
	if !revokeCalled {
		t.Fatalf("expected revoke to be called")
	}
}

// 既にeditor 以上の権限を持つユーザーが受理するとErrAlreadyMemberを返す
func TestAcceptInvite_AlreadyMember(t *testing.T) {
	svc := &Service{now: nowFunc(time.Now())}
	svc.Invites = &stubInvitesStore{
		getFn: func(ctx context.Context, code string) (db.FormInvite, error) {
			return stubInvite("formA", code, time.Now().Add(time.Hour), false), nil
		},
	}
	svc.Roles = &stubRolesStore{
		getFn: func(ctx context.Context, arg db.GetUserFormRoleParams) (string, error) { return "editor", nil },
	}

	err := svc.AcceptInvite(context.Background(), "code999", uuid.Nil)
	if !errors.Is(err, ErrAlreadyMember) {
		t.Fatalf("want ErrAlreadyMember, got %v", err)
	}
}

// 期限切れの招待コードを受理しようとした際にErrInviteExpiredを返す
func TestAcceptInvite_Expired(t *testing.T) {
	now := time.Unix(10, 0)
	svc := &Service{now: nowFunc(now)}
	svc.Invites = &stubInvitesStore{
		getFn: func(ctx context.Context, code string) (db.FormInvite, error) {
			return stubInvite("formA", code, now, false), nil
		},
	}
	svc.Roles = &stubRolesStore{
		getFn: func(ctx context.Context, arg db.GetUserFormRoleParams) (string, error) { return "", pgx.ErrNoRows },
	}

	err := svc.AcceptInvite(context.Background(), "code", uuid.Nil)
	if !errors.Is(err, ErrInviteExpired) {
		t.Fatalf("want ErrInviteExpired, got %v", err)
	}
}

func nowFunc(t time.Time) func() time.Time {
	return func() time.Time { return t }
}
