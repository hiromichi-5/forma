package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type NotificationRepository interface {
	ListSettings(ctx context.Context, formID uuid.UUID) ([]entity.NotificationSetting, error)
	GetSetting(
		ctx context.Context,
		formID uuid.UUID,
		notificationType entity.NotificationType,
	) (entity.NotificationSetting, error)
	UpsertSetting(
		ctx context.Context,
		setting entity.NotificationSetting,
	) (entity.NotificationSetting, error)
	CreateSent(
		ctx context.Context,
		ticketID uuid.UUID,
		notificationType entity.NotificationType,
		sentBy *uuid.UUID,
		sentAt time.Time,
	) (entity.TicketNotification, error)
	GetLatestSent(
		ctx context.Context,
		ticketID uuid.UUID,
		notificationType entity.NotificationType,
	) (entity.TicketNotification, error)
	ListLatestSent(ctx context.Context, ticketID uuid.UUID) ([]entity.TicketNotification, error)
}
