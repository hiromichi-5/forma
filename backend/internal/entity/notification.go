package entity

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeStatusChange     NotificationType = "status_change"
	NotificationTypeAssigneeAssigned NotificationType = "assignee_assigned"
)

type NotificationMode string

const (
	NotificationModeAlways  NotificationMode = "always"
	NotificationModeConfirm NotificationMode = "confirm"
	NotificationModeOff     NotificationMode = "off"
)

type NotificationSetting struct {
	FormID           uuid.UUID
	NotificationType NotificationType
	Mode             NotificationMode
	IncludeDetail    bool
	UpdatedAt        time.Time
}

type TicketNotification struct {
	ID               uuid.UUID
	TicketID         uuid.UUID
	NotificationType NotificationType
	SentBy           *uuid.UUID
	SentAt           time.Time
}

var notificationTypes = []NotificationType{
	NotificationTypeStatusChange,
	NotificationTypeAssigneeAssigned,
}

func NotificationTypes() []NotificationType {
	return notificationTypes
}

// Valid は t が既知の通知種別かどうかを返す。
func (t NotificationType) Valid() bool {
	return slices.Contains(notificationTypes, t)
}

func (m NotificationMode) Valid() bool {
	switch m {
	case NotificationModeAlways, NotificationModeConfirm, NotificationModeOff:
		return true
	default:
		return false
	}
}
