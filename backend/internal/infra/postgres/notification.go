package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.NotificationRepository = (*NotificationRepository)(nil)

type NotificationRepository struct {
	q *db.Queries
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{q: db.New(pool)}
}

func (r *NotificationRepository) ListSettings(
	ctx context.Context,
	formID uuid.UUID,
) ([]entity.NotificationSetting, error) {
	rows, err := r.q.ListFormNotificationSettings(ctx, toUUID(formID))
	if err != nil {
		return nil, repoError(err)
	}
	settings := make([]entity.NotificationSetting, 0, len(rows))
	for _, row := range rows {
		settings = append(settings, toNotificationSetting(row))
	}
	return settings, nil
}

func (r *NotificationRepository) GetSetting(
	ctx context.Context,
	formID uuid.UUID,
	notificationType entity.NotificationType,
) (entity.NotificationSetting, error) {
	row, err := r.q.GetFormNotificationSetting(ctx, db.GetFormNotificationSettingParams{
		FormID:           toUUID(formID),
		NotificationType: db.NotificationType(notificationType),
	})
	if err != nil {
		return entity.NotificationSetting{}, repoError(err)
	}
	return toNotificationSetting(row), nil
}

func (r *NotificationRepository) UpsertSetting(
	ctx context.Context,
	setting entity.NotificationSetting,
) (entity.NotificationSetting, error) {
	row, err := r.q.UpsertFormNotificationSetting(ctx, db.UpsertFormNotificationSettingParams{
		FormID:           toUUID(setting.FormID),
		NotificationType: db.NotificationType(setting.NotificationType),
		Mode:             db.NotificationMode(setting.Mode),
		IncludeDetail:    setting.IncludeDetail,
	})
	if err != nil {
		return entity.NotificationSetting{}, repoError(err)
	}
	return toNotificationSetting(row), nil
}

func (r *NotificationRepository) CreateSent(
	ctx context.Context,
	ticketID uuid.UUID,
	notificationType entity.NotificationType,
	sentBy *uuid.UUID,
	sentAt time.Time,
) (entity.TicketNotification, error) {
	row, err := r.q.CreateTicketNotification(ctx, db.CreateTicketNotificationParams{
		ID:               toUUID(uuid.New()),
		TicketID:         toUUID(ticketID),
		NotificationType: db.NotificationType(notificationType),
		SentBy:           toNullUUID(sentBy),
		SentAt:           toTimestamptz(sentAt),
	})
	if err != nil {
		return entity.TicketNotification{}, repoError(err)
	}
	return toTicketNotification(row), nil
}

func (r *NotificationRepository) GetLatestSent(
	ctx context.Context,
	ticketID uuid.UUID,
	notificationType entity.NotificationType,
) (entity.TicketNotification, error) {
	row, err := r.q.GetLatestTicketNotification(ctx, db.GetLatestTicketNotificationParams{
		TicketID:         toUUID(ticketID),
		NotificationType: db.NotificationType(notificationType),
	})
	if err != nil {
		return entity.TicketNotification{}, repoError(err)
	}
	return toTicketNotification(row), nil
}

func (r *NotificationRepository) ListLatestSent(
	ctx context.Context,
	ticketID uuid.UUID,
) ([]entity.TicketNotification, error) {
	rows, err := r.q.ListLatestTicketNotifications(ctx, toUUID(ticketID))
	if err != nil {
		return nil, repoError(err)
	}
	notifications := make([]entity.TicketNotification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, toTicketNotification(row))
	}
	return notifications, nil
}

func toNotificationSetting(row db.FormNotificationSetting) entity.NotificationSetting {
	return entity.NotificationSetting{
		FormID:           fromUUID(row.FormID),
		NotificationType: entity.NotificationType(row.NotificationType),
		Mode:             entity.NotificationMode(row.Mode),
		IncludeDetail:    row.IncludeDetail,
		UpdatedAt:        fromTimestamptz(row.UpdatedAt),
	}
}

func toTicketNotification(row db.TicketNotification) entity.TicketNotification {
	return entity.TicketNotification{
		ID:               fromUUID(row.ID),
		TicketID:         fromUUID(row.TicketID),
		NotificationType: entity.NotificationType(row.NotificationType),
		SentBy:           fromNullUUID(row.SentBy),
		SentAt:           fromTimestamptz(row.SentAt),
	}
}
