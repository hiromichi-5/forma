package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormUseCase_RegisterForm(t *testing.T) {
	t.Run("正常系: フォームを登録できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)

		fetcher := &mockFormFetcher{
			getFormFunc: func(_ context.Context, formID string) (*repository.GoogleForm, error) {
				return &repository.GoogleForm{
					FormID: formID,
					Title:  "Test Form",
				}, nil
			},
		}
		uc := newFormUseCase(fetcher)

		form, err := uc.RegisterForm(
			ctx,
			"https://docs.google.com/forms/d/e/abc123def456ghij/viewform",
			userID,
		)
		require.NoError(t, err)
		assert.Equal(t, "Test Form", form.Title)
		assert.NotEmpty(t, form.ID)
	})

	t.Run("準正常系: 同じフォームを重複登録すると FORM_ALREADY_REGISTERED エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)

		fetcher := &mockFormFetcher{
			getFormFunc: func(_ context.Context, formID string) (*repository.GoogleForm, error) {
				return &repository.GoogleForm{
					FormID: formID,
					Title:  "Test Form",
				}, nil
			},
		}
		uc := newFormUseCase(fetcher)

		_, err := uc.RegisterForm(
			ctx,
			"https://docs.google.com/forms/d/e/abc123def456ghij/viewform",
			userID,
		)
		require.NoError(t, err)

		_, err = uc.RegisterForm(
			ctx,
			"https://docs.google.com/forms/d/e/abc123def456ghij/viewform",
			userID,
		)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeFormAlreadyRegistered, appErr.Code)
	})

	t.Run("準正常系: Google Forms API が 403 を返す場合 FORM_NOT_SHARED エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)

		fetcher := &mockFormFetcher{
			getFormFunc: func(_ context.Context, _ string) (*repository.GoogleForm, error) {
				return nil, repository.ErrForbidden
			},
		}
		uc := newFormUseCase(fetcher)

		_, err := uc.RegisterForm(
			ctx,
			"https://docs.google.com/forms/d/e/abc123def456ghij/viewform",
			userID,
		)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeFormNotShared, appErr.Code)
	})

	t.Run("正常系: 登録と同時に初回同期が行われチケットが作成されること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)

		fetcher := &mockFormFetcher{
			getFormFunc: func(_ context.Context, formID string) (*repository.GoogleForm, error) {
				return &repository.GoogleForm{
					FormID: formID,
					Title:  "Test Form",
				}, nil
			},
			listResponsesFunc: func(_ context.Context, _, _, _ string) (*repository.GoogleFormResponsePage, error) {
				return &repository.GoogleFormResponsePage{
					Responses: []repository.GoogleFormResponse{
						{
							ResponseID:  "resp-1",
							SubmittedAt: time.Now(),
							AnswersJSON: []byte(`{}`),
						},
					},
				}, nil
			},
		}
		uc := newFormUseCase(fetcher)

		form, err := uc.RegisterForm(
			ctx,
			"https://docs.google.com/forms/d/e/abc123def456ghij/viewform",
			userID,
		)
		require.NoError(t, err)

		ticketUC := newTicketUseCase()
		tickets, err := ticketUC.ListTickets(ctx, form.ID, userID, nil)
		require.NoError(t, err)
		assert.Len(t, tickets, 1)
	})

	t.Run("準正常系: 初回同期に失敗してもフォーム登録自体は成功すること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)

		fetcher := &mockFormFetcher{
			getFormFunc: func(_ context.Context, formID string) (*repository.GoogleForm, error) {
				return &repository.GoogleForm{
					FormID: formID,
					Title:  "Test Form",
				}, nil
			},
		}
		syncer := &mockFormSyncer{
			syncFormOnceFunc: func(context.Context, uuid.UUID, uuid.UUID) (int, time.Time, error) {
				return 0, time.Time{}, errors.New("sync failed")
			},
		}
		uc := newFormUseCaseWithSyncer(fetcher, syncer)

		form, err := uc.RegisterForm(
			ctx,
			"https://docs.google.com/forms/d/e/abc123def456ghij/viewform",
			userID,
		)
		require.NoError(t, err)
		assert.Equal(t, "Test Form", form.Title)
	})
}

func TestFormUseCase_GetForm(t *testing.T) {
	t.Run("正常系: メンバーがフォーム詳細を取得できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "My Form", userID)

		uc := newFormUseCase(&mockFormFetcher{})
		form, err := uc.GetForm(ctx, formID, userID)
		require.NoError(t, err)
		assert.Equal(t, "My Form", form.Title)
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
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "My Form", adminID)

		uc := newFormUseCase(&mockFormFetcher{})
		_, err := uc.GetForm(ctx, formID, outsiderID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}

func TestFormUseCase_DeleteForm(t *testing.T) {
	t.Run("正常系: admin がフォームを削除できること", func(t *testing.T) {
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
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "My Form", adminID)
		testutil.CreateTicket(t, ctx, testPool, formID, statusID, "response1")

		uc := newFormUseCase(&mockFormFetcher{})
		err := uc.DeleteForm(ctx, formID, adminID)
		require.NoError(t, err)

		_, err = uc.GetForm(ctx, formID, adminID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})

	t.Run("準正常系: editor は削除できず FORBIDDEN エラーになること", func(t *testing.T) {
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
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "My Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)

		uc := newFormUseCase(&mockFormFetcher{})
		err := uc.DeleteForm(ctx, formID, editorID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeForbidden, appErr.Code)
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
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "My Form", adminID)

		uc := newFormUseCase(&mockFormFetcher{})
		err := uc.DeleteForm(ctx, formID, outsiderID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}

func TestFormUseCase_ListForms(t *testing.T) {
	t.Run("正常系: ユーザーがアクセス可能なフォーム一覧を取得できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"admin@example.com",
			"password123",
			"Admin",
		)
		testutil.CreateForm(t, ctx, testPool, "gform1", "Form 1", userID)
		testutil.CreateForm(t, ctx, testPool, "gform2", "Form 2", userID)

		uc := newFormUseCase(&mockFormFetcher{})
		forms, err := uc.ListForms(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, forms, 2)
	})
}

func TestFormUseCase_ListQuestions(t *testing.T) {
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
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "My Form", adminID)

		uc := newFormUseCase(&mockFormFetcher{})
		_, err := uc.ListQuestions(ctx, formID, outsiderID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}
