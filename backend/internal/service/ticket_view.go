package service

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hiromichi-5/forma/backend/internal/db"
)

type TicketStatus struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color *string   `json:"color"`
}

type TicketAssignee struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
}

type TicketSummary struct {
	ID              uuid.UUID       `json:"id"`
	FormID          string          `json:"form_id"`
	FormTitle       string          `json:"form_title"`
	ResponseID      string          `json:"response_id"`
	RespondentEmail *string         `json:"respondent_email"`
	Status          TicketStatus    `json:"status"`
	Priority        string          `json:"priority"`
	TitleQuestionID *string         `json:"title_question_id"`
	Title           string          `json:"title"`
	Assignee        *TicketAssignee `json:"assignee,omitempty"`
	SubmittedAt     time.Time       `json:"submitted_at"`
	CreatedAt       time.Time       `json:"created_at"`
}

type TicketAnswer struct {
	QuestionID    string   `json:"question_id"`
	QuestionTitle string   `json:"question_title"`
	QuestionType  string   `json:"question_type"`
	Values        []string `json:"values"`
	DisplayValue  string   `json:"display_value"`
}

type TicketDetail struct {
	TicketSummary
	Answers []TicketAnswer `json:"answers"`
}

type formQuestionSet struct {
	ordered        []db.FormQuestion
	byID           map[string]db.FormQuestion
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

func newFormQuestionSet(questions []db.FormQuestion) formQuestionSet {
	set := formQuestionSet{
		ordered:        make([]db.FormQuestion, len(questions)),
		byID:           make(map[string]db.FormQuestion, len(questions)),
		defaultTitleID: "",
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

func buildTicketSummary(row db.ListTicketsRow, answers map[string][]string, questions formQuestionSet) TicketSummary {
	titleQuestionID := strings.TrimSpace(row.TitleQuestionID.String)
	if titleQuestionID == "" {
		titleQuestionID = questions.defaultTitleID
	}
	title := deriveTitle(titleQuestionID, answers, questions, row.FormTitle, row.ResponseID)
	status := TicketStatus{
		ID:    uuid.UUID(row.StatusID.Bytes),
		Name:  row.StatusName,
		Color: ticketTextPtr(row.StatusColor),
	}
	summary := TicketSummary{
		ID:              uuid.UUID(row.ID.Bytes),
		FormID:          uuid.UUID(row.FormID.Bytes).String(),
		FormTitle:       row.FormTitle,
		ResponseID:      row.ResponseID,
		RespondentEmail: ticketTextPtr(row.RespondentEmail),
		Status:          status,
		Priority:        row.Priority,
		TitleQuestionID: stringPtr(titleQuestionID),
		Title:           title,
		Assignee:        buildAssignee(row.AssigneeID, row.AssigneeDisplayName.String, row.AssigneeEmail),
		SubmittedAt:     timeFromTimestamptz(row.SubmittedAt),
		CreatedAt:       timeFromTimestamptz(row.CreatedAt),
	}
	if summary.TitleQuestionID != nil && *summary.TitleQuestionID == "" {
		summary.TitleQuestionID = nil
	}
	return summary
}

func buildTicketDetail(row db.GetTicketRow, answers map[string][]string, questions formQuestionSet) TicketDetail {
	listRow := db.ListTicketsRow(row)
	summary := buildTicketSummary(listRow, answers, questions)
	detail := TicketDetail{TicketSummary: summary}
	detail.Answers = buildTicketAnswers(answers, questions)
	return detail
}

func buildTicketAnswers(answers map[string][]string, questions formQuestionSet) []TicketAnswer {
	result := make([]TicketAnswer, 0, len(questions.ordered))
	used := make(map[string]struct{}, len(answers))
	for _, q := range questions.ordered {
		vals := answers[q.QuestionID]
		display := joinValues(vals)
		result = append(result, TicketAnswer{
			QuestionID:    q.QuestionID,
			QuestionTitle: q.Title,
			QuestionType:  q.QuestionType,
			Values:        cloneStrings(vals),
			DisplayValue:  display,
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
		display := joinValues(vals)
		result = append(result, TicketAnswer{
			QuestionID:    id,
			QuestionTitle: id,
			QuestionType:  "unknown",
			Values:        cloneStrings(vals),
			DisplayValue:  display,
		})
	}
	return result
}

func deriveTitle(titleQuestionID string, answers map[string][]string, questions formQuestionSet, formTitle, responseID string) string {
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

func buildAssignee(id pgtype.UUID, display string, email pgtype.Text) *TicketAssignee {
	if !id.Valid {
		return nil
	}
	name := strings.TrimSpace(display)
	emailValue := email.String
	if name == "" {
		name = strings.TrimSpace(emailValue)
	}
	return &TicketAssignee{
		ID:          uuid.UUID(id.Bytes),
		DisplayName: name,
		Email:       emailValue,
	}
}

func timeFromTimestamptz(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func ticketTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	value := t.String
	return &value
}

func stringPtr(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	value := v
	return &value
}
