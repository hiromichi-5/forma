package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketUseCase_ListTickets(t *testing.T) {
	t.Run("正常系: メンバーがチケット一覧を取得できること", func(t *testing.T) {
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
		testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")
		testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-2")

		uc := newTicketUseCase()
		page, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{})
		require.NoError(t, err)
		assert.Len(t, page.Tickets, 2)
		assert.Nil(t, page.NextCursor)
	})

	t.Run("正常系: ステータスでフィルタできること", func(t *testing.T) {
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
		testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		uc := newTicketUseCase()
		page, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{
			StatusIDs: []uuid.UUID{defaultStatusID},
		})
		require.NoError(t, err)
		assert.Len(t, page.Tickets, 1)
	})

	t.Run("正常系: 複数のステータスでフィルタできること", func(t *testing.T) {
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
		statusIDs := formStatusIDs(t, ctx, formID)
		testutil.CreateTicket(t, ctx, testPool, formID, statusIDs[0], "resp-1")
		testutil.CreateTicket(t, ctx, testPool, formID, statusIDs[1], "resp-2")
		testutil.CreateTicket(t, ctx, testPool, formID, statusIDs[2], "resp-3")

		uc := newTicketUseCase()
		page, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{
			StatusIDs: statusIDs[:2],
		})
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"resp-1", "resp-2"}, responseIDs(page.Tickets))
	})

	t.Run("正常系: メールアドレスの部分一致で検索でき、% と _ はワイルドカードにならないこと", func(t *testing.T) {
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
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.CreateTicketWithRespondent(
			t,
			ctx,
			testPool,
			formID,
			statusID,
			"resp-1",
			"A_B@example.com",
		)
		testutil.CreateTicketWithRespondent(
			t,
			ctx,
			testPool,
			formID,
			statusID,
			"resp-2",
			"axb@example.com",
		)
		testutil.CreateTicket(t, ctx, testPool, formID, statusID, "resp-3")

		uc := newTicketUseCase()
		query := "a_b"
		page, err := uc.ListTickets(
			ctx,
			formID,
			adminID,
			usecase.ListTicketsInput{EmailQuery: &query},
		)
		require.NoError(t, err)
		assert.Equal(t, []string{"resp-1"}, responseIDs(page.Tickets))
	})

	t.Run("正常系: 回答日時の新しい順にカーソルでページングできること", func(t *testing.T) {
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
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		for i, id := range []string{"resp-1", "resp-2", "resp-3"} {
			ticketID := testutil.CreateTicket(t, ctx, testPool, formID, statusID, id)
			testutil.SetTicketSubmittedAt(
				t,
				ctx,
				testPool,
				ticketID,
				base.Add(time.Duration(i)*time.Hour),
			)
		}

		uc := newTicketUseCase()
		first, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{Limit: 2})
		require.NoError(t, err)
		assert.Equal(t, []string{"resp-3", "resp-2"}, responseIDs(first.Tickets))
		require.NotNil(t, first.NextCursor)

		second, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{
			After: first.NextCursor,
			Limit: 2,
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"resp-1"}, responseIDs(second.Tickets))
		assert.Nil(t, second.NextCursor)
	})

	t.Run("正常系: 回答日時が同じでも重複・欠落なくページングできること", func(t *testing.T) {
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
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		submittedAt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		for _, id := range []string{"resp-1", "resp-2", "resp-3"} {
			ticketID := testutil.CreateTicket(t, ctx, testPool, formID, statusID, id)
			testutil.SetTicketSubmittedAt(t, ctx, testPool, ticketID, submittedAt)
		}

		uc := newTicketUseCase()
		var got []string
		var after *usecase.TicketCursor
		for {
			page, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{
				After: after,
				Limit: 1,
			})
			require.NoError(t, err)
			got = append(got, responseIDs(page.Tickets)...)
			if page.NextCursor == nil {
				break
			}
			after = page.NextCursor
		}
		assert.ElementsMatch(t, []string{"resp-1", "resp-2", "resp-3"}, got)
	})

	t.Run("準正常系: 件数が範囲外なら VALIDATION_ERROR になること", func(t *testing.T) {
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

		uc := newTicketUseCase()
		for _, limit := range []int{-1, usecase.MaxTicketPageSize + 1} {
			_, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{Limit: limit})
			var appErr *entity.Error
			require.True(t, errors.As(err, &appErr))
			assert.Equal(t, entity.CodeValidation, appErr.Code)
		}
	})

	t.Run("準正常系: 他フォームのステータスを指定すると RESOURCE_HIDDEN エラーになること", func(t *testing.T) {
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
		_, otherStatusID := testutil.CreateForm(t, ctx, testPool, "gform2", "Other", adminID)

		uc := newTicketUseCase()
		_, err := uc.ListTickets(ctx, formID, adminID, usecase.ListTicketsInput{
			StatusIDs: []uuid.UUID{otherStatusID},
		})
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
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

		uc := newTicketUseCase()
		_, err := uc.ListTickets(ctx, formID, outsiderID, usecase.ListTicketsInput{})
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}

func TestTicketUseCase_CountTicketsByStatus(t *testing.T) {
	t.Run("正常系: 件数0のステータスも含めて表示順に返し、メール検索を反映すること", func(t *testing.T) {
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
		statusIDs := formStatusIDs(t, ctx, formID)
		testutil.CreateTicketWithRespondent(
			t,
			ctx,
			testPool,
			formID,
			statusIDs[0],
			"resp-1",
			"a@example.com",
		)
		testutil.CreateTicketWithRespondent(
			t,
			ctx,
			testPool,
			formID,
			statusIDs[0],
			"resp-2",
			"b@example.com",
		)
		testutil.CreateTicketWithRespondent(
			t,
			ctx,
			testPool,
			formID,
			statusIDs[2],
			"resp-3",
			"c@example.com",
		)

		uc := newTicketUseCase()
		counts, err := uc.CountTicketsByStatus(ctx, formID, adminID, nil)
		require.NoError(t, err)
		assert.Equal(t, []usecase.TicketStatusCount{
			{StatusID: statusIDs[0], Count: 2},
			{StatusID: statusIDs[1], Count: 0},
			{StatusID: statusIDs[2], Count: 1},
		}, counts)

		query := "a@"
		counts, err = uc.CountTicketsByStatus(ctx, formID, adminID, &query)
		require.NoError(t, err)
		assert.Equal(t, int64(1), counts[0].Count)
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

		uc := newTicketUseCase()
		_, err := uc.CountTicketsByStatus(ctx, formID, outsiderID, nil)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}

func formStatusIDs(t *testing.T, ctx context.Context, formID uuid.UUID) []uuid.UUID {
	t.Helper()
	statuses, err := newStatusRepo().List(ctx, formID)
	require.NoError(t, err)
	ids := make([]uuid.UUID, len(statuses))
	for i, s := range statuses {
		ids[i] = s.ID
	}
	return ids
}

func responseIDs(tickets []usecase.TicketSummary) []string {
	ids := make([]string, len(tickets))
	for i, ticket := range tickets {
		ids[i] = ticket.ResponseID
	}
	return ids
}

func TestTicketUseCase_GetTicket(t *testing.T) {
	t.Run("正常系: チケット詳細を取得できること", func(t *testing.T) {
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
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		uc := newTicketUseCase()
		detail, err := uc.GetTicket(ctx, ticketID, adminID)
		require.NoError(t, err)
		assert.Equal(t, ticketID, detail.ID)
		assert.Equal(t, "resp-1", detail.ResponseID)
	})

	t.Run("準正常系: 存在しないチケットで RESOURCE_HIDDEN エラーになること", func(t *testing.T) {
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

		uc := newTicketUseCase()
		_, err := uc.GetTicket(ctx, testutil.RandomUUID(), adminID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}

func TestTicketUseCase_UpdateTicket(t *testing.T) {
	t.Run("正常系: チケット更新に成功すると更新イベントを発行すること", func(t *testing.T) {
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
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		statusUC := newStatusUseCase()
		statuses, err := statusUC.ListStatuses(ctx, formID, adminID)
		require.NoError(t, err)
		newStatusID := statuses[1].ID

		publisher := &recordingEventPublisher{}
		uc := newTicketUseCaseWithPublisher(publisher)
		_, _, err = uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{StatusID: &newStatusID},
		)
		require.NoError(t, err)

		events := publisher.events()
		require.Len(t, events, 1)
		assert.Equal(t, formID, events[0].FormID)
		assert.Equal(t, ticketID, events[0].TicketID)
	})

	t.Run("準正常系: チケット更新に失敗した場合は更新イベントを発行しないこと", func(t *testing.T) {
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
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		publisher := &recordingEventPublisher{}
		uc := newTicketUseCaseWithPublisher(publisher)
		invalid := entity.Priority("invalid")
		_, _, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{Priority: &invalid},
		)
		require.Error(t, err)
		assert.Empty(t, publisher.events())
	})

	t.Run("正常系: ステータスを変更すると履歴が記録されること", func(t *testing.T) {
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
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		statusUC := newStatusUseCase()
		statuses, err := statusUC.ListStatuses(ctx, formID, adminID)
		require.NoError(t, err)
		newStatusID := statuses[1].ID // 対応中

		uc := newTicketUseCase()
		detail, _, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{StatusID: &newStatusID},
		)
		require.NoError(t, err)
		assert.Equal(t, newStatusID, detail.Status.ID)

		// 履歴が記録されていることを確認
		histories, err := uc.ListTicketHistories(ctx, ticketID, adminID)
		require.NoError(t, err)
		assert.Len(t, histories, 1)
		assert.Equal(t, entity.FieldStatus, histories[0].FieldName)
	})

	t.Run("正常系: 担当者を変更できること", func(t *testing.T) {
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
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		uc := newTicketUseCase()
		detail, _, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{Assignee: entity.SetAssignee(adminID)},
		)
		require.NoError(t, err)
		require.NotNil(t, detail.Assignee)
		assert.Equal(t, adminID, detail.Assignee.ID)
	})

	t.Run("正常系: 優先度を変更できること", func(t *testing.T) {
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
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		uc := newTicketUseCase()
		priority := entity.PriorityHigh
		detail, _, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{Priority: &priority},
		)
		require.NoError(t, err)
		assert.Equal(t, entity.PriorityHigh, detail.Priority)
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
		formID, defaultStatusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		uc := newTicketUseCase()
		priority := entity.PriorityHigh
		_, _, err := uc.UpdateTicket(
			ctx,
			ticketID,
			outsiderID,
			usecase.UpdateTicketInput{Priority: &priority},
		)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})

	t.Run("準正常系: 無効な優先度で VALIDATION_ERROR になること", func(t *testing.T) {
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
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		uc := newTicketUseCase()
		invalid := entity.Priority("invalid")
		_, _, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{Priority: &invalid},
		)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeValidation, appErr.Code)
	})
}

type recordingEventPublisher struct {
	mu     sync.Mutex
	record []usecase.TicketEvent
}

func (p *recordingEventPublisher) PublishTicketUpdated(
	_ context.Context,
	event usecase.TicketEvent,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.record = append(p.record, event)
	return nil
}

func (p *recordingEventPublisher) events() []usecase.TicketEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]usecase.TicketEvent, len(p.record))
	copy(out, p.record)
	return out
}

func TestTicketUseCase_ListTicketHistories(t *testing.T) {
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
		formID, defaultStatusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, defaultStatusID, "resp-1")

		uc := newTicketUseCase()
		_, err := uc.ListTicketHistories(ctx, ticketID, outsiderID)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
}
