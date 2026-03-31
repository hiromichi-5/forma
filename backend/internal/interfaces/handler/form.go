package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
)

type FormUseCase interface {
	RegisterForm(ctx context.Context, formURL string, userID uuid.UUID) (entity.Form, error)
	ListForms(ctx context.Context, userID uuid.UUID) ([]entity.Form, error)
	GetForm(ctx context.Context, formID, userID uuid.UUID) (entity.Form, error)
	UpdateTitleQuestion(ctx context.Context, formID, userID uuid.UUID, questionID *string) error
	ListQuestions(ctx context.Context, formID, userID uuid.UUID) ([]entity.FormQuestion, error)
}

type FormHandler struct {
	uc FormUseCase
}

func NewFormHandler(uc FormUseCase) *FormHandler {
	return &FormHandler{uc: uc}
}

type registerReq struct {
	URL string `json:"url" binding:"required"`
}

func (h *FormHandler) PostV1Forms(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	form, err := h.uc.RegisterForm(c, req.URL, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": form.ID.String()})
}

func (h *FormHandler) GetV1Forms(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	forms, err := h.uc.ListForms(c, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"forms": toFormSummaryListResp(forms)})
}

func (h *FormHandler) GetV1FormsId(c *gin.Context) {
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

	form, err := h.uc.GetForm(c, formID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toFormResp(form))
}

type updateFormReq struct {
	TitleQuestionID nullableStringPayload `json:"title_question_id"`
}

type nullableStringPayload struct {
	set   bool
	null  bool
	value string
}

func (n *nullableStringPayload) UnmarshalJSON(data []byte) error {
	n.set = true
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		n.null = true
		n.value = ""
		return nil
	}

	if err := json.Unmarshal(trimmed, &n.value); err != nil {
		return err
	}
	n.null = false
	return nil
}

func (h *FormHandler) PatchV1FormsId(c *gin.Context) {
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

	var req updateFormReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if !req.TitleQuestionID.set {
		c.Status(http.StatusNoContent)
		return
	}

	var questionID *string
	if !req.TitleQuestionID.null {
		questionID = &req.TitleQuestionID.value
	}

	if err := h.uc.UpdateTitleQuestion(c, formID, userID, questionID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FormHandler) GetV1FormsFormIdQuestions(c *gin.Context) {
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

	questions, err := h.uc.ListQuestions(c, formID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": toQuestionListResp(questions)})
}
