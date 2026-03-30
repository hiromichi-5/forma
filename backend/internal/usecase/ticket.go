package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
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
	if err := uc.requireEditor(ctx, formID, userID); err != nil {
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

	if err := uc.requireEditor(ctx, ticket.FormID, userID); err != nil {
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

	if err := uc.requireEditor(ctx, ticket.FormID, userID); err != nil {
		return TicketDetail{}, err
	}

	changedByName, err := uc.getUserDisplayName(ctx, userID)
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
			oldStatus, err := uc.statusRepo.GetByID(ctx, current.StatusID)
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
					name, err := uc.getUserDisplayName(ctx, *current.AssigneeID)
					if err != nil {
						return err
					}
					oldAssigneeNameTx = &name
				}
				if newAssigneePtr != nil {
					name, err := uc.getUserDisplayName(ctx, *newAssigneePtr)
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

	if err := uc.requireEditor(ctx, ticket.FormID, userID); err != nil {
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

func (uc *TicketUseCase) getUserDisplayName(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
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

func (uc *TicketUseCase) requireEditor(ctx context.Context, formID, userID uuid.UUID) error {
	role, err := uc.memberRepo.GetRole(ctx, userID, formID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeForbidden)
		}
		return err
	}
	if role != entity.RoleAdmin && role != entity.RoleEditor {
		return entity.NewError(entity.CodeForbidden)
	}
	return nil
}

type ticketHistoryChange struct {
	fieldName string
	oldValue  *string
	newValue  *string
}

func buildTicketHistoryChanges(
	statusChanged bool, oldStatusName, newStatusName string,
	assigneeChanged bool, oldAssigneeName, newAssigneeName *string,
	priorityChanged bool, oldPriority, newPriority string,
) []ticketHistoryChange {
	changes := make([]ticketHistoryChange, 0, 3)
	if statusChanged {
		old := oldStatusName
		new_ := newStatusName
		changes = append(changes, ticketHistoryChange{
			fieldName: "status",
			oldValue:  &old,
			newValue:  &new_,
		})
	}
	if assigneeChanged {
		changes = append(changes, ticketHistoryChange{
			fieldName: "assignee",
			oldValue:  oldAssigneeName,
			newValue:  newAssigneeName,
		})
	}
	if priorityChanged {
		old := oldPriority
		new_ := newPriority
		changes = append(changes, ticketHistoryChange{
			fieldName: "priority",
			oldValue:  &old,
			newValue:  &new_,
		})
	}
	return changes
}

type formQuestionSet struct {
	ordered        []entity.FormQuestion
	byID           map[string]entity.FormQuestion
	defaultTitleID string
}

type storedResponse struct {
	Answers map[string]storedAnswer `json:"answers"`
}

type storedAnswer struct {
	QuestionID  string           `json:"questionId"`
	TextAnswers *storedTextBlock `json:"textAnswers"`
}

type storedTextBlock struct {
	Answers []storedTextAnswer `json:"answers"`
}

type storedTextAnswer struct {
	Value string `json:"value"`
}

func newFormQuestionSet(questions []entity.FormQuestion) formQuestionSet {
	set := formQuestionSet{
		ordered: make([]entity.FormQuestion, len(questions)),
		byID:    make(map[string]entity.FormQuestion, len(questions)),
	}
	copy(set.ordered, questions)
	for _, q := range questions {
		set.byID[q.QuestionID] = q
		if set.defaultTitleID == "" && isPreferredTitleType(q.QuestionType) {
			set.defaultTitleID = q.QuestionID
		}
	}
	return set
}

func isPreferredTitleType(t string) bool {
	switch strings.ToLower(t) {
	case "text", "paragraph", "radio", "choice", "drop_down", "checkbox":
		return true
	default:
		return false
	}
}

func parseResponseAnswers(payload []byte) (map[string][]string, error) {
	if len(payload) == 0 {
		return map[string][]string{}, nil
	}
	var wrapper storedResponse
	if err := json.Unmarshal(payload, &wrapper); err == nil && wrapper.Answers != nil {
		return extractAnswers(wrapper.Answers), nil
	}
	var raw map[string]storedAnswer
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return extractAnswers(raw), nil
}

func extractAnswers(raw map[string]storedAnswer) map[string][]string {
	results := make(map[string][]string, len(raw))
	for key, ans := range raw {
		values := extractAnswerValues(ans)
		if len(values) == 0 {
			results[key] = []string{}
			continue
		}
		results[key] = values
	}
	return results
}

func extractAnswerValues(ans storedAnswer) []string {
	if ans.TextAnswers == nil {
		return nil
	}
	out := make([]string, 0, len(ans.TextAnswers.Answers))
	for _, v := range ans.TextAnswers.Answers {
		trimmed := strings.TrimSpace(v.Value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func buildTicketSummaryWithAnswers(
	ticket entity.Ticket,
	form entity.Form,
	statusMap map[uuid.UUID]entity.FormStatus,
	memberMap map[uuid.UUID]entity.Member,
	questions formQuestionSet,
	answers map[string][]string,
) TicketSummary {
	titleQuestionID := ""
	if form.TitleQuestionID != nil {
		titleQuestionID = *form.TitleQuestionID
	}
	if titleQuestionID == "" {
		titleQuestionID = questions.defaultTitleID
	}
	title := deriveTitle(titleQuestionID, answers, questions, form.Title, ticket.ResponseID)

	status := TicketStatus{}
	if s, ok := statusMap[ticket.StatusID]; ok {
		status = TicketStatus{ID: s.ID, Name: s.Name, Color: s.Color}
	}

	var titleQIDPtr *string
	if titleQuestionID != "" {
		titleQIDPtr = &titleQuestionID
	}

	summary := TicketSummary{
		ID:              ticket.ID,
		FormID:          ticket.FormID,
		FormTitle:       form.Title,
		ResponseID:      ticket.ResponseID,
		RespondentEmail: ticket.RespondentEmail,
		Status:          status,
		Priority:        ticket.Priority,
		TitleQuestionID: titleQIDPtr,
		Title:           title,
		Assignee:        buildAssigneeFromMap(ticket.AssigneeID, memberMap),
		SubmittedAt:     ticket.SubmittedAt,
		CreatedAt:       ticket.CreatedAt,
	}
	return summary
}

func buildTicketDetail(
	ticket entity.Ticket,
	form entity.Form,
	statusMap map[uuid.UUID]entity.FormStatus,
	memberMap map[uuid.UUID]entity.Member,
	questions formQuestionSet,
	answers map[string][]string,
) TicketDetail {
	summary := buildTicketSummaryWithAnswers(ticket, form, statusMap, memberMap, questions, answers)
	return TicketDetail{
		TicketSummary: summary,
		Answers:       buildTicketAnswers(answers, questions),
	}
}

func buildTicketAnswers(answers map[string][]string, questions formQuestionSet) []TicketAnswer {
	result := make([]TicketAnswer, 0, len(questions.ordered))
	used := make(map[string]struct{}, len(answers))
	for _, q := range questions.ordered {
		vals := answers[q.QuestionID]
		result = append(result, TicketAnswer{
			QuestionID:    q.QuestionID,
			QuestionTitle: q.Title,
			QuestionType:  q.QuestionType,
			Values:        cloneStrings(vals),
			DisplayValue:  joinValues(vals),
		})
		used[q.QuestionID] = struct{}{}
	}
	extraIDs := make([]string, 0)
	for id := range answers {
		if _, ok := used[id]; !ok {
			extraIDs = append(extraIDs, id)
		}
	}
	sort.Strings(extraIDs)
	for _, id := range extraIDs {
		vals := answers[id]
		result = append(result, TicketAnswer{
			QuestionID:    id,
			QuestionTitle: id,
			QuestionType:  "unknown",
			Values:        cloneStrings(vals),
			DisplayValue:  joinValues(vals),
		})
	}
	return result
}

func deriveTitle(
	titleQuestionID string,
	answers map[string][]string,
	questions formQuestionSet,
	formTitle, responseID string,
) string {
	if values := answers[titleQuestionID]; len(values) > 0 {
		joined := joinValues(values)
		if trimmed := strings.TrimSpace(joined); trimmed != "" {
			return trimmed
		}
	}
	if questions.defaultTitleID != "" && questions.defaultTitleID != titleQuestionID {
		if values := answers[questions.defaultTitleID]; len(values) > 0 {
			joined := joinValues(values)
			if trimmed := strings.TrimSpace(joined); trimmed != "" {
				return trimmed
			}
		}
	}
	for _, values := range answers {
		joined := joinValues(values)
		if trimmed := strings.TrimSpace(joined); trimmed != "" {
			return trimmed
		}
	}
	if trimmed := strings.TrimSpace(formTitle); trimmed != "" {
		return trimmed
	}
	return responseID
}

func buildAssigneeFromMap(
	assigneeID *uuid.UUID,
	memberMap map[uuid.UUID]entity.Member,
) *TicketAssignee {
	if assigneeID == nil {
		return nil
	}
	m, ok := memberMap[*assigneeID]
	if !ok {
		return nil
	}
	name := strings.TrimSpace(m.DisplayName)
	if name == "" {
		name = strings.TrimSpace(m.Email)
	}
	return &TicketAssignee{
		ID:          m.ID,
		DisplayName: name,
		Email:       m.Email,
	}
}

func joinValues(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func isValidTicketPriority(p string) bool {
	switch p {
	case entity.PriorityHigh, entity.PriorityMedium, entity.PriorityLow:
		return true
	default:
		return false
	}
}
