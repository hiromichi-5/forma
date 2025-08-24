package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type AuthHandler struct {
	Svc interface {
		Authenticate(ctx context.Context, email, password string) (string, error)
	}
	JWT auth.Signer
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) PostV1AuthLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}
	uid, err := h.Svc.Authenticate(c, req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "AUTH_INVALID_CREDENTIALS", "message": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}
	tok, err := h.JWT.Issue(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tok})
}
