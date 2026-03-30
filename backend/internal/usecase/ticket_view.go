package usecase

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
)

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

type formQuestionSet struct {
	ordered        []entity.FormQuestion
	defaultTitleID string
}

func newFormQuestionSet(questions []entity.FormQuestion) formQuestionSet {
	set := formQuestionSet{
		ordered: questions,
	}
	if len(questions) > 0 {
		set.defaultTitleID = questions[0].QuestionID
	}
	return set
}

func parseResponseAnswers(payload []byte) (map[string][]string, error) {
	if len(payload) == 0 {
		return map[string][]string{}, nil
	}
	var wrapper storedResponse
	if err := json.Unmarshal(payload, &wrapper); err == nil && wrapper.Answers != nil {
		return flattenAnswers(wrapper.Answers), nil
	}
	var raw map[string]storedAnswer
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return flattenAnswers(raw), nil
}

func flattenAnswers(raw map[string]storedAnswer) map[string][]string {
	results := make(map[string][]string, len(raw))
	for key, ans := range raw {
		results[key] = extractTextValues(ans)
	}
	return results
}

func extractTextValues(ans storedAnswer) []string {
	if ans.TextAnswers == nil {
		return []string{}
	}
	out := make([]string, 0, len(ans.TextAnswers.Answers))
	for _, v := range ans.TextAnswers.Answers {
		out = append(out, v.Value)
	}
	return out
}

func buildSummary(
	ticket entity.Ticket,
	fctx formContext,
	answers map[string][]string,
) TicketSummary {
	titleQID := fctx.titleQuestionID()
	title := deriveTitle(
		titleQID,
		answers,
		fctx.questions,
		fctx.form.Title,
		ticket.ResponseID,
	)

	status := TicketStatus{}
	if s, ok := fctx.statuses[ticket.StatusID]; ok {
		status = TicketStatus{ID: s.ID, Name: s.Name, Color: s.Color}
	}

	var titleQIDPtr *string
	if titleQID != "" {
		titleQIDPtr = &titleQID
	}

	return TicketSummary{
		ID:              ticket.ID,
		FormID:          ticket.FormID,
		FormTitle:       fctx.form.Title,
		ResponseID:      ticket.ResponseID,
		RespondentEmail: ticket.RespondentEmail,
		Status:          status,
		Priority:        ticket.Priority,
		TitleQuestionID: titleQIDPtr,
		Title:           title,
		Assignee:        buildAssignee(ticket.AssigneeID, fctx.members),
		SubmittedAt:     ticket.SubmittedAt,
		CreatedAt:       ticket.CreatedAt,
	}
}

func buildDetail(
	ticket entity.Ticket,
	fctx formContext,
	answers map[string][]string,
) TicketDetail {
	return TicketDetail{
		TicketSummary: buildSummary(ticket, fctx, answers),
		Answers:       buildAnswerList(answers, fctx.questions),
	}
}

func buildAnswerList(answers map[string][]string, questions formQuestionSet) []TicketAnswer {
	result := make([]TicketAnswer, 0, len(questions.ordered))
	used := make(map[string]struct{}, len(answers))

	for _, q := range questions.ordered {
		vals := answers[q.QuestionID]
		if vals == nil {
			vals = []string{}
		}
		result = append(result, TicketAnswer{
			QuestionID:    q.QuestionID,
			QuestionTitle: q.Title,
			QuestionType:  q.QuestionType,
			Values:        vals,
			DisplayValue:  joinValues(vals),
		})
		used[q.QuestionID] = struct{}{}
	}

	var extraIDs []string
	for id := range answers {
		if _, ok := used[id]; !ok {
			extraIDs = append(extraIDs, id)
		}
	}
	sort.Strings(extraIDs)
	for _, id := range extraIDs {
		vals := answers[id]
		if vals == nil {
			vals = []string{}
		}
		result = append(result, TicketAnswer{
			QuestionID:    id,
			QuestionTitle: id,
			QuestionType:  "unknown",
			Values:        vals,
			DisplayValue:  joinValues(vals),
		})
	}

	return result
}

func deriveTitle(
	titleQID string,
	answers map[string][]string,
	questions formQuestionSet,
	formTitle, responseID string,
) string {
	if title := joinValues(answers[titleQID]); title != "" {
		return title
	}

	used := make(map[string]struct{}, len(answers))
	for _, q := range questions.ordered {
		used[q.QuestionID] = struct{}{}
		if q.QuestionID == titleQID {
			continue
		}
		if title := joinValues(answers[q.QuestionID]); title != "" {
			return title
		}
	}

	var extraIDs []string
	for id := range answers {
		if _, ok := used[id]; !ok {
			extraIDs = append(extraIDs, id)
		}
	}
	sort.Strings(extraIDs)
	for _, id := range extraIDs {
		if title := joinValues(answers[id]); title != "" {
			return title
		}
	}
	if formTitle != "" {
		return formTitle
	}
	return responseID
}

func buildAssignee(
	assigneeID *uuid.UUID,
	members map[uuid.UUID]entity.Member,
) *TicketAssignee {
	if assigneeID == nil {
		return nil
	}
	m, ok := members[*assigneeID]
	if !ok {
		return nil
	}
	name := m.DisplayName
	if name == "" {
		name = m.Email
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
