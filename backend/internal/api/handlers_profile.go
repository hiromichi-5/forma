package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type ProfileHandler struct {
	Svc interface {
		GetProfile(ctx context.Context, userID string) (db.User, error)
		UpdateDisplayName(ctx context.Context, userID, displayName string) (db.User, error)
		DeleteProfile(ctx context.Context, userID string) error
	}
}

type updateDisplayNameReq struct {
	DisplayName string `json:"display_name" binding:"required"`
}

type userProfileResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (h *ProfileHandler) GetV1Me(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}

	user, err := h.Svc.GetProfile(c, userID)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": "USER_NOT_FOUND"})
			return
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}

	resp := userProfileResp{
		ID:          userID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProfileHandler) PutV1Me(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}

	var req updateDisplayNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}

	user, err := h.Svc.UpdateDisplayName(c, userID, req.DisplayName)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": "USER_NOT_FOUND"})
			return
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}

	resp := userProfileResp{
		ID:          userID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProfileHandler) DeleteV1Me(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}

	err := h.Svc.DeleteProfile(c, userID)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": "USER_NOT_FOUND"})
			return
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}

	c.Status(http.StatusNoContent)
}
