package usecase

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

var reFormID = regexp.MustCompile(`/forms/d/([a-zA-Z0-9_-]+)/`)

type FormUseCase struct {
	formRepo   repository.FormRepository
	memberRepo repository.MemberRepository
	statusRepo repository.StatusRepository
	fetcher    repository.FormFetcher
	uow        repository.UnitOfWork[repository.FormRepos]
}

func NewFormUseCase(
	formRepo repository.FormRepository,
	memberRepo repository.MemberRepository,
	statusRepo repository.StatusRepository,
	fetcher repository.FormFetcher,
	uow repository.UnitOfWork[repository.FormRepos],
) *FormUseCase {
	return &FormUseCase{
		formRepo:   formRepo,
		memberRepo: memberRepo,
		statusRepo: statusRepo,
		fetcher:    fetcher,
		uow:        uow,
	}
}

func (uc *FormUseCase) RegisterForm(
	ctx context.Context,
	formURL string,
	userID uuid.UUID,
) (entity.Form, error) {
	googleFormID, err := extractFormID(formURL)
	if err != nil {
		return entity.Form{}, entity.NewError(entity.CodeValidation)
	}

	gf, err := uc.fetcher.GetForm(ctx, googleFormID)
	if err != nil {
		if errors.Is(err, repository.ErrForbidden) {
			return entity.Form{}, entity.NewError(entity.CodeFormNotShared)
		}
		if errors.Is(err, repository.ErrNotFound) {
			return entity.Form{}, entity.NewError(entity.CodeFormNotFound)
		}
		return entity.Form{}, err
	}
	if gf == nil {
		return entity.Form{}, entity.NewError(entity.CodeFormNotFound)
	}

	title := gf.Title
	if title == "" {
		title = googleFormID
	}

	var description *string
	if gf.Description != "" {
		description = &gf.Description
	}

	var form entity.Form
	err = uc.uow.Do(ctx, func(repos repository.FormRepos) error {
		var txErr error
		form, txErr = repos.Form.Create(ctx, entity.Form{
			ID:          uuid.New(),
			FormID:      googleFormID,
			Title:       title,
			Description: description,
		})
		if txErr != nil {
			return txErr
		}

		if txErr = repos.Member.Upsert(ctx, userID, form.ID, entity.RoleAdmin); txErr != nil {
			return txErr
		}

		return uc.initFormStatuses(ctx, repos.Status, form.ID)
	})
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return entity.Form{}, entity.NewError(entity.CodeFormAlreadyRegistered)
		}
		return entity.Form{}, err
	}

	return form, nil
}

func (uc *FormUseCase) initFormStatuses(
	ctx context.Context,
	statusRepo repository.StatusRepository,
	formID uuid.UUID,
) error {
	statuses := []struct {
		name      string
		color     string
		order     int32
		isDefault bool
	}{
		{name: "未対応", color: "#E53935", order: 1, isDefault: true},
		{name: "対応中", color: "#FB8C00", order: 2, isDefault: false},
		{name: "対応完了", color: "#43A047", order: 3, isDefault: false},
	}
	for _, st := range statuses {
		color := st.color
		if _, err := statusRepo.Create(ctx, entity.FormStatus{
			ID:           uuid.New(),
			FormID:       formID,
			Name:         st.name,
			Color:        &color,
			DisplayOrder: st.order,
			IsDefault:    st.isDefault,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (uc *FormUseCase) ListForms(ctx context.Context, userID uuid.UUID) ([]entity.Form, error) {
	return uc.memberRepo.ListAccessibleForms(ctx, userID)
}

func (uc *FormUseCase) GetForm(ctx context.Context, formID, userID uuid.UUID) (entity.Form, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return entity.Form{}, err
	}

	form, err := uc.formRepo.GetByID(ctx, formID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.Form{}, entity.NewError(entity.CodeFormNotFound)
		}
		return entity.Form{}, err
	}
	return form, nil
}

func (uc *FormUseCase) UpdateTitleQuestion(
	ctx context.Context,
	formID, userID uuid.UUID,
	questionID *string,
) error {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return err
	}

	if questionID != nil {
		questions, err := uc.formRepo.ListQuestions(ctx, formID)
		if err != nil {
			return err
		}
		found := false
		for _, q := range questions {
			if q.QuestionID == *questionID {
				found = true
				break
			}
		}
		if !found {
			return entity.NewError(entity.CodeValidation)
		}
	}

	if err := uc.formRepo.UpdateTitleQuestion(ctx, formID, questionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeFormNotFound)
		}
		return err
	}
	return nil
}

func (uc *FormUseCase) ListQuestions(
	ctx context.Context,
	formID, userID uuid.UUID,
) ([]entity.FormQuestion, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return nil, err
	}
	return uc.formRepo.ListQuestions(ctx, formID)
}

func extractFormID(u string) (string, error) {
	if m := reFormID.FindStringSubmatch(u); len(m) == 2 {
		return m[1], nil
	}
	if len(u) >= 20 && !strings.Contains(u, "/") {
		return u, nil
	}
	if _, err := url.ParseRequestURI(u); err != nil {
		return "", entity.NewError(entity.CodeValidation)
	}
	return "", entity.NewError(entity.CodeValidation)
}

func mapFormFetcherError(err error) error {
	if errors.Is(err, repository.ErrForbidden) {
		return entity.NewError(entity.CodeFormNotShared)
	}
	if errors.Is(err, repository.ErrNotFound) {
		return entity.NewError(entity.CodeFormNotFound)
	}
	return err
}
