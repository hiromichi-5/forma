package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusUseCase_ListStatuses(t *testing.T) {
	t.Run("正常系: メンバーがステータス一覧を取得できること", func(t *testing.T) {
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

		uc := newStatusUseCase()
		statuses, err := uc.ListStatuses(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Len(t, statuses, 3) // デフォルト3つ
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

		uc := newStatusUseCase()
		_, err := uc.ListStatuses(ctx, formID, outsiderID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}

func TestStatusUseCase_CreateStatus(t *testing.T) {
	t.Run("正常系: 非デフォルトのステータスを作成できること", func(t *testing.T) {
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

		uc := newStatusUseCase()
		color := "#0000FF"
		status, err := uc.CreateStatus(ctx, formID, adminID, usecase.CreateStatusInput{
			Name:         "保留",
			Color:        &color,
			DisplayOrder: 4,
		})
		require.NoError(t, err)
		assert.Equal(t, "保留", status.Name)
		assert.False(t, status.IsDefault)
	})

	t.Run("正常系: デフォルトのステータスを作成すると既存デフォルトが解除されること", func(t *testing.T) {
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

		uc := newStatusUseCase()
		newDefault, err := uc.CreateStatus(ctx, formID, adminID, usecase.CreateStatusInput{
			Name:         "新デフォルト",
			DisplayOrder: 4,
			IsDefault:    true,
		})
		require.NoError(t, err)
		assert.True(t, newDefault.IsDefault)

		// 既存のデフォルトが解除されたことを確認
		statuses, err := uc.ListStatuses(ctx, formID, adminID)
		require.NoError(t, err)
		defaultCount := 0
		for _, s := range statuses {
			if s.IsDefault {
				defaultCount++
				assert.Equal(t, newDefault.ID, s.ID)
			}
		}
		assert.Equal(t, 1, defaultCount)
	})

	t.Run("準正常系: 空の名前で VALIDATION_ERROR になること", func(t *testing.T) {
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

		uc := newStatusUseCase()
		_, err := uc.CreateStatus(ctx, formID, adminID, usecase.CreateStatusInput{DisplayOrder: 4})
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeValidation, appErr.Code)
	})
}

func TestStatusUseCase_UpdateStatus(t *testing.T) {
	t.Run("正常系: ステータス名を更新できること", func(t *testing.T) {
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

		uc := newStatusUseCase()
		statuses, err := uc.ListStatuses(ctx, formID, adminID)
		require.NoError(t, err)

		newName := "対応済み"
		updated, err := uc.UpdateStatus(
			ctx,
			formID,
			adminID,
			statuses[2].ID,
			usecase.UpdateStatusInput{Name: &newName},
		)
		require.NoError(t, err)
		assert.Equal(t, "対応済み", updated.Name)
	})

	t.Run("準正常系: 他フォームのステータスを更新すると RESOURCE_HIDDEN エラーになること", func(t *testing.T) {
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
		formID1, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form 1", adminID)
		formID2, _ := testutil.CreateForm(t, ctx, testPool, "gform2", "Form 2", adminID)

		uc := newStatusUseCase()
		statuses2, err := uc.ListStatuses(ctx, formID2, adminID)
		require.NoError(t, err)
		require.NotEmpty(t, statuses2)

		_, err = uc.UpdateStatus(
			ctx,
			formID1,
			adminID,
			statuses2[0].ID,
			usecase.UpdateStatusInput{},
		)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}

func TestStatusUseCase_DeleteStatus(t *testing.T) {
	t.Run("正常系: チケットが紐づかない非デフォルトステータスを削除できること", func(t *testing.T) {
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

		uc := newStatusUseCase()
		// 非デフォルトのステータスを特定（対応中 or 対応完了）
		statuses, err := uc.ListStatuses(ctx, formID, adminID)
		require.NoError(t, err)
		var nonDefault entity.FormStatus
		for _, s := range statuses {
			if !s.IsDefault {
				nonDefault = s
				break
			}
		}
		require.NotEmpty(t, nonDefault.ID)

		err = uc.DeleteStatus(ctx, formID, adminID, nonDefault.ID)
		require.NoError(t, err)

		statuses, err = uc.ListStatuses(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Len(t, statuses, 2)
		for _, s := range statuses {
			assert.NotEqual(t, nonDefault.ID, s.ID)
		}
	})

	t.Run("準正常系: デフォルトステータスを削除すると VALIDATION_ERROR になること", func(t *testing.T) {
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
		formID, defaultStatusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newStatusUseCase()
		err := uc.DeleteStatus(ctx, formID, adminID, defaultStatusID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeValidation, appErr.Code)
	})

	t.Run("準正常系: チケットが紐づくステータスを削除すると VALIDATION_ERROR になること", func(t *testing.T) {
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

		uc := newStatusUseCase()
		// 非デフォルトステータスを作成してチケットを紐づける
		color := "#0000FF"
		status, err := uc.CreateStatus(ctx, formID, adminID, usecase.CreateStatusInput{
			Name:         "テスト",
			Color:        &color,
			DisplayOrder: 4,
		})
		require.NoError(t, err)

		testutil.CreateTicket(t, ctx, testPool, formID, status.ID, "resp-1")

		err = uc.DeleteStatus(ctx, formID, adminID, status.ID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeValidation, appErr.Code)
	})
}
