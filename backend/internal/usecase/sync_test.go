package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncUseCase_SyncFormOnce(t *testing.T) {
	t.Run("正常系: レスポンスを同期してチケットが作成されること", func(t *testing.T) {
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
		formID, _ := testutil.CreateForm(t, ctx, testPool, "google-form-id-1", "Form", adminID)

		answersJSON, _ := json.Marshal(map[string]any{
			"q1": map[string]any{
				"questionId": "q1",
				"textAnswers": map[string]any{
					"answers": []map[string]string{{"value": "answer1"}},
				},
			},
		})

		fetcher := &mockFormFetcher{
			getFormFunc: func(_ context.Context, _ string) (*repository.GoogleForm, error) {
				return &repository.GoogleForm{
					FormID: "google-form-id-1",
					Title:  "Form",
					Items: []repository.GoogleFormItem{
						{
							Title: "Question 1",
							Questions: []repository.GoogleFormQuestion{
								{QuestionID: "q1", QuestionType: "TEXT"},
							},
						},
					},
				}, nil
			},
			listResponsesFunc: func(_ context.Context, _, _, _ string) (*repository.GoogleFormResponsePage, error) {
				return &repository.GoogleFormResponsePage{
					Responses: []repository.GoogleFormResponse{
						{
							ResponseID:      "resp-1",
							RespondentEmail: "user@example.com",
							SubmittedAt:     time.Now(),
							AnswersJSON:     answersJSON,
						},
						{
							ResponseID:      "resp-2",
							RespondentEmail: "user2@example.com",
							SubmittedAt:     time.Now(),
							AnswersJSON:     answersJSON,
						},
					},
				}, nil
			},
		}

		uc := newSyncUseCase(fetcher)
		newTickets, lastSync, err := uc.SyncFormOnce(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Equal(t, 2, newTickets)
		assert.False(t, lastSync.IsZero())

		// チケットが作成されたことを確認
		ticketUC := newTicketUseCase()
		tickets, err := ticketUC.ListTickets(ctx, formID, adminID, nil)
		require.NoError(t, err)
		assert.Len(t, tickets, 2)
	})

	t.Run("正常系: 重複レスポンスは新規チケットとしてカウントされないこと", func(t *testing.T) {
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
		formID, defaultStatusID := testutil.CreateForm(
			t,
			ctx,
			testPool,
			"google-form-id-1",
			"Form",
			adminID,
		)

		// 既存チケットを作成
		testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		fetcher := &mockFormFetcher{
			getFormFunc: func(_ context.Context, _ string) (*repository.GoogleForm, error) {
				return &repository.GoogleForm{
					FormID: "google-form-id-1",
					Title:  "Form",
				}, nil
			},
			listResponsesFunc: func(_ context.Context, _, _, _ string) (*repository.GoogleFormResponsePage, error) {
				return &repository.GoogleFormResponsePage{
					Responses: []repository.GoogleFormResponse{
						{
							ResponseID:  "resp-1", // 重複
							SubmittedAt: time.Now(),
							AnswersJSON: []byte(`{}`),
						},
					},
				}, nil
			},
		}

		uc := newSyncUseCase(fetcher)
		newTickets, _, err := uc.SyncFormOnce(ctx, formID, adminID)
		require.NoError(t, err)
		assert.Equal(t, 0, newTickets)
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
		formID, _ := testutil.CreateForm(t, ctx, testPool, "google-form-id-1", "Form", adminID)

		fetcher := &mockFormFetcher{}
		uc := newSyncUseCase(fetcher)
		_, _, err := uc.SyncFormOnce(ctx, formID, outsiderID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}
