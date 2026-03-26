package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type InvitesHandler struct {
	Svc interface {
		CreateInvite(ctx context.Context, formID, email, role string, actor uuid.UUID) (db.FormInvite, error)
		ListInvites(ctx context.Context, formID string, actor uuid.UUID) ([]db.FormInvite, error)
		DeleteInvite(ctx context.Context, formID, inviteID string, actor uuid.UUID) error
		AcceptInvite(ctx context.Context, inviteID string, actor uuid.UUID) error
	}
}

type createInviteReq struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"  binding:"required,oneof=admin editor"`
}

type createInviteResp struct {
	InviteID  string `json:"invite_id"`
	ExpiresAt string `json:"expires_at"`
}

type inviteResp struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	InvitedBy string `json:"invited_by"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

func (h *InvitesHandler) PostV1FormsFormIdInvites(c *gin.Context) {
	formID := c.Param("form_id")
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	actor, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	var req createInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}

	invite, err := h.Svc.CreateInvite(c, formID, req.Email, req.Role, actor)
	if err != nil {
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
	if !invite.ExpiresAt.Valid {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}

	c.JSON(http.StatusCreated, createInviteResp{
		InviteID:  uuid.UUID(invite.ID.Bytes).String(),
		ExpiresAt: invite.ExpiresAt.Time.UTC().Format(time.RFC3339),
	})
}

func (h *InvitesHandler) GetV1FormsFormIdInvites(c *gin.Context) {
	formID := c.Param("form_id")
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	actor, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	invites, err := h.Svc.ListInvites(c, formID, actor)
	if err != nil {
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

	out := make([]inviteResp, 0, len(invites))
	for _, inv := range invites {
		if !inv.ExpiresAt.Valid || !inv.CreatedAt.Valid {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
		out = append(out, inviteResp{
			ID:        uuid.UUID(inv.ID.Bytes).String(),
			Email:     inv.Email,
			Role:      string(inv.Role),
			InvitedBy: uuid.UUID(inv.InvitedBy.Bytes).String(),
			ExpiresAt: inv.ExpiresAt.Time.UTC().Format(time.RFC3339),
			CreatedAt: inv.CreatedAt.Time.UTC().Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"invites": out})
}

func (h *InvitesHandler) DeleteV1FormsFormIdInvitesInviteId(c *gin.Context) {
	formID := c.Param("form_id")
	inviteID := c.Param("invite_id")
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	actor, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	if err := h.Svc.DeleteInvite(c, formID, inviteID, actor); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden),
			errors.Is(err, service.ErrInviteNotFound),
			errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *InvitesHandler) PostV1InvitesInviteIdAccept(c *gin.Context, inviteID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	actor, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	err = h.Svc.AcceptInvite(c, inviteID, actor)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInviteNotFound),
			errors.Is(err, service.ErrInviteExpired),
			errors.Is(err, service.ErrForbidden),
			errors.Is(err, service.ErrUserNotFound),
			errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
