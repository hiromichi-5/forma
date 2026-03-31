package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
)

type InviteUseCase interface {
	CreateInvite(
		ctx context.Context,
		formID, userID uuid.UUID,
		email, role string,
	) (entity.Invite, error)
	ListInvites(ctx context.Context, formID, userID uuid.UUID) ([]entity.Invite, error)
	DeleteInvite(ctx context.Context, formID, userID, inviteID uuid.UUID) error
	AcceptInvite(ctx context.Context, inviteID, userID uuid.UUID) error
}

type InviteHandler struct {
	uc InviteUseCase
}

func NewInviteHandler(uc InviteUseCase) *InviteHandler {
	return &InviteHandler{uc: uc}
}

type createInviteReq struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"  binding:"required,oneof=admin editor"`
}

func (h *InviteHandler) PostV1FormsFormIdInvites(c *gin.Context) {
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

	var req createInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	invite, err := h.uc.CreateInvite(c, formID, userID, req.Email, req.Role)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, createInviteResp{
		InviteID:  invite.ID.String(),
		ExpiresAt: invite.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func (h *InviteHandler) GetV1FormsFormIdInvites(c *gin.Context) {
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

	invites, err := h.uc.ListInvites(c, formID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"invites": toInviteListResp(invites)})
}

func (h *InviteHandler) DeleteV1FormsFormIdInvitesInviteId(c *gin.Context) {
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
	inviteID, err := uuid.Parse(c.Param("invite_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.DeleteInvite(c, formID, userID, inviteID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InviteHandler) PostV1InvitesInviteIdAccept(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}
	inviteID, err := uuid.Parse(c.Param("invite_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.AcceptInvite(c, inviteID, userID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
