package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type MembersHandler struct {
	Svc interface {
		RequireEditor(ctx context.Context, formID string, actor uuid.UUID) error
		RequireAdmin(ctx context.Context, formID string, actor uuid.UUID) error
		ListMembers(ctx context.Context, formID string) ([]service.Member, error)
		AddMember(ctx context.Context, formID, email, role string) error
		ChangeRole(ctx context.Context, formID, userID, role string) error
		RemoveMember(ctx context.Context, formID, userID string) error
	}
}

type memberAddReq struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"  binding:"required,oneof=admin editor"`
}
type memberRoleUpdateReq struct {
	Role string `json:"role" binding:"required,oneof=admin editor"`
}

func (h *MembersHandler) GetV1FormsFormIdMembers(c *gin.Context) {
	formID := c.Param("form_id")
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
	if err := h.Svc.RequireEditor(c, formID, uid); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	ms, err := h.Svc.ListMembers(c, formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": ms})
}

func (h *MembersHandler) PostV1FormsFormIdMembers(c *gin.Context) {
	formID := c.Param("form_id")
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
	if err := h.Svc.RequireAdmin(c, formID, uid); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	var req memberAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	if err := h.Svc.AddMember(c, formID, req.Email, req.Role); err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.Status(http.StatusCreated)
}

func (h *MembersHandler) PutV1FormsFormIdMembersUserId(c *gin.Context) {
	formID := c.Param("form_id")
	userID := c.Param("user_id")
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
	if err := h.Svc.RequireAdmin(c, formID, uid); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	var req memberRoleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	if err := h.Svc.ChangeRole(c, formID, userUUID.String(), req.Role); err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MembersHandler) DeleteV1FormsFormIdMembersUserId(c *gin.Context) {
	formID := c.Param("form_id")
	userID := c.Param("user_id")
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
	if err := h.Svc.RequireAdmin(c, formID, uid); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	if err := h.Svc.RemoveMember(c, formID, userUUID.String()); err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
