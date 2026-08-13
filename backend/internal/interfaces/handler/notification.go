package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
)

type NotificationUseCase interface {
	GetSettings(ctx context.Context, formID, userID uuid.UUID) (usecase.NotificationSettings, error)
	UpdateSettings(
		ctx context.Context,
		formID, userID uuid.UUID,
		inputs []usecase.NotificationSettingInput,
	) (usecase.NotificationSettings, error)
	SendNotification(
		ctx context.Context,
		ticketID, userID uuid.UUID,
		notificationType entity.NotificationType,
	) (entity.TicketNotification, error)
}

type NotificationHandler struct {
	uc NotificationUseCase
}

func NewNotificationHandler(uc NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

type updateNotificationSettingsReq struct {
	Settings []notificationSettingReq `json:"settings" binding:"required,min=1,dive"`
}

type notificationSettingReq struct {
	NotificationType string `json:"notification_type" binding:"required"`
	Mode             string `json:"mode"              binding:"required"`
	IncludeDetail    bool   `json:"include_detail"`
}

type sendNotificationReq struct {
	NotificationType string `json:"notification_type" binding:"required"`
}

func (h *NotificationHandler) GetV1FormsFormIdNotificationSettings(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}
	formID, err := uuid.Parse(c.Param("form_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	settings, err := h.uc.GetSettings(c, formID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toNotificationSettingsResp(settings))
}

func (h *NotificationHandler) PatchV1FormsFormIdNotificationSettings(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}
	formID, err := uuid.Parse(c.Param("form_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	var req updateNotificationSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	inputs := make([]usecase.NotificationSettingInput, len(req.Settings))
	for i, s := range req.Settings {
		inputs[i] = usecase.NotificationSettingInput{
			NotificationType: entity.NotificationType(s.NotificationType),
			Mode:             entity.NotificationMode(s.Mode),
			IncludeDetail:    s.IncludeDetail,
		}
	}

	settings, err := h.uc.UpdateSettings(c, formID, userID, inputs)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toNotificationSettingsResp(settings))
}

func (h *NotificationHandler) PostV1TicketsTicketIdNotifications(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}
	ticketID, err := uuid.Parse(c.Param("ticket_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	var req sendNotificationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	sent, err := h.uc.SendNotification(
		c,
		ticketID,
		userID,
		entity.NotificationType(req.NotificationType),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSentNotificationResp(sent))
}
