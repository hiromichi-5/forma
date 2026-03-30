package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

type TicketStatus struct {
	ID    uuid.UUID
	Name  string
	Color *string
}

type TicketAssignee struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
}

type TicketSummary struct {
	ID              uuid.UUID
	FormID          uuid.UUID
	FormTitle       string
	ResponseID      string
	RespondentEmail *string
	Status          TicketStatus
	Priority        string
	TitleQuestionID *string
	Title           string
	Assignee        *TicketAssignee
	SubmittedAt     time.Time
	CreatedAt       time.Time
}

type TicketAnswer struct {
	QuestionID    string
	QuestionTitle string
	QuestionType  string
	Values        []string
	DisplayValue  string
}

type TicketDetail struct {
	TicketSummary
	Answers []TicketAnswer
}

type TicketUseCase struct {
	ticketRepo repository.TicketRepository
	formRepo   repository.FormRepository
	statusRepo repository.StatusRepository
	memberRepo repository.MemberRepository
	userRepo   repository.UserRepository
	txm        repository.TxManager
}

func NewTicketUseCase(
	ticketRepo repository.TicketRepository,
	formRepo repository.FormRepository,
	statusRepo repository.StatusRepository,
	memberRepo repository.MemberRepository,
	userRepo repository.UserRepository,
	txm repository.TxManager,
) *TicketUseCase {
	return &TicketUseCase{
		ticketRepo: ticketRepo,
		formRepo:   formRepo,
		statusRepo: statusRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		txm:        txm,
	}
}

func (uc *TicketUseCase) ListTickets(
	ctx context.Context,
	formID, userID uuid.UUID,
	statusID *uuid.UUID,
) ([]TicketSummary, error) {
	if err := requireEditor(ctx, uc.memberRepo, formID, userID); err != nil {
		return nil, err
	}

	tickets, err := uc.ticketRepo.List(ctx, formID, statusID)
	if err != nil {
		return nil, err
	}
	if len(tickets) == 0 {
		return []TicketSummary{}, nil
	}

	form, err := uc.formRepo.GetByID(ctx, formID)
	if err != nil {
		return nil, err
	}

	statusMap, err := uc.buildStatusMap(ctx, formID)
	if err != nil {
		return nil, err
	}

	memberMap, err := uc.buildMemberMap(ctx, formID)
	if err != nil {
		return nil, err
	}

	questions, err := uc.formRepo.ListQuestions(ctx, formID)
	if err != nil {
		return nil, err
	}
	qset := newFormQuestionSet(questions)

	summaries := make([]TicketSummary, 0, len(tickets))
	for _, t := range tickets {
		answers, err := parseResponseAnswers(t.Answers)
		if err != nil {
			return nil, err
		}
		summaries = append(
			summaries,
			buildTicketSummaryWithAnswers(t, form, statusMap, memberMap, qset, answers),
		)
	}

	return summaries, nil
}

func (uc *TicketUseCase) GetTicket(
	ctx context.Context,
	ticketID, userID uuid.UUID,
) (TicketDetail, error) {
	ticket, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return TicketDetail{}, entity.NewError(entity.CodeForbidden)
		}
		return TicketDetail{}, err
	}

	if err := requireEditor(ctx, uc.memberRepo, ticket.FormID, userID); err != nil {
		return TicketDetail{}, err
	}

	form, err := uc.formRepo.GetByID(ctx, ticket.FormID)
	if err != nil {
		return TicketDetail{}, err
	}

	statusMap, err := uc.buildStatusMap(ctx, ticket.FormID)
	if err != nil {
		return TicketDetail{}, err
	}

	memberMap, err := uc.buildMemberMap(ctx, ticket.FormID)
	if err != nil {
		return TicketDetail{}, err
	}

	questions, err := uc.formRepo.ListQuestions(ctx, ticket.FormID)
	if err != nil {
		return TicketDetail{}, err
	}
	qset := newFormQuestionSet(questions)

	answers, err := parseResponseAnswers(ticket.Answers)
	if err != nil {
		return TicketDetail{}, err
	}

	return buildTicketDetail(ticket, form, statusMap, memberMap, qset, answers), nil
}

func (uc *TicketUseCase) UpdateTicket(
	ctx context.Context,
	ticketID, userID uuid.UUID,
	statusID *uuid.UUID,
	assigneeID *uuid.UUID,
	clearAssignee bool,
	priority *string,
) (TicketDetail, error) {
	ticket, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return TicketDetail{}, entity.NewError(entity.CodeForbidden)
		}
		return TicketDetail{}, err
	}

	if err := requireEditor(ctx, uc.memberRepo, ticket.FormID, userID); err != nil {
		return TicketDetail{}, err
	}

	changedByName, err := getUserDisplayName(ctx, uc.userRepo, userID)
	if err != nil {
		return TicketDetail{}, err
	}

	if clearAssignee && assigneeID != nil {
		return TicketDetail{}, entity.NewError(entity.CodeValidation)
	}

	if assigneeID != nil {
		if err := uc.validateAssignee(ctx, ticket.FormID, *assigneeID); err != nil {
			return TicketDetail{}, err
		}
	}

	var newStatusName string
	var newStatusUUID uuid.UUID
	if statusID != nil {
		newStatus, err := uc.statusRepo.GetByID(ctx, *statusID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return TicketDetail{}, entity.NewError(entity.CodeValidation)
			}
			return TicketDetail{}, err
		}
		if newStatus.FormID != ticket.FormID {
			return TicketDetail{}, entity.NewError(entity.CodeValidation)
		}
		newStatusName = newStatus.Name
		newStatusUUID = newStatus.ID
	}

	var normalizedPriority string
	if priority != nil {
		normalizedPriority = strings.ToLower(strings.TrimSpace(*priority))
		if !isValidTicketPriority(normalizedPriority) {
			return TicketDetail{}, entity.NewError(entity.CodeValidation)
		}
	}

	if err := uc.txm.Do(ctx, func(repos repository.Repos) error {
		current, err := repos.Ticket.GetByID(ctx, ticketID)
		if err != nil {
			return err
		}

		var statusChanged bool
		var oldStatusNameTx string
		if statusID != nil && newStatusUUID != current.StatusID {
			oldStatus, err := repos.Status.GetByID(ctx, current.StatusID)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return entity.NewError(entity.CodeValidation)
				}
				return err
			}
			oldStatusNameTx = oldStatus.Name
			statusChanged = true
			if err := repos.Ticket.UpdateStatus(ctx, ticketID, newStatusUUID); err != nil {
				return err
			}
		}

		var assigneeChanged bool
		var oldAssigneeNameTx, newAssigneeNameTx *string
		if clearAssignee || assigneeID != nil {
			var newAssigneePtr *uuid.UUID
			if assigneeID != nil {
				newAssigneePtr = assigneeID
			}

			oldMatch := (current.AssigneeID == nil && newAssigneePtr == nil) ||
				(current.AssigneeID != nil && newAssigneePtr != nil && *current.AssigneeID == *newAssigneePtr)

			if !oldMatch {
				assigneeChanged = true
				if current.AssigneeID != nil {
					name, err := getUserDisplayName(ctx, repos.User, *current.AssigneeID)
					if err != nil {
						return err
					}
					oldAssigneeNameTx = &name
				}
				if newAssigneePtr != nil {
					name, err := getUserDisplayName(ctx, repos.User, *newAssigneePtr)
					if err != nil {
						return err
					}
					newAssigneeNameTx = &name
				}
				if err := repos.Ticket.UpdateAssignee(ctx, ticketID, newAssigneePtr); err != nil {
					return err
				}
			}
		}

		var priorityChanged bool
		var oldPriorityTx string
		if priority != nil && normalizedPriority != current.Priority {
			oldPriorityTx = current.Priority
			priorityChanged = true
			if err := repos.Ticket.UpdatePriority(ctx, ticketID, normalizedPriority); err != nil {
				return err
			}
		}

		histories := buildTicketHistoryChanges(
			statusChanged, oldStatusNameTx, newStatusName,
			assigneeChanged, oldAssigneeNameTx, newAssigneeNameTx,
			priorityChanged, oldPriorityTx, normalizedPriority,
		)
		for _, h := range histories {
			if _, err := repos.Ticket.CreateHistory(ctx, entity.TicketHistory{
				ID:            uuid.New(),
				TicketID:      ticketID,
				ChangedBy:     &userID,
				ChangedByName: changedByName,
				FieldName:     h.fieldName,
				OldValue:      h.oldValue,
				NewValue:      h.newValue,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return TicketDetail{}, err
	}

	return uc.GetTicket(ctx, ticketID, userID)
}

func (uc *TicketUseCase) ListTicketHistories(
	ctx context.Context,
	ticketID, userID uuid.UUID,
) ([]entity.TicketHistory, error) {
	ticket, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, entity.NewError(entity.CodeForbidden)
		}
		return nil, err
	}

	if err := requireEditor(ctx, uc.memberRepo, ticket.FormID, userID); err != nil {
		return nil, err
	}

	return uc.ticketRepo.ListHistories(ctx, ticketID)
}

func (uc *TicketUseCase) buildStatusMap(
	ctx context.Context,
	formID uuid.UUID,
) (map[uuid.UUID]entity.FormStatus, error) {
	statuses, err := uc.statusRepo.List(ctx, formID)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]entity.FormStatus, len(statuses))
	for _, s := range statuses {
		m[s.ID] = s
	}
	return m, nil
}

func (uc *TicketUseCase) buildMemberMap(
	ctx context.Context,
	formID uuid.UUID,
) (map[uuid.UUID]entity.Member, error) {
	members, err := uc.memberRepo.List(ctx, formID)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]entity.Member, len(members))
	for _, mb := range members {
		m[mb.ID] = mb
	}
	return m, nil
}

func getUserDisplayName(
	ctx context.Context,
	userRepo repository.UserRepository,
	userID uuid.UUID,
) (string, error) {
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", entity.NewError(entity.CodeUserNotFound)
		}
		return "", err
	}
	return user.DisplayName, nil
}

func (uc *TicketUseCase) validateAssignee(
	ctx context.Context,
	formID uuid.UUID,
	assigneeID uuid.UUID,
) error {
	_, err := uc.memberRepo.GetRole(ctx, assigneeID, formID)
	if err != nil {
		return entity.NewError(entity.CodeValidation)
	}
	return nil
}

func isValidTicketPriority(p string) bool {
	switch p {
	case entity.PriorityHigh, entity.PriorityMedium, entity.PriorityLow:
		return true
	default:
		return false
	}
}
