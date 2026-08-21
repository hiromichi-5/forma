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

type StatusUseCase interface {
	ListStatuses(ctx context.Context, formID, userID uuid.UUID) ([]entity.FormStatus, error)
	CreateStatus(
		ctx context.Context,
		formID, userID uuid.UUID,
		in usecase.CreateStatusInput,
	) (entity.FormStatus, error)
	UpdateStatus(
		ctx context.Context,
		formID, userID, statusID uuid.UUID,
		in usecase.UpdateStatusInput,
	) (entity.FormStatus, error)
	DeleteStatus(ctx context.Context, formID, userID, statusID uuid.UUID) error
}

type StatusHandler struct {
	uc StatusUseCase
}

func NewStatusHandler(uc StatusUseCase) *StatusHandler {
	return &StatusHandler{uc: uc}
}

type createStatusReq struct {
	Name         string  `json:"name"          binding:"required"`
	Color        *string `json:"color"`
	DisplayOrder int32   `json:"display_order" binding:"required"`
	IsDefault    bool    `json:"is_default"`
}

type updateStatusReq struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	DisplayOrder *int32  `json:"display_order"`
	IsDefault    *bool   `json:"is_default"`
}

func (h *StatusHandler) GetV1FormsIdStatuses(c *gin.Context) {
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

	statuses, err := h.uc.ListStatuses(c, formID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"statuses": toStatusListResp(statuses)})
}

func (h *StatusHandler) PostV1FormsIdStatuses(c *gin.Context) {
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

	var req createStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	status, err := h.uc.CreateStatus(c, formID, userID, usecase.CreateStatusInput{
		Name:         req.Name,
		Color:        req.Color,
		DisplayOrder: req.DisplayOrder,
		IsDefault:    req.IsDefault,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toStatusResp(status))
}

func (h *StatusHandler) PatchV1FormsIdStatusesStatusId(c *gin.Context) {
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
	statusID, err := uuid.Parse(c.Param("status_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	status, err := h.uc.UpdateStatus(c, formID, userID, statusID, usecase.UpdateStatusInput{
		Name:         req.Name,
		Color:        req.Color,
		DisplayOrder: req.DisplayOrder,
		IsDefault:    req.IsDefault,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStatusResp(status))
}

func (h *StatusHandler) DeleteV1FormsIdStatusesStatusId(c *gin.Context) {
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
	statusID, err := uuid.Parse(c.Param("status_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.DeleteStatus(c, formID, userID, statusID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
