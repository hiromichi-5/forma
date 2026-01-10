package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type AuthHandler struct {
	Svc interface {
		Authenticate(ctx context.Context, email, password string) (string, error)
		Signup(ctx context.Context, email, password, displayName string) (string, error)
		Logout(ctx context.Context, sessionID string) error
		VerifyEmail(ctx context.Context, token string) error
		ResendEmailVerification(ctx context.Context, email string) error
		RequestPasswordReset(ctx context.Context, email string) error
		ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
	}
	Cookie AuthCookieConfig
}

type AuthCookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResp struct {
	SessionID string `json:"session_id"`
}

func (h *AuthHandler) PostV1AuthLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}
	sid, err := h.Svc.Authenticate(c, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"code": "AUTH_INVALID_CREDENTIALS", "message": "invalid credentials"})
			return
		case service.ErrEmailNotVerified:
			c.JSON(http.StatusForbidden, gin.H{"code": "EMAIL_NOT_VERIFIED"})
			return
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	h.setAuthCookie(c, sid)
	c.JSON(http.StatusOK, loginResp{SessionID: sid})
}

type signupReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1"`
}

type signupResp struct {
	ID string `json:"id"`
}

func (h *AuthHandler) PostV1AuthSignup(c *gin.Context) {
	var req signupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}
	uid, err := h.Svc.Signup(c, req.Email, req.Password, req.DisplayName)
	if err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		case service.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{"code": "CONFLICT", "message": "email already exists"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	if _, err := uuid.Parse(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}
	c.JSON(http.StatusCreated, signupResp{ID: uid})
}

func (h *AuthHandler) PostV1AuthLogout(c *gin.Context) {
	sid, ok := h.sessionIDFromCookie(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	if err := h.Svc.Logout(c, sid); err != nil {
		switch err {
		case service.ErrValidation, service.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	h.clearAuthCookie(c)
	c.Status(http.StatusNoContent)
}

type verifyEmailReq struct {
	Token string `json:"token" binding:"required"`
}

func (h *AuthHandler) PostV1AuthVerifyEmail(c *gin.Context) {
	var req verifyEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}
	if err := h.Svc.VerifyEmail(c, req.Token); err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		case service.ErrTokenNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": "TOKEN_NOT_FOUND"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	c.Status(http.StatusNoContent)
}

type resendVerificationReq struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) PostV1AuthVerifyEmailResend(c *gin.Context) {
	var req resendVerificationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}
	if err := h.Svc.ResendEmailVerification(c, req.Email); err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	c.Status(http.StatusAccepted)
}

type passwordResetReq struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) PostV1AuthPasswordReset(c *gin.Context) {
	var req passwordResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}
	if err := h.Svc.RequestPasswordReset(c, req.Email); err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	c.Status(http.StatusAccepted)
}

type passwordResetConfirmReq struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) PostV1AuthPasswordResetConfirm(c *gin.Context) {
	var req passwordResetConfirmReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "bad request"})
		return
	}
	if err := h.Svc.ConfirmPasswordReset(c, req.Token, req.NewPassword); err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		case service.ErrTokenNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": "TOKEN_NOT_FOUND"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) cookieDefaults() (string, string, http.SameSite) {
	name := h.Cookie.Name
	if name == "" {
		name = "forma_token"
	}
	path := h.Cookie.Path
	if path == "" {
		path = "/"
	}
	sameSite := h.Cookie.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}
	return name, path, sameSite
}

func (h *AuthHandler) setAuthCookie(c *gin.Context, sessionID string) {
	name, path, sameSite := h.cookieDefaults()
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    sessionID,
		Path:     path,
		Domain:   h.Cookie.Domain,
		Secure:   h.Cookie.Secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	name, path, sameSite := h.cookieDefaults()
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   h.Cookie.Domain,
		Secure:   h.Cookie.Secure,
		HttpOnly: true,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}

func (h *AuthHandler) sessionIDFromCookie(c *gin.Context) (string, bool) {
	name, _, _ := h.cookieDefaults()
	cookie, err := c.Request.Cookie(name)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	if _, err := uuid.Parse(cookie.Value); err != nil {
		return "", false
	}
	return cookie.Value, true
}
