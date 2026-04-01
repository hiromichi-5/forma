package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

type AuthUseCase interface {
	Authenticate(ctx context.Context, email, password string) (entity.Session, error)
	Signup(ctx context.Context, email, password, displayName string) (uuid.UUID, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
	ResendEmailVerification(ctx context.Context, email string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

type AuthHandler struct {
	uc     AuthUseCase
	cookie CookieConfig
}

func NewAuthHandler(uc AuthUseCase, cookie CookieConfig) *AuthHandler {
	return &AuthHandler{uc: uc, cookie: cookie}
}

type loginReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) PostV1AuthLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}
	session, err := h.uc.Authenticate(c, req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}
	h.setAuthCookie(c, session.ID.String())
	c.JSON(http.StatusOK, loginResp{SessionID: session.ID.String()})
}

type signupReq struct {
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1"`
}

func (h *AuthHandler) PostV1AuthSignup(c *gin.Context) {
	var req signupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}
	userID, err := h.uc.Signup(c, req.Email, req.Password, req.DisplayName)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, signupResp{ID: userID.String()})
}

func (h *AuthHandler) PostV1AuthLogout(c *gin.Context) {
	sid, ok := h.sessionIDFromCookie(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}
	if err := h.uc.Logout(c, sid); err != nil {
		handleError(c, err)
		return
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
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}
	if err := h.uc.VerifyEmail(c, req.Token); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type resendVerificationReq struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) PostV1AuthVerifyEmailResend(c *gin.Context) {
	var req resendVerificationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}
	if err := h.uc.ResendEmailVerification(c, req.Email); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

type passwordResetReq struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) PostV1AuthPasswordReset(c *gin.Context) {
	var req passwordResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}
	if err := h.uc.RequestPasswordReset(c, req.Email); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

type passwordResetConfirmReq struct {
	Token       string `json:"token"        binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) PostV1AuthPasswordResetConfirm(c *gin.Context) {
	var req passwordResetConfirmReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}
	if err := h.uc.ConfirmPasswordReset(c, req.Token, req.NewPassword); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) cookieDefaults() (string, string, http.SameSite) {
	name := h.cookie.Name
	if name == "" {
		name = "forma_token"
	}
	path := h.cookie.Path
	if path == "" {
		path = "/"
	}
	sameSite := h.cookie.SameSite
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
		Domain:   h.cookie.Domain,
		Secure:   h.cookie.Secure,
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
		Domain:   h.cookie.Domain,
		Secure:   h.cookie.Secure,
		HttpOnly: true,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}

func (h *AuthHandler) sessionIDFromCookie(c *gin.Context) (uuid.UUID, bool) {
	name, _, _ := h.cookieDefaults()
	cookie, err := c.Request.Cookie(name)
	if err != nil || cookie.Value == "" {
		return uuid.UUID{}, false
	}
	sid, err := uuid.Parse(cookie.Value)
	if err != nil {
		return uuid.UUID{}, false
	}
	return sid, true
}
