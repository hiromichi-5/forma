package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (entity.User, error)
	UpdateDisplayName(
		ctx context.Context,
		userID uuid.UUID,
		displayName string,
	) (entity.User, error)
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
}

type ProfileHandler struct {
	uc ProfileUseCase
}

func NewProfileHandler(uc ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{uc: uc}
}

type updateDisplayNameReq struct {
	DisplayName string `json:"display_name" binding:"required"`
}

type changePasswordReq struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}

func (h *ProfileHandler) GetV1Me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	user, err := h.uc.GetProfile(c, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserProfileResp(user))
}

func (h *ProfileHandler) PatchV1Me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	var req updateDisplayNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	user, err := h.uc.UpdateDisplayName(c, userID, req.DisplayName)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserProfileResp(user))
}

func (h *ProfileHandler) DeleteV1Me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	if err := h.uc.DeleteProfile(c, userID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProfileHandler) PatchV1MePassword(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.ChangePassword(c, userID, req.CurrentPassword, req.NewPassword); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
