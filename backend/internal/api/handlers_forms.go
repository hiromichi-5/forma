package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type FormsHandler struct{ S *service.Service }

type registerReq struct {
	URL        string `json:"url" binding:"required"`
	PollingSec int    `json:"polling_sec"`
}

func (h *FormsHandler) PostV1Forms(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if uid == uuid.Nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	if req.PollingSec == 0 {
		req.PollingSec = 30
	}

	formID, err := h.S.RegisterForm(c, req.URL, req.PollingSec, uid)
	if err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
			return
		case service.ErrFormsNotShared:
			c.JSON(403, gin.H{"code": "FORMS_NOT_SHARED"})
			return
		case service.ErrFormsNotFound:
			c.JSON(404, gin.H{"code": "FORMS_NOT_FOUND"})
			return
		default:
			c.JSON(500, gin.H{"code": "INTERNAL"})
			return
		}
	}
	c.JSON(201, gin.H{"form_id": formID})
}

func (h *FormsHandler) GetV1Forms(c *gin.Context) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	fs, err := h.S.ListForms(c, uid)
	if err != nil {
		c.JSON(500, gin.H{"code": "INTERNAL"})
		return
	}
	c.JSON(200, gin.H{"forms": fs})
}

func (h *FormsHandler) GetV1FormsFormIdHealth(c *gin.Context, formID string) {
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if uid == uuid.Nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	res, err := h.S.Health(c, formID, uid)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
			return
		case service.ErrFormsNotShared:
			c.JSON(403, gin.H{"code": "FORMS_NOT_SHARED"})
			return
		case service.ErrFormsNotFound:
			c.JSON(404, gin.H{"code": "FORMS_NOT_FOUND"})
			return
		default:
			c.JSON(500, gin.H{"code": "INTERNAL"})
			return
		}
	}
	c.JSON(200, res)
}
