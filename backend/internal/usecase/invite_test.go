package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/infra/postgres"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInviteUseCase_CreateInvite(t *testing.T) {
	t.Run("正常系: admin が招待を作成できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newInviteUseCase()
		invite, err := uc.CreateInvite(
			ctx,
			formID,
			adminID,
			"invitee@example.com",
			entity.RoleEditor,
		)
		require.NoError(t, err)
		assert.Equal(t, "invitee@example.com", invite.Email)
		assert.Equal(t, entity.RoleEditor, invite.Role)
	})

	t.Run("準正常系: editor は招待できず FORBIDDEN エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		editorID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"editor@example.com",
			"password123",
			"Editor",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)

		uc := newInviteUseCase()
		_, err := uc.CreateInvite(ctx, formID, editorID, "invitee@example.com", entity.RoleEditor)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeForbidden, appErr.Code)
	})

	t.Run("準正常系: メール送信失敗時に招待レコードが削除されること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		emailErr := errors.New("ses error")
		uc := usecase.NewInviteUseCase(
			newInviteRepo(),
			newMemberRepo(),
			newUserRepo(),
			postgres.NewInviteUoW(testPool),
			&mockEmailSender{
				sendEmailFunc: func(_ context.Context, _ repository.SendEmailInput) error {
					return emailErr
				},
			},
			"http://localhost:5173",
		)

		_, err := uc.CreateInvite(ctx, formID, adminID, "invitee@example.com", entity.RoleEditor)
		require.ErrorIs(t, err, emailErr)

		// 招待が残っていないことを確認（再招待が成功すること）
		ucOK := newInviteUseCase()
		invite, err := ucOK.CreateInvite(
			ctx,
			formID,
			adminID,
			"invitee@example.com",
			entity.RoleEditor,
		)
		require.NoError(t, err)
		assert.Equal(t, "invitee@example.com", invite.Email)
	})

	t.Run("準正常系: 既にメンバーのユーザーを招待すると ALREADY_MEMBER エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		editorID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"editor@example.com",
			"password123",
			"Editor",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)

		uc := newInviteUseCase()
		_, err := uc.CreateInvite(ctx, formID, adminID, "editor@example.com", entity.RoleEditor)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeAlreadyMember, appErr.Code)
	})
}

func TestInviteUseCase_AcceptInvite(t *testing.T) {
	t.Run("正常系: 招待を受諾してメンバーになれること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		inviteeID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"invitee@example.com",
			"password123",
			"Invitee",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newInviteUseCase()
		invite, err := uc.CreateInvite(
			ctx,
			formID,
			adminID,
			"invitee@example.com",
			entity.RoleEditor,
		)
		require.NoError(t, err)

		err = uc.AcceptInvite(ctx, invite.ID, inviteeID)
		require.NoError(t, err)

		// メンバーになったことを確認
		memberUC := newMemberUseCase()
		members, err := memberUC.ListMembers(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Len(t, members, 2)
	})

	t.Run("準正常系: メールアドレスが一致しない場合 RESOURCE_HIDDEN エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		wrongUserID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"wrong@example.com",
			"password123",
			"Wrong",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newInviteUseCase()
		invite, err := uc.CreateInvite(
			ctx,
			formID,
			adminID,
			"invitee@example.com",
			entity.RoleEditor,
		)
		require.NoError(t, err)

		err = uc.AcceptInvite(ctx, invite.ID, wrongUserID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})

	t.Run("準正常系: 存在しない招待で INVITE_NOT_FOUND エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"user@example.com",
			"password123",
			"User",
		)

		uc := newInviteUseCase()
		err := uc.AcceptInvite(ctx, testutil.RandomUUID(), userID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeInviteNotFound, appErr.Code)
	})
}

func TestInviteUseCase_DeleteInvite(t *testing.T) {
	t.Run("正常系: admin が招待を削除できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newInviteUseCase()
		invite, err := uc.CreateInvite(
			ctx,
			formID,
			adminID,
			"invitee@example.com",
			entity.RoleEditor,
		)
		require.NoError(t, err)

		err = uc.DeleteInvite(ctx, formID, adminID, invite.ID)
		require.NoError(t, err)

		invites, err := uc.ListInvites(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Len(t, invites, 0)
	})

	t.Run("準正常系: 存在しない招待で INVITE_NOT_FOUND エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newInviteUseCase()
		err := uc.DeleteInvite(ctx, formID, adminID, testutil.RandomUUID())
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeInviteNotFound, appErr.Code)
	})
}

func TestInviteUseCase_ListInvites(t *testing.T) {
	t.Run("準正常系: 非 admin は FORBIDDEN エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		editorID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"editor@example.com",
			"password123",
			"Editor",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)

		uc := newInviteUseCase()
		_, err := uc.ListInvites(ctx, formID, editorID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeForbidden, appErr.Code)
	})
}
