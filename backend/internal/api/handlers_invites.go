package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type inviteIssueResponse struct {
	Code string `json:"code"`
}

type inviteAcceptRequest struct {
	Code string `json:"code" binding:"required"`
}

func toAPIInvite(inv db.FormInvite) (FormInvite, bool) {
	if !inv.CreatedAt.Valid || !inv.ExpiresAt.Valid || !inv.CreatedBy.Valid {
		return FormInvite{}, false
	}
	return FormInvite{
		Code:      inv.Code,
		FormId:    inv.FormID,
		Role:      FormInviteRole(inv.Role),
		Revoked:   inv.Revoked,
		ExpiresAt: inv.ExpiresAt.Time,
		CreatedAt: inv.CreatedAt.Time,
		CreatedBy: openapi_types.UUID(inv.CreatedBy.Bytes),
	}, true
}

func (h *FormsHandler) PostV1FormsFormIdInvites(c *gin.Context, formID string) {
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

	invite, err := h.S.CreateInvite(c, formID, actor)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		case service.ErrCodeGeneration:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.JSON(http.StatusCreated, inviteIssueResponse{Code: invite.Code})
}

func (h *FormsHandler) GetV1FormsFormIdInvites(c *gin.Context, formID string) {
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

	invites, err := h.S.ListInvites(c, formID, actor)
	if err != nil {
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}

	out := make([]FormInvite, 0, len(invites))
	for _, inv := range invites {
		apiInv, ok := toAPIInvite(inv)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
		out = append(out, apiInv)
	}

	c.JSON(http.StatusOK, ListFormInvitesResponse{Invites: out})
}

func (h *FormsHandler) DeleteV1FormsFormIdInvitesCode(c *gin.Context, formID, code string) {
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

	_, err = h.S.RevokeInvite(c, formID, code, actor)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		case service.ErrInviteNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": "INVITE_NOT_FOUND"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *FormsHandler) PostV1InvitesAccept(c *gin.Context) {
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

	var req inviteAcceptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}

	err = h.S.AcceptInvite(c, req.Code, actor)
	if err != nil {
		switch err {
		case service.ErrInviteNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": "INVITE_NOT_FOUND"})
		case service.ErrInviteExpired, service.ErrInviteRevoked:
			c.JSON(http.StatusGone, gin.H{"code": "INVITE_EXPIRED"})
		case service.ErrAlreadyMember:
			c.JSON(http.StatusConflict, gin.H{"code": "ALREADY_MEMBER"})
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
