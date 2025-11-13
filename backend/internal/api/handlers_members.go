package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type memberAddReq struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin editor"`
}
type memberRoleUpdateReq struct {
	Role string `json:"role" binding:"required,oneof=admin editor"`
}

func (h *FormsHandler) GetV1FormsFormIdMembers(c *gin.Context) {
	formID := c.Param("form_id")
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if err := h.S.RequireAdmin(c, formID, uid); err != nil {
		c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		return
	}
	ms, err := h.S.ListMembers(c, formID)
	if err != nil {
		c.JSON(500, gin.H{"code": "INTERNAL"})
		return
	}
	c.JSON(200, gin.H{"members": ms})
}

func (h *FormsHandler) PostV1FormsFormIdMembers(c *gin.Context) {
	formID := c.Param("form_id")
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if err := h.S.RequireAdmin(c, formID, uid); err != nil {
		c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		return
	}
	var req memberAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	if err := h.S.AddMember(c, formID, req.Email, req.Role); err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		case service.ErrUserNotFound:
			c.JSON(404, gin.H{"code": "USER_NOT_FOUND"})
		default:
			c.JSON(500, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.Status(201)
}

func (h *FormsHandler) PutV1FormsFormIdMembersUserId(c *gin.Context) {
	formID := c.Param("form_id")
	userID := c.Param("user_id")
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if err := h.S.RequireAdmin(c, formID, uid); err != nil {
		c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	var req memberRoleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	if err := h.S.ChangeRole(c, formID, userUUID.String(), req.Role); err != nil {
		if err == service.ErrValidation {
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
			return
		}
		c.JSON(500, gin.H{"code": "INTERNAL"})
		return
	}
	c.Status(200)
}

func (h *FormsHandler) DeleteV1FormsFormIdMembersUserId(c *gin.Context) {
	formID := c.Param("form_id")
	userID := c.Param("user_id")
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if err := h.S.RequireAdmin(c, formID, uid); err != nil {
		c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	if err := h.S.RemoveMember(c, formID, userUUID.String()); err != nil {
		if err == service.ErrValidation {
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
			return
		}
		c.JSON(500, gin.H{"code": "INTERNAL"})
		return
	}
	c.Status(204)
}
