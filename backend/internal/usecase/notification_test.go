package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingEmailSender struct {
	sent []repository.SendEmailInput
	err  error
}

func (r *recordingEmailSender) SendEmail(
	_ context.Context,
	input repository.SendEmailInput,
) error {
	if r.err != nil {
		return r.err
	}
	r.sent = append(r.sent, input)
	return nil
}

func setNotificationMode(
	t *testing.T,
	ctx context.Context,
	formID, adminID uuid.UUID,
	notificationType entity.NotificationType,
	mode entity.NotificationMode,
	includeDetail bool,
) {
	t.Helper()
	uc := newNotificationUseCase(&mockEmailSender{})
	_, err := uc.UpdateSettings(ctx, formID, adminID, []usecase.NotificationSettingInput{
		{NotificationType: notificationType, Mode: mode, IncludeDetail: includeDetail},
	})
	require.NoError(t, err)
}

func TestNotificationUseCase_GetSettings(t *testing.T) {
	t.Run("正常系: 未設定でも全種別が off として返ること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newNotificationUseCase(&mockEmailSender{})
		settings, err := uc.GetSettings(ctx, formID, adminID)
		require.NoError(t, err)

		require.Len(t, settings.Settings, 2)
		for _, s := range settings.Settings {
			assert.Equal(t, entity.NotificationModeOff, s.Mode)
			assert.False(t, s.IncludeDetail)
		}
		assert.Nil(t, settings.EmailCollectionType)
	})

	t.Run("正常系: Editor も取得できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		editorID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "editor@example.com", "password123", "Editor",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)

		uc := newNotificationUseCase(&mockEmailSender{})
		_, err := uc.GetSettings(ctx, formID, editorID)
		require.NoError(t, err)
	})

	t.Run("準正常系: 非メンバーは RESOURCE_HIDDEN エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		outsiderID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "outsider@example.com", "password123", "Outsider",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newNotificationUseCase(&mockEmailSender{})
		_, err := uc.GetSettings(ctx, formID, outsiderID)

		var domainErr *entity.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, entity.CodeResourceHidden, domainErr.Code)
	})
}

func TestNotificationUseCase_UpdateSettings(t *testing.T) {
	t.Run("正常系: Admin が設定を更新できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newNotificationUseCase(&mockEmailSender{})
		settings, err := uc.UpdateSettings(ctx, formID, adminID, []usecase.NotificationSettingInput{
			{
				NotificationType: entity.NotificationTypeStatusChange,
				Mode:             entity.NotificationModeAlways,
				IncludeDetail:    true,
			},
		})
		require.NoError(t, err)

		byType := make(map[entity.NotificationType]entity.NotificationSetting)
		for _, s := range settings.Settings {
			byType[s.NotificationType] = s
		}
		assert.Equal(t, entity.NotificationModeAlways,
			byType[entity.NotificationTypeStatusChange].Mode)
		assert.True(t, byType[entity.NotificationTypeStatusChange].IncludeDetail)
		// 指定しなかった種別は既定値のまま。
		assert.Equal(t, entity.NotificationModeOff,
			byType[entity.NotificationTypeAssigneeAssigned].Mode)
	})

	t.Run("準正常系: Editor は FORBIDDEN エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		editorID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "editor@example.com", "password123", "Editor",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		testutil.AddMember(t, ctx, testPool, editorID, formID, entity.RoleEditor)

		uc := newNotificationUseCase(&mockEmailSender{})
		_, err := uc.UpdateSettings(ctx, formID, editorID, []usecase.NotificationSettingInput{
			{
				NotificationType: entity.NotificationTypeStatusChange,
				Mode:             entity.NotificationModeAlways,
			},
		})

		var domainErr *entity.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, entity.CodeForbidden, domainErr.Code)
	})

	t.Run("準正常系: 不正なモードは VALIDATION_ERROR になること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newNotificationUseCase(&mockEmailSender{})
		_, err := uc.UpdateSettings(ctx, formID, adminID, []usecase.NotificationSettingInput{
			{NotificationType: entity.NotificationTypeStatusChange, Mode: "sometimes"},
		})

		var domainErr *entity.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, entity.CodeValidation, domainErr.Code)
	})

	t.Run("準正常系: 同一種別の重複指定は VALIDATION_ERROR になること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, _ := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)

		uc := newNotificationUseCase(&mockEmailSender{})
		_, err := uc.UpdateSettings(ctx, formID, adminID, []usecase.NotificationSettingInput{
			{
				NotificationType: entity.NotificationTypeStatusChange,
				Mode:             entity.NotificationModeAlways,
			},
			{
				NotificationType: entity.NotificationTypeStatusChange,
				Mode:             entity.NotificationModeOff,
			},
		})

		var domainErr *entity.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, entity.CodeValidation, domainErr.Code)
	})
}

func TestNotificationUseCase_SendNotification(t *testing.T) {
	t.Run("正常系: confirm モードで回答者へ送信できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeConfirm, false)

		sender := &recordingEmailSender{}
		uc := newNotificationUseCase(sender)
		sent, err := uc.SendNotification(
			ctx, ticketID, adminID, entity.NotificationTypeStatusChange,
		)
		require.NoError(t, err)

		assert.Equal(t, entity.NotificationTypeStatusChange, sent.NotificationType)
		require.Len(t, sender.sent, 1)
		assert.Equal(t, []string{"respondent@example.com"}, sender.sent[0].To)
		assert.Equal(t, repository.TemplateTicketStatusChanged, sender.sent[0].TemplateName)
	})

	t.Run("正常系: include_detail でステータス名を含むテンプレートが使われること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeConfirm, true)

		sender := &recordingEmailSender{}
		uc := newNotificationUseCase(sender)
		_, err := uc.SendNotification(ctx, ticketID, adminID, entity.NotificationTypeStatusChange)
		require.NoError(t, err)

		require.Len(t, sender.sent, 1)
		assert.Equal(t, repository.TemplateTicketStatusChangedDetailed, sender.sent[0].TemplateName)
		assert.Equal(t, "未対応", sender.sent[0].TemplateData["status_name"])
	})

	t.Run("準正常系: off の種別は NOTIFICATION_DISABLED になること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)

		sender := &recordingEmailSender{}
		uc := newNotificationUseCase(sender)
		_, err := uc.SendNotification(ctx, ticketID, adminID, entity.NotificationTypeStatusChange)

		var domainErr *entity.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, entity.CodeNotificationDisabled, domainErr.Code)
		assert.Empty(t, sender.sent)
	})

	t.Run(
		"準正常系: 回答者のメールアドレスがない場合 RESPONDENT_EMAIL_MISSING になること",
		func(t *testing.T) {
			truncate(t)
			ctx := context.Background()
			adminID := testutil.CreateVerifiedUser(
				t, ctx, testPool, "admin@example.com", "password123", "Admin",
			)
			formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
			ticketID := testutil.CreateTicket(t, ctx, testPool, formID, statusID, "resp-1")
			setNotificationMode(t, ctx, formID, adminID,
				entity.NotificationTypeStatusChange, entity.NotificationModeConfirm, false)

			sender := &recordingEmailSender{}
			uc := newNotificationUseCase(sender)
			_, err := uc.SendNotification(
				ctx, ticketID, adminID, entity.NotificationTypeStatusChange,
			)

			var domainErr *entity.Error
			require.True(t, errors.As(err, &domainErr))
			assert.Equal(t, entity.CodeRespondentEmailMissing, domainErr.Code)
			assert.Empty(t, sender.sent)
		},
	)

	t.Run("準正常系: 短時間の再送は NOTIFICATION_RATE_LIMITED になること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeConfirm, false)

		sender := &recordingEmailSender{}
		uc := newNotificationUseCase(sender)
		_, err := uc.SendNotification(ctx, ticketID, adminID, entity.NotificationTypeStatusChange)
		require.NoError(t, err)

		_, err = uc.SendNotification(ctx, ticketID, adminID, entity.NotificationTypeStatusChange)

		var domainErr *entity.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, entity.CodeNotificationRateLimited, domainErr.Code)
		assert.Len(t, sender.sent, 1)
	})

	t.Run("正常系: 送信に失敗した場合は記録されず再試行できること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeConfirm, false)

		sender := &recordingEmailSender{err: errors.New("smtp down")}
		uc := newNotificationUseCase(sender)
		_, err := uc.SendNotification(ctx, ticketID, adminID, entity.NotificationTypeStatusChange)
		require.Error(t, err)

		// 記録が残っていないため、レートリミットに阻まれず再試行できる。
		sender.err = nil
		_, err = uc.SendNotification(ctx, ticketID, adminID, entity.NotificationTypeStatusChange)
		require.NoError(t, err)
		assert.Len(t, sender.sent, 1)
	})

	t.Run("準正常系: 非メンバーは RESOURCE_HIDDEN エラーになること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		outsiderID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "outsider@example.com", "password123", "Outsider",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)

		uc := newNotificationUseCase(&mockEmailSender{})
		_, err := uc.SendNotification(
			ctx, ticketID, outsiderID, entity.NotificationTypeStatusChange,
		)

		var domainErr *entity.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, entity.CodeResourceHidden, domainErr.Code)
	})
}

func TestTicketUseCase_UpdateTicket_Notification(t *testing.T) {
	t.Run("正常系: always ならステータス変更で自動送信されること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeAlways, false)

		newStatusID := otherStatusID(t, ctx, formID, statusID)

		sender := &recordingEmailSender{}
		uc := newTicketUseCaseWith(&noopEventPublisher{}, newNotificationUseCase(sender))
		detail, results, err := uc.UpdateTicket(
			ctx, ticketID, adminID, usecase.UpdateTicketInput{StatusID: &newStatusID},
		)
		require.NoError(t, err)

		require.Len(t, results, 1)
		assert.Equal(t, entity.NotificationTypeStatusChange, results[0].NotificationType)
		assert.True(t, results[0].Sent)
		assert.Len(t, sender.sent, 1)

		// 最終送信日時が詳細に反映されること。
		var lastSentAt *string
		for _, n := range detail.Notifications {
			if n.NotificationType == entity.NotificationTypeStatusChange && n.LastSentAt != nil {
				formatted := n.LastSentAt.String()
				lastSentAt = &formatted
			}
		}
		assert.NotNil(t, lastSentAt)
	})

	t.Run("正常系: confirm では自動送信されないこと", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeConfirm, false)

		newStatusID := otherStatusID(t, ctx, formID, statusID)

		sender := &recordingEmailSender{}
		uc := newTicketUseCaseWith(&noopEventPublisher{}, newNotificationUseCase(sender))
		_, results, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{StatusID: &newStatusID},
		)
		require.NoError(t, err)

		assert.Empty(t, results)
		assert.Empty(t, sender.sent)
	})

	t.Run("正常系: 担当者の解除では送信されないこと", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeAssigneeAssigned, entity.NotificationModeAlways, false)

		sender := &recordingEmailSender{}
		uc := newTicketUseCaseWith(&noopEventPublisher{}, newNotificationUseCase(sender))

		_, results, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{Assignee: usecase.SetAssignee(adminID)},
		)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.True(t, results[0].Sent)

		_, results, err = uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{Assignee: usecase.ClearAssignee()},
		)
		require.NoError(t, err)
		assert.Empty(t, results)
		assert.Len(t, sender.sent, 1)
	})

	t.Run("正常系: 回答者のメールアドレスがない場合は失敗として扱わないこと", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicket(t, ctx, testPool, formID, statusID, "resp-1")
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeAlways, false)

		newStatusID := otherStatusID(t, ctx, formID, statusID)

		sender := &recordingEmailSender{}
		uc := newTicketUseCaseWith(&noopEventPublisher{}, newNotificationUseCase(sender))
		_, results, err := uc.UpdateTicket(
			ctx,
			ticketID,
			adminID,
			usecase.UpdateTicketInput{StatusID: &newStatusID},
		)
		require.NoError(t, err)

		assert.Empty(t, results)
		assert.Empty(t, sender.sent)
	})

	t.Run("正常系: 送信に失敗しても更新は成功し failed が返ること", func(t *testing.T) {
		truncate(t)
		ctx := context.Background()
		adminID := testutil.CreateVerifiedUser(
			t, ctx, testPool, "admin@example.com", "password123", "Admin",
		)
		formID, statusID := testutil.CreateForm(t, ctx, testPool, "gform1", "Form", adminID)
		ticketID := testutil.CreateTicketWithRespondent(
			t, ctx, testPool, formID, statusID, "resp-1", "respondent@example.com",
		)
		setNotificationMode(t, ctx, formID, adminID,
			entity.NotificationTypeStatusChange, entity.NotificationModeAlways, false)

		newStatusID := otherStatusID(t, ctx, formID, statusID)

		sender := &recordingEmailSender{err: errors.New("smtp down")}
		uc := newTicketUseCaseWith(&noopEventPublisher{}, newNotificationUseCase(sender))
		detail, results, err := uc.UpdateTicket(
			ctx, ticketID, adminID, usecase.UpdateTicketInput{StatusID: &newStatusID},
		)
		require.NoError(t, err)

		assert.Equal(t, newStatusID, detail.Status.ID)
		require.Len(t, results, 1)
		assert.False(t, results[0].Sent)
	})
}

// otherStatusID は既定ステータス以外のステータス ID を1件返す。
func otherStatusID(t *testing.T, ctx context.Context, formID, exclude uuid.UUID) uuid.UUID {
	t.Helper()
	statuses, err := newStatusRepo().List(ctx, formID)
	require.NoError(t, err)
	for _, s := range statuses {
		if s.ID != exclude {
			return s.ID
		}
	}
	t.Fatalf("no other status found for form %s", formID)
	return uuid.Nil
}
