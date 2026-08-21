package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/logger"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

const notificationRateLimitWindow = 5 * time.Minute

type NotificationSettings struct {
	EmailCollectionType *entity.EmailCollectionType
	Settings            []entity.NotificationSetting
}

type NotificationSettingInput struct {
	NotificationType entity.NotificationType
	Mode             entity.NotificationMode
	IncludeDetail    bool
}

type NotificationResult struct {
	NotificationType entity.NotificationType
	Sent             bool
}

type NotificationUseCase struct {
	notificationRepo repository.NotificationRepository
	ticketRepo       repository.TicketRepository
	formRepo         repository.FormRepository
	statusRepo       repository.StatusRepository
	memberRepo       repository.MemberRepository
	userRepo         repository.UserRepository
	uow              repository.UnitOfWork[repository.NotificationRepos]
	emailSender      repository.EmailSender
	now              func() time.Time
}

func NewNotificationUseCase(
	notificationRepo repository.NotificationRepository,
	ticketRepo repository.TicketRepository,
	formRepo repository.FormRepository,
	statusRepo repository.StatusRepository,
	memberRepo repository.MemberRepository,
	userRepo repository.UserRepository,
	uow repository.UnitOfWork[repository.NotificationRepos],
	emailSender repository.EmailSender,
) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
		ticketRepo:       ticketRepo,
		formRepo:         formRepo,
		statusRepo:       statusRepo,
		memberRepo:       memberRepo,
		userRepo:         userRepo,
		uow:              uow,
		emailSender:      emailSender,
		now:              time.Now,
	}
}

func (uc *NotificationUseCase) GetSettings(
	ctx context.Context,
	formID, userID uuid.UUID,
) (NotificationSettings, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return NotificationSettings{}, err
	}

	form, err := uc.formRepo.GetByID(ctx, formID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return NotificationSettings{}, entity.NewError(entity.CodeFormNotFound)
		}
		return NotificationSettings{}, err
	}

	stored, err := uc.notificationRepo.ListSettings(ctx, formID)
	if err != nil {
		return NotificationSettings{}, err
	}

	return NotificationSettings{
		EmailCollectionType: form.EmailCollectionType,
		Settings:            mergeSettings(formID, stored),
	}, nil
}

func (uc *NotificationUseCase) UpdateSettings(
	ctx context.Context,
	formID, userID uuid.UUID,
	inputs []NotificationSettingInput,
) (NotificationSettings, error) {
	if err := requireAdmin(ctx, uc.memberRepo, formID, userID); err != nil {
		return NotificationSettings{}, err
	}

	if err := validateSettingInputs(inputs); err != nil {
		return NotificationSettings{}, err
	}

	if _, err := uc.formRepo.GetByID(ctx, formID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return NotificationSettings{}, entity.NewError(entity.CodeFormNotFound)
		}
		return NotificationSettings{}, err
	}

	if err := uc.uow.Do(ctx, func(repos repository.NotificationRepos) error {
		for _, in := range inputs {
			if _, err := repos.Notification.UpsertSetting(ctx, entity.NotificationSetting{
				FormID:           formID,
				NotificationType: in.NotificationType,
				Mode:             in.Mode,
				IncludeDetail:    in.IncludeDetail,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return NotificationSettings{}, err
	}

	logger.From(ctx).Info("notification settings updated",
		"form_id", formID.String(),
	)

	return uc.GetSettings(ctx, formID, userID)
}

func (uc *NotificationUseCase) SendNotification(
	ctx context.Context,
	ticketID, userID uuid.UUID,
	notificationType entity.NotificationType,
) (entity.TicketNotification, error) {
	if !notificationType.Valid() {
		return entity.TicketNotification{}, entity.NewError(entity.CodeValidation)
	}

	ticket, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.TicketNotification{}, entity.NewError(entity.CodeResourceHidden)
		}
		return entity.TicketNotification{}, err
	}

	if err := requireEditor(ctx, uc.memberRepo, ticket.FormID, userID); err != nil {
		return entity.TicketNotification{}, err
	}

	setting, err := uc.resolveSetting(ctx, ticket.FormID, notificationType)
	if err != nil {
		return entity.TicketNotification{}, err
	}
	if setting.Mode == entity.NotificationModeOff {
		return entity.TicketNotification{}, entity.NewError(entity.CodeNotificationDisabled)
	}

	if err := uc.ensureNotRateLimited(ctx, ticketID, notificationType); err != nil {
		return entity.TicketNotification{}, err
	}

	return uc.sendAndRecord(ctx, ticket, setting, userID)
}

func (uc *NotificationUseCase) NotifyTicketUpdated(
	ctx context.Context,
	ticketID, userID uuid.UUID,
	notificationTypes []entity.NotificationType,
) []NotificationResult {
	if len(notificationTypes) == 0 {
		return nil
	}

	log := logger.From(ctx)

	ticket, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		log.Error("notification skipped: ticket load failed",
			"ticket_id", ticketID.String(), "error", err)
		return nil
	}

	results := make([]NotificationResult, 0, len(notificationTypes))
	for _, t := range notificationTypes {
		setting, err := uc.resolveSetting(ctx, ticket.FormID, t)
		if err != nil {
			log.Error("notification skipped: setting load failed",
				"ticket_id", ticketID.String(), "notification_type", t, "error", err)
			continue
		}
		if setting.Mode != entity.NotificationModeAlways {
			continue
		}

		if _, err := uc.sendAndRecord(ctx, ticket, setting, userID); err != nil {
			// 回答者のメールアドレスがない場合は送信対象外であり、失敗として扱わない。
			var domainErr *entity.Error
			if errors.As(err, &domainErr) &&
				domainErr.Code == entity.CodeRespondentEmailMissing {
				continue
			}
			log.Error("notification send failed",
				"ticket_id", ticketID.String(), "notification_type", t, "error", err)
			results = append(results, NotificationResult{NotificationType: t, Sent: false})
			continue
		}
		results = append(results, NotificationResult{NotificationType: t, Sent: true})
	}
	return results
}

func (uc *NotificationUseCase) ListLatestSent(
	ctx context.Context,
	ticketID uuid.UUID,
) ([]entity.TicketNotification, error) {
	return uc.notificationRepo.ListLatestSent(ctx, ticketID)
}

func (uc *NotificationUseCase) sendAndRecord(
	ctx context.Context,
	ticket entity.Ticket,
	setting entity.NotificationSetting,
	userID uuid.UUID,
) (entity.TicketNotification, error) {
	if ticket.RespondentEmail == nil {
		return entity.TicketNotification{}, entity.NewError(entity.CodeRespondentEmailMissing)
	}

	form, err := uc.formRepo.GetByID(ctx, ticket.FormID)
	if err != nil {
		return entity.TicketNotification{}, err
	}

	template, data, err := uc.buildMail(ctx, ticket, form, setting)
	if err != nil {
		return entity.TicketNotification{}, err
	}

	if err := uc.emailSender.SendEmail(ctx, repository.SendEmailInput{
		To:           []string{*ticket.RespondentEmail},
		TemplateName: template,
		TemplateData: data,
	}); err != nil {
		return entity.TicketNotification{}, err
	}

	sent, err := uc.notificationRepo.CreateSent(
		ctx,
		ticket.ID,
		setting.NotificationType,
		&userID,
		uc.now(),
	)
	if err != nil {
		return entity.TicketNotification{}, err
	}

	logger.From(ctx).Info("notification sent",
		"ticket_id", ticket.ID.String(),
		"notification_type", setting.NotificationType,
	)

	return sent, nil
}

func (uc *NotificationUseCase) buildMail(
	ctx context.Context,
	ticket entity.Ticket,
	form entity.Form,
	setting entity.NotificationSetting,
) (template string, data map[string]string, err error) {
	data = map[string]string{"form_title": form.Title}

	switch setting.NotificationType {
	case entity.NotificationTypeStatusChange:
		if !setting.IncludeDetail {
			return repository.TemplateTicketStatusChanged, data, nil
		}
		status, err := uc.statusRepo.GetByID(ctx, ticket.StatusID)
		if err != nil {
			return "", nil, err
		}
		data["status_name"] = status.Name
		return repository.TemplateTicketStatusChangedDetailed, data, nil

	case entity.NotificationTypeAssigneeAssigned:
		if ticket.AssigneeID == nil {
			return "", nil, entity.NewError(entity.CodeValidation)
		}
		if !setting.IncludeDetail {
			return repository.TemplateTicketAssigned, data, nil
		}
		name, err := getUserDisplayName(ctx, uc.userRepo, *ticket.AssigneeID)
		if err != nil {
			return "", nil, err
		}
		data["assignee_name"] = name
		return repository.TemplateTicketAssignedDetailed, data, nil

	default:
		return "", nil, entity.NewError(entity.CodeValidation)
	}
}

func (uc *NotificationUseCase) resolveSetting(
	ctx context.Context,
	formID uuid.UUID,
	notificationType entity.NotificationType,
) (entity.NotificationSetting, error) {
	setting, err := uc.notificationRepo.GetSetting(ctx, formID, notificationType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return defaultSetting(formID, notificationType), nil
		}
		return entity.NotificationSetting{}, err
	}
	return setting, nil
}

func (uc *NotificationUseCase) ensureNotRateLimited(
	ctx context.Context,
	ticketID uuid.UUID,
	notificationType entity.NotificationType,
) error {
	latest, err := uc.notificationRepo.GetLatestSent(ctx, ticketID, notificationType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}
	if uc.now().Sub(latest.SentAt) < notificationRateLimitWindow {
		return entity.NewError(entity.CodeNotificationRateLimited)
	}
	return nil
}

func validateSettingInputs(inputs []NotificationSettingInput) error {
	if len(inputs) == 0 {
		return entity.NewError(entity.CodeValidation)
	}
	seen := make(map[entity.NotificationType]struct{}, len(inputs))
	for _, in := range inputs {
		if !in.NotificationType.Valid() || !in.Mode.Valid() {
			return entity.NewError(entity.CodeValidation)
		}
		if _, ok := seen[in.NotificationType]; ok {
			return entity.NewError(entity.CodeValidation)
		}
		seen[in.NotificationType] = struct{}{}
	}
	return nil
}

func mergeSettings(
	formID uuid.UUID,
	stored []entity.NotificationSetting,
) []entity.NotificationSetting {
	byType := make(map[entity.NotificationType]entity.NotificationSetting, len(stored))
	for _, s := range stored {
		byType[s.NotificationType] = s
	}

	types := entity.NotificationTypes()
	settings := make([]entity.NotificationSetting, 0, len(types))
	for _, t := range types {
		if s, ok := byType[t]; ok {
			settings = append(settings, s)
			continue
		}
		settings = append(settings, defaultSetting(formID, t))
	}
	return settings
}

func defaultSetting(
	formID uuid.UUID,
	notificationType entity.NotificationType,
) entity.NotificationSetting {
	return entity.NotificationSetting{
		FormID:           formID,
		NotificationType: notificationType,
		Mode:             entity.NotificationModeOff,
		IncludeDetail:    false,
	}
}
