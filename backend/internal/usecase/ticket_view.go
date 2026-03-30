package usecase

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

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

	return TicketSummary{
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
