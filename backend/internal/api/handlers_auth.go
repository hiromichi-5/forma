package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type AuthHandler struct {
	Svc interface {
		Authenticate(ctx context.Context, email, password string) (string, error)
		Signup(ctx context.Context, email, password, displayName string) (string, error)
	}
	JWT    auth.Signer
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
	h.setAuthCookie(c, tok)
	c.Status(http.StatusNoContent)
}

type signupReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1"`
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
	tok, err := h.JWT.Issue(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}
	h.setAuthCookie(c, tok)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) PostV1AuthLogout(c *gin.Context) {
	h.clearAuthCookie(c)
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

func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
	name, path, sameSite := h.cookieDefaults()
	maxAge := int(h.JWT.TTL / time.Second)
	if maxAge <= 0 {
		maxAge = 0
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     path,
		Domain:   h.Cookie.Domain,
		Secure:   h.Cookie.Secure,
		HttpOnly: true,
		SameSite: sameSite,
		MaxAge:   maxAge,
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
