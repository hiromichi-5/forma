package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

type TicketSummary struct {
	ID              uuid.UUID
	FormID          uuid.UUID
	FormTitle       string
	ResponseID      string
	RespondentEmail *string
	Status          entity.FormStatus
	Priority        entity.Priority
	TitleQuestionID *string
	Title           string
	Assignee        *entity.UserRef
	SubmittedAt     time.Time
	CreatedAt       time.Time
}

type TicketAnswer struct {
	QuestionID    string
	QuestionTitle string
	QuestionType  string
	Values        []string
}

type TicketNotificationInfo struct {
	NotificationType entity.NotificationType
	LastSentAt       *time.Time
}

type TicketDetail struct {
	TicketSummary
	Answers       []TicketAnswer
	Notifications []TicketNotificationInfo
}

type UpdateTicketInput struct {
	StatusID *uuid.UUID
	Assignee entity.AssigneeChange
	Priority *entity.Priority
}

type TicketNotifier interface {
	NotifyTicketUpdated(
		ctx context.Context,
		ticketID, userID uuid.UUID,
		notificationTypes []entity.NotificationType,
	) []NotificationResult
	ListLatestSent(ctx context.Context, ticketID uuid.UUID) ([]entity.TicketNotification, error)
}

type TicketUseCase struct {
	ticketRepo repository.TicketRepository
	formRepo   repository.FormRepository
	statusRepo repository.StatusRepository
	memberRepo repository.MemberRepository
	userRepo   repository.UserRepository
	authz      *Authorizer
	uow        repository.UnitOfWork[repository.TicketRepos]
	publisher  EventPublisher
	notifier   TicketNotifier
}

func NewTicketUseCase(
	ticketRepo repository.TicketRepository,
	formRepo repository.FormRepository,
	statusRepo repository.StatusRepository,
	memberRepo repository.MemberRepository,
	userRepo repository.UserRepository,
	authz *Authorizer,
	uow repository.UnitOfWork[repository.TicketRepos],
	publisher EventPublisher,
	notifier TicketNotifier,
) *TicketUseCase {
	return &TicketUseCase{
		ticketRepo: ticketRepo,
		formRepo:   formRepo,
		statusRepo: statusRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		authz:      authz,
		uow:        uow,
		publisher:  publisher,
		notifier:   notifier,
	}
}

type formContext struct {
	form      entity.Form
	statuses  map[uuid.UUID]entity.FormStatus
	members   map[uuid.UUID]entity.Member
	questions formQuestionSet
}

func (fc formContext) titleQuestionID() string {
	if fc.form.TitleQuestionID != nil {
		return *fc.form.TitleQuestionID
	}
	return fc.questions.defaultTitleID
}

func (uc *TicketUseCase) loadFormContext(
	ctx context.Context,
	formID uuid.UUID,
) (formContext, error) {
	form, err := uc.formRepo.GetByID(ctx, formID)
	if err != nil {
		return formContext{}, err
	}

	statuses, err := uc.statusRepo.List(ctx, formID)
	if err != nil {
		return formContext{}, err
	}
	statusMap := make(map[uuid.UUID]entity.FormStatus, len(statuses))
	for _, s := range statuses {
		statusMap[s.ID] = s
	}

	members, err := uc.memberRepo.List(ctx, formID)
	if err != nil {
		return formContext{}, err
	}
	memberMap := make(map[uuid.UUID]entity.Member, len(members))
	for _, m := range members {
		memberMap[m.ID] = m
	}

	questions, err := uc.formRepo.ListQuestions(ctx, formID)
	if err != nil {
		return formContext{}, err
	}

	return formContext{
		form:      form,
		statuses:  statusMap,
		members:   memberMap,
		questions: newFormQuestionSet(questions),
	}, nil
}

func (uc *TicketUseCase) ListTickets(
	ctx context.Context,
	formID, userID uuid.UUID,
	statusID *uuid.UUID,
) ([]TicketSummary, error) {
	if err := uc.authz.RequireEditor(ctx, formID, userID); err != nil {
		return nil, err
	}

	if statusID != nil {
		if _, err := uc.getVisibleStatus(ctx, formID, *statusID); err != nil {
			return nil, err
		}
	}

	tickets, err := uc.ticketRepo.List(ctx, formID, statusID)
	if err != nil {
		return nil, err
	}
	if len(tickets) == 0 {
		return []TicketSummary{}, nil
	}

	fctx, err := uc.loadFormContext(ctx, formID)
	if err != nil {
		return nil, err
	}

	summaries := make([]TicketSummary, 0, len(tickets))
	for _, t := range tickets {
		answers, err := parseResponseAnswers(t.Answers)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, buildSummary(t, fctx, answers))
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
			return TicketDetail{}, entity.NewError(entity.CodeResourceHidden)
		}
		return TicketDetail{}, err
	}

	if err := uc.authz.RequireEditor(ctx, ticket.FormID, userID); err != nil {
		return TicketDetail{}, err
	}

	fctx, err := uc.loadFormContext(ctx, ticket.FormID)
	if err != nil {
		return TicketDetail{}, err
	}

	answers, err := parseResponseAnswers(ticket.Answers)
	if err != nil {
		return TicketDetail{}, err
	}

	detail := buildDetail(ticket, fctx, answers)

	sent, err := uc.notifier.ListLatestSent(ctx, ticketID)
	if err != nil {
		return TicketDetail{}, err
	}
	detail.Notifications = buildNotificationInfo(sent)

	return detail, nil
}

func buildNotificationInfo(sent []entity.TicketNotification) []TicketNotificationInfo {
	lastSentAt := make(map[entity.NotificationType]time.Time, len(sent))
	for _, s := range sent {
		lastSentAt[s.NotificationType] = s.SentAt
	}

	types := entity.NotificationTypes()
	infos := make([]TicketNotificationInfo, 0, len(types))
	for _, t := range types {
		info := TicketNotificationInfo{NotificationType: t}
		if at, ok := lastSentAt[t]; ok {
			v := at
			info.LastSentAt = &v
		}
		infos = append(infos, info)
	}
	return infos
}

func (uc *TicketUseCase) UpdateTicket(
	ctx context.Context,
	ticketID, userID uuid.UUID,
	in UpdateTicketInput,
) (TicketDetail, []NotificationResult, error) {
	ticket, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return TicketDetail{}, nil, entity.NewError(entity.CodeResourceHidden)
		}
		return TicketDetail{}, nil, err
	}

	if err := uc.authz.RequireEditor(ctx, ticket.FormID, userID); err != nil {
		return TicketDetail{}, nil, err
	}

	changedByName, err := getUserDisplayName(ctx, uc.userRepo, userID)
	if err != nil {
		return TicketDetail{}, nil, err
	}

	var newStatus *entity.FormStatus
	if in.StatusID != nil {
		s, err := uc.getVisibleStatus(ctx, ticket.FormID, *in.StatusID)
		if err != nil {
			return TicketDetail{}, nil, err
		}
		newStatus = &s
	}

	newAssignee := in.Assignee.UserID()
	if newAssignee != nil {
		if err := uc.validateAssignee(ctx, ticket.FormID, *newAssignee); err != nil {
			return TicketDetail{}, nil, err
		}
	}

	if in.Priority != nil && !in.Priority.Valid() {
		return TicketDetail{}, nil, entity.NewError(entity.CodeValidation)
	}

	var notifyTypes []entity.NotificationType

	if err := uc.uow.Do(ctx, func(repos repository.TicketRepos) error {
		notifyTypes = nil

		current, err := repos.Ticket.GetByID(ctx, ticketID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeResourceHidden)
			}
			return err
		}

		rec := changeRecorder{
			ticketID:  ticketID,
			changedBy: userID,
			name:      changedByName,
		}

		if oldStatusID, changed := current.ChangeStatus(newStatus); changed {
			oldStatus, err := repos.Status.GetByID(ctx, oldStatusID)
			if err != nil {
				return err
			}
			rec.record(entity.FieldStatus, strPtr(oldStatus.Name), strPtr(newStatus.Name))
			notifyTypes = append(notifyTypes, entity.NotificationTypeStatusChange)
		}

		if oldAssigneeID, changed := current.ChangeAssignee(in.Assignee); changed {
			oldName, err := resolveUserName(ctx, repos.User, oldAssigneeID)
			if err != nil {
				return err
			}
			newName, err := resolveUserName(ctx, repos.User, newAssignee)
			if err != nil {
				return err
			}
			rec.record(entity.FieldAssignee, oldName, newName)
			// 担当者の解除は通知の対象外
			if newAssignee != nil {
				notifyTypes = append(notifyTypes, entity.NotificationTypeAssigneeAssigned)
			}
		}

		if oldPriority, changed := current.ChangePriority(in.Priority); changed {
			rec.record(
				entity.FieldPriority,
				strPtr(string(oldPriority)),
				strPtr(string(current.Priority)),
			)
		}

		if len(rec.entries) == 0 {
			return nil
		}

		if err := repos.Ticket.Save(ctx, current); err != nil {
			return err
		}
		return rec.save(ctx, repos.Ticket)
	}); err != nil {
		return TicketDetail{}, nil, err
	}

	//nolint:errcheck,gosec
	uc.publisher.PublishTicketUpdated(ctx, TicketEvent{FormID: ticket.FormID, TicketID: ticketID})

	results := uc.notifier.NotifyTicketUpdated(ctx, ticketID, userID, notifyTypes)

	detail, err := uc.GetTicket(ctx, ticketID, userID)
	if err != nil {
		return TicketDetail{}, nil, err
	}
	return detail, results, nil
}

func (uc *TicketUseCase) ListTicketHistories(
	ctx context.Context,
	ticketID, userID uuid.UUID,
) ([]entity.TicketHistory, error) {
	ticket, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, entity.NewError(entity.CodeResourceHidden)
		}
		return nil, err
	}

	if err := uc.authz.RequireEditor(ctx, ticket.FormID, userID); err != nil {
		return nil, err
	}

	return uc.ticketRepo.ListHistories(ctx, ticketID)
}

type changeRecorder struct {
	ticketID  uuid.UUID
	changedBy uuid.UUID
	name      string
	entries   []entity.TicketHistory
}

func (r *changeRecorder) record(field entity.TicketField, oldValue, newValue *string) {
	r.entries = append(r.entries, entity.TicketHistory{
		ID:            uuid.New(),
		TicketID:      r.ticketID,
		ChangedBy:     &r.changedBy,
		ChangedByName: r.name,
		FieldName:     field,
		OldValue:      oldValue,
		NewValue:      newValue,
	})
}

func (r *changeRecorder) save(ctx context.Context, ticketRepo repository.TicketRepository) error {
	for _, h := range r.entries {
		if _, err := ticketRepo.CreateHistory(ctx, h); err != nil {
			return err
		}
	}
	return nil
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

func resolveUserName(
	ctx context.Context,
	userRepo repository.UserRepository,
	userID *uuid.UUID,
) (*string, error) {
	if userID == nil {
		return nil, nil
	}
	name, err := getUserDisplayName(ctx, userRepo, *userID)
	if err != nil {
		return nil, err
	}
	return &name, nil
}

func (uc *TicketUseCase) validateAssignee(
	ctx context.Context,
	formID, assigneeID uuid.UUID,
) error {
	if _, err := uc.userRepo.GetByID(ctx, assigneeID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeUserNotFound)
		}
		return err
	}

	_, err := uc.memberRepo.GetRole(ctx, formID, assigneeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeResourceHidden)
		}
		return err
	}
	return nil
}

func (uc *TicketUseCase) getVisibleStatus(
	ctx context.Context,
	formID, statusID uuid.UUID,
) (entity.FormStatus, error) {
	status, err := uc.statusRepo.GetByID(ctx, statusID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.FormStatus{}, entity.NewError(entity.CodeResourceHidden)
		}
		return entity.FormStatus{}, err
	}
	if status.FormID != formID {
		return entity.FormStatus{}, entity.NewError(entity.CodeResourceHidden)
	}
	return status, nil
}

func strPtr(s string) *string {
	return &s
}
