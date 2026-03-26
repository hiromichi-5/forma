package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type FormsHandler struct{ S *service.Service }

type registerReq struct {
	URL string `json:"url" binding:"required"`
}

func (h *FormsHandler) PostV1Forms(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	formID, err := h.S.RegisterForm(c, req.URL, uid)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
			return
		case errors.Is(err, service.ErrFormsNotShared):
			c.JSON(http.StatusNotFound, gin.H{"code": "FORMS_NOT_FOUND"})
			return
		case errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "FORMS_NOT_FOUND"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"id": formID.String()})
}

func (h *FormsHandler) GetV1Forms(c *gin.Context) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	fs, err := h.S.ListForms(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"forms": fs})
}

func (h *FormsHandler) GetV1FormsId(c *gin.Context, formID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	form, err := h.S.GetForm(c, formID, uid)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.JSON(http.StatusOK, form)
}

type updateFormReq struct {
	TitleQuestionID *string `json:"title_question_id"`
}

func (h *FormsHandler) PatchV1FormsId(c *gin.Context, formID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	var req updateFormReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		return
	}

	if err := h.S.UpdateFormTitleQuestion(c, formID, req.TitleQuestionID, uid); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *FormsHandler) GetV1FormsFormIdQuestions(c *gin.Context, formID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	questions, err := h.S.ListFormQuestions(c, formID, uid)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": questions})
}

type updateTitleQuestionReq struct {
	TitleQuestionID *string `json:"title_question_id"`
}

func (h *FormsHandler) PatchV1FormsFormIdTitleQuestion(c *gin.Context, formID string) {
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if uid == uuid.Nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	var req updateTitleQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		return
	}

	var questionID *string
	if req.TitleQuestionID != nil && *req.TitleQuestionID != "" {
		questionID = req.TitleQuestionID
	}

	if err := h.S.UpdateFormTitleQuestion(c, formID, questionID, uid); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(404, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(500, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.Status(204)
}
