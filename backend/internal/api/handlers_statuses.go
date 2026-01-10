package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type statusesService interface {
	ListFormStatuses(ctx context.Context, formID string, actor uuid.UUID) ([]service.FormStatus, error)
	CreateFormStatus(ctx context.Context, formID, name string, color *string, displayOrder int32, isDefault bool, actor uuid.UUID) (service.FormStatus, error)
	UpdateFormStatus(ctx context.Context, formID, statusID string, name, color *string, displayOrder *int32, actor uuid.UUID) (service.FormStatus, error)
	SetDefaultFormStatus(ctx context.Context, formID, statusID string, actor uuid.UUID) (service.FormStatus, error)
	DeleteFormStatus(ctx context.Context, formID, statusID string, actor uuid.UUID) error
}

type StatusesHandler struct{ Svc statusesService }

type createStatusReq struct {
	Name         string  `json:"name" binding:"required"`
	Color        *string `json:"color"`
	DisplayOrder int32   `json:"display_order" binding:"required"`
	IsDefault    bool    `json:"is_default"`
}

type updateStatusReq struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	DisplayOrder *int32  `json:"display_order"`
}

func (h *StatusesHandler) GetV1FormsIdStatuses(c *gin.Context, formID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	statuses, err := h.Svc.ListFormStatuses(c, formID, uid)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}

func (h *StatusesHandler) PostV1FormsIdStatuses(c *gin.Context, formID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	var req createStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}

	status, err := h.Svc.CreateFormStatus(c, formID, req.Name, req.Color, req.DisplayOrder, req.IsDefault, uid)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(http.StatusCreated, status)
}

func (h *StatusesHandler) PatchV1FormsIdStatusesStatusId(c *gin.Context, formID, statusID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}

	status, err := h.Svc.UpdateFormStatus(c, formID, statusID, req.Name, req.Color, req.DisplayOrder, uid)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *StatusesHandler) PostV1FormsIdStatusesStatusIdDefault(c *gin.Context, formID, statusID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	status, err := h.Svc.SetDefaultFormStatus(c, formID, statusID, uid)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *StatusesHandler) DeleteV1FormsIdStatusesStatusId(c *gin.Context, formID, statusID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	if err := h.Svc.DeleteFormStatus(c, formID, statusID, uid); err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		case service.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{"code": "CONFLICT"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
