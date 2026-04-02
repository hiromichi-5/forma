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

func TestProfileUseCase_GetProfile(t *testing.T) {
	t.Run("正常系: プロフィールを取得できること", func(t *testing.T) {
		truncate(t)
		uc := newProfileUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"profile@example.com",
			"password123",
			"Profile User",
		)

		user, err := uc.GetProfile(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "profile@example.com", user.Email)
		assert.Equal(t, "Profile User", user.DisplayName)
	})

	t.Run("準正常系: 存在しないユーザーで USER_NOT_FOUND エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newProfileUseCase()
		ctx := context.Background()

		_, err := uc.GetProfile(ctx, testutil.RandomUUID())
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeUserNotFound, appErr.Code)
	})
}

func TestProfileUseCase_UpdateDisplayName(t *testing.T) {
	t.Run("正常系: 表示名を更新できること", func(t *testing.T) {
		truncate(t)
		uc := newProfileUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"update@example.com",
			"password123",
			"Old Name",
		)

		user, err := uc.UpdateDisplayName(ctx, userID, "New Name")
		require.NoError(t, err)
		assert.Equal(t, "New Name", user.DisplayName)
	})

	t.Run("準正常系: 空文字で VALIDATION_ERROR になること", func(t *testing.T) {
		truncate(t)
		uc := newProfileUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"update@example.com",
			"password123",
			"Name",
		)

		_, err := uc.UpdateDisplayName(ctx, userID, "")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeValidation, appErr.Code)
	})
}

func TestProfileUseCase_ChangePassword(t *testing.T) {
	t.Run("正常系: パスワードを変更できること", func(t *testing.T) {
		truncate(t)
		profileUC := newProfileUseCase()
		authUC := newAuthUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"chpw@example.com",
			"oldpass123",
			"User",
		)

		err := profileUC.ChangePassword(ctx, userID, "oldpass123", "newpass123")
		require.NoError(t, err)

		// 新パスワードでログインできることを確認
		session, err := authUC.Authenticate(ctx, "chpw@example.com", "newpass123")
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
	})

	t.Run("準正常系: 現在のパスワードが間違っている場合 INCORRECT_PASSWORD エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newProfileUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"chpw@example.com",
			"password123",
			"User",
		)

		err := uc.ChangePassword(ctx, userID, "wrongpass", "newpass123")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeIncorrectPassword, appErr.Code)
	})

	t.Run("準正常系: 新パスワードが短すぎる場合 VALIDATION_ERROR になること", func(t *testing.T) {
		truncate(t)
		uc := newProfileUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"chpw@example.com",
			"password123",
			"User",
		)

		err := uc.ChangePassword(ctx, userID, "password123", "short")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeValidation, appErr.Code)
	})
}

func TestProfileUseCase_DeleteProfile(t *testing.T) {
	t.Run("正常系: プロフィールを削除できること", func(t *testing.T) {
		truncate(t)
		uc := newProfileUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"del@example.com",
			"password123",
			"User",
		)

		err := uc.DeleteProfile(ctx, userID)
		require.NoError(t, err)

		// 削除後に取得できないことを確認
		_, err = uc.GetProfile(ctx, userID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeUserNotFound, appErr.Code)
	})
}
