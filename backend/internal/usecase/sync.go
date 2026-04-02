package usecase

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/logger"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

type SyncUseCase struct {
	formRepo   repository.FormRepository
	ticketRepo repository.TicketRepository
	statusRepo repository.StatusRepository
	memberRepo repository.MemberRepository
	fetcher    repository.FormFetcher
}

func NewSyncUseCase(
	formRepo repository.FormRepository,
	ticketRepo repository.TicketRepository,
	statusRepo repository.StatusRepository,
	memberRepo repository.MemberRepository,
	fetcher repository.FormFetcher,
) *SyncUseCase {
	return &SyncUseCase{
		formRepo:   formRepo,
		ticketRepo: ticketRepo,
		statusRepo: statusRepo,
		memberRepo: memberRepo,
		fetcher:    fetcher,
	}
}

func (uc *SyncUseCase) SyncFormOnce(
	ctx context.Context,
	formID, userID uuid.UUID,
) (newTickets int, lastSync time.Time, err error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return 0, time.Time{}, err
	}

	log := logger.From(ctx)
	log.Info("form sync started", "form_id", formID.String())

	form, err := uc.formRepo.GetByID(ctx, formID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, time.Time{}, entity.NewError(entity.CodeFormNotFound)
		}
		return 0, time.Time{}, err
	}

	if err := uc.refreshFormQuestions(ctx, form); err != nil {
		return 0, time.Time{}, err
	}

	filter := ""
	if form.SyncedAt != nil {
		formatted := form.SyncedAt.UTC().Format(time.RFC3339)
		filter = "timestamp >= " + formatted
	}

	var all []repository.GoogleFormResponse
	token := ""
	for {
		page, e := uc.fetcher.ListResponses(ctx, form.FormID, filter, token)
		if e != nil {
			return 0, time.Time{}, mapFormFetcherError(e)
		}
		if page != nil {
			all = append(all, page.Responses...)
		}
		if page == nil || page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].SubmittedAt.Before(all[j].SubmittedAt)
	})

	defaultStatus, err := uc.statusRepo.GetDefault(ctx, form.ID)
	if err != nil {
		return 0, time.Time{}, err
	}

	var maxSubmitted time.Time
	for _, r := range all {
		if r.ResponseID == "" {
			continue
		}

		var respondentEmail *string
		if r.RespondentEmail != "" {
			respondentEmail = &r.RespondentEmail
		}

		created, e := uc.ticketRepo.Create(ctx, entity.Ticket{
			ID:              uuid.New(),
			FormID:          form.ID,
			ResponseID:      r.ResponseID,
			RespondentEmail: respondentEmail,
			Answers:         r.AnswersJSON,
			StatusID:        defaultStatus.ID,
			Priority:        entity.PriorityMedium,
			SubmittedAt:     r.SubmittedAt,
		})
		if e != nil {
			return 0, time.Time{}, e
		}
		if created {
			newTickets++
		}
		if r.SubmittedAt.After(maxSubmitted) {
			maxSubmitted = r.SubmittedAt
		}
	}

	if !maxSubmitted.IsZero() {
		if err := uc.formRepo.UpdateSyncedAt(ctx, form.ID, maxSubmitted); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return 0, time.Time{}, entity.NewError(entity.CodeFormNotFound)
			}
			return 0, time.Time{}, err
		}
	}

	log.Info("form sync completed",
		"form_id", formID.String(),
		"new_tickets", newTickets,
	)

	return newTickets, maxSubmitted, nil
}

func (uc *SyncUseCase) refreshFormQuestions(ctx context.Context, form entity.Form) error {
	gf, err := uc.fetcher.GetForm(ctx, form.FormID)
	if err != nil {
		return mapFormFetcherError(err)
	}
	if gf == nil {
		return entity.NewError(entity.CodeFormNotFound)
	}

	for _, item := range gf.Items {
		for _, q := range item.Questions {
			if q.QuestionID == "" {
				continue
			}
			title := item.Title
			if title == "" {
				title = q.QuestionID
			}

			if err := uc.formRepo.UpsertQuestion(ctx, form.ID, entity.FormQuestion{
				QuestionID:   q.QuestionID,
				Title:        title,
				QuestionType: q.QuestionType,
				Options:      q.Choices,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
