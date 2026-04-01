package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemberUseCase_AddMember(t *testing.T) {
	t.Run("正常系: admin がメンバーを追加できること", func(t *testing.T) {
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
		testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"newmember@example.com",
			"password123",
			"New Member",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newMemberUseCase()
		err := uc.AddMember(ctx, formID, adminID, "newmember@example.com", entity.RoleEditor)
		require.NoError(t, err)

		members, err := uc.ListMembers(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Len(t, members, 2)
	})

	t.Run("準正常系: editor はメンバー追加できず FORBIDDEN エラーになること", func(t *testing.T) {
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
		testutil.CreateVerifiedUser(t, ctx, testPool, "target@example.com", "password123", "Target")
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)

		uc := newMemberUseCase()
		err := uc.AddMember(ctx, formID, editorID, "target@example.com", entity.RoleEditor)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeForbidden, appErr.Code)
	})

	t.Run("準正常系: 既にメンバーの場合 ALREADY_MEMBER エラーになること", func(t *testing.T) {
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

		uc := newMemberUseCase()
		err := uc.AddMember(ctx, formID, adminID, "admin@example.com", entity.RoleEditor)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeAlreadyMember, appErr.Code)
	})

	t.Run("準正常系: 存在しないユーザーのメールアドレスで USER_NOT_FOUND エラーになること", func(t *testing.T) {
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

		uc := newMemberUseCase()
		err := uc.AddMember(ctx, formID, adminID, "nobody@example.com", entity.RoleEditor)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeUserNotFound, appErr.Code)
	})
}

func TestMemberUseCase_ChangeRole(t *testing.T) {
	t.Run("正常系: admin がメンバーのロールを変更できること", func(t *testing.T) {
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

		uc := newMemberUseCase()
		err := uc.ChangeRole(ctx, formID, adminID, editorID, entity.RoleAdmin)
		require.NoError(t, err)

		members, err := uc.ListMembers(ctx, formID, adminID)
		require.NoError(t, err)
		var changed *entity.Member
		for i := range members {
			if members[i].ID == editorID {
				changed = &members[i]
				break
			}
		}
		require.NotNil(t, changed)
		assert.Equal(t, entity.RoleAdmin, changed.Role)
	})

	t.Run("準正常系: 唯一の admin を editor に降格すると LAST_ADMIN エラーになること", func(t *testing.T) {
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

		uc := newMemberUseCase()
		err := uc.ChangeRole(ctx, formID, adminID, adminID, entity.RoleEditor)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeLastAdmin, appErr.Code)
	})
}

func TestMemberUseCase_RemoveMember(t *testing.T) {
	t.Run("正常系: admin がメンバーを削除できること", func(t *testing.T) {
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

		uc := newMemberUseCase()
		err := uc.RemoveMember(ctx, formID, adminID, editorID)
		require.NoError(t, err)

		members, err := uc.ListMembers(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Len(t, members, 1)
	})

	t.Run("準正常系: 唯一の admin を削除すると LAST_ADMIN エラーになること", func(t *testing.T) {
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

		uc := newMemberUseCase()
		err := uc.RemoveMember(ctx, formID, adminID, adminID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeLastAdmin, appErr.Code)
	})

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
		targetID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"target@example.com",
			"password123",
			"Target",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)
		testutil.AddMember(t, ctx, testPool, targetID, formID, entity.RoleEditor)

		uc := newMemberUseCase()
		err := uc.RemoveMember(ctx, formID, editorID, targetID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeForbidden, appErr.Code)
	})
}

func TestMemberUseCase_ListMembers(t *testing.T) {
	t.Run("正常系: editor がメンバー一覧を取得できること", func(t *testing.T) {
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

		uc := newMemberUseCase()
		members, err := uc.ListMembers(ctx, formID, editorID)
		require.NoError(t, err)
		assert.Len(t, members, 2)
	})

	t.Run("準正常系: 非メンバーは RESOURCE_HIDDEN エラーになること", func(t *testing.T) {
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
		outsiderID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"outsider@example.com",
			"password123",
			"Outsider",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newMemberUseCase()
		_, err := uc.ListMembers(ctx, formID, outsiderID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}
