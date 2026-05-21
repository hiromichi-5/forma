package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"

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
		tickets, err := uc.ListTickets(ctx, formID, adminID, nil)
		require.NoError(t, err)
		assert.Len(t, tickets, 2)
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
		tickets, err := uc.ListTickets(ctx, formID, adminID, &defaultStatusID)
		require.NoError(t, err)
		assert.Len(t, tickets, 1)
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
		_, err := uc.ListTickets(ctx, formID, outsiderID, nil)
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeResourceHidden, appErr.Code)
	})
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
		_, err = uc.UpdateTicket(ctx, ticketID, adminID, &newStatusID, nil, false, nil)
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
		invalid := "invalid"
		_, err := uc.UpdateTicket(ctx, ticketID, adminID, nil, nil, false, &invalid)
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
		detail, err := uc.UpdateTicket(ctx, ticketID, adminID, &newStatusID, nil, false, nil)
		require.NoError(t, err)
		assert.Equal(t, newStatusID, detail.Status.ID)

		// 履歴が記録されていることを確認
		histories, err := uc.ListTicketHistories(ctx, ticketID, adminID)
		require.NoError(t, err)
		assert.Len(t, histories, 1)
		assert.Equal(t, "status", histories[0].FieldName)
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
		detail, err := uc.UpdateTicket(ctx, ticketID, adminID, nil, &adminID, false, nil)
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
		detail, err := uc.UpdateTicket(ctx, ticketID, adminID, nil, nil, false, &priority)
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
		_, err := uc.UpdateTicket(ctx, ticketID, outsiderID, nil, nil, false, &priority)
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
		invalid := "invalid"
		_, err := uc.UpdateTicket(ctx, ticketID, adminID, nil, nil, false, &invalid)
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

func (p *recordingEventPublisher) PublishTicketUpdated(_ context.Context, event usecase.TicketEvent) error {
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
