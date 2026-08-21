package handler

import (
	"time"

	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
)

type loginResp struct {
	SessionID string `json:"session_id"`
}

type signupResp struct {
	ID string `json:"id"`
}

type userProfileResp struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
	VerifiedAt  *string `json:"verified_at"`
}

func toUserProfileResp(u entity.User) userProfileResp {
	return userProfileResp{
		ID:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		VerifiedAt:  timePtr(u.VerifiedAt),
	}
}

type formResp struct {
	ID              string  `json:"id"`
	FormID          string  `json:"form_id"`
	Title           string  `json:"title"`
	Description     *string `json:"description"`
	TitleQuestionID *string `json:"title_question_id"`
	CreatedAt       string  `json:"created_at"`
}

type formSummaryResp struct {
	ID     string `json:"id"`
	FormID string `json:"form_id"`
	Title  string `json:"title"`
}

func toFormResp(f entity.Form) formResp {
	return formResp{
		ID:              f.ID.String(),
		FormID:          f.GoogleFormID,
		Title:           f.Title,
		Description:     f.Description,
		TitleQuestionID: f.TitleQuestionID,
		CreatedAt:       f.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toFormSummaryListResp(forms []entity.Form) []formSummaryResp {
	out := make([]formSummaryResp, len(forms))
	for i, f := range forms {
		out[i] = formSummaryResp{
			ID:     f.ID.String(),
			FormID: f.GoogleFormID,
			Title:  f.Title,
		}
	}
	return out
}

type questionResp struct {
	QuestionID   string   `json:"question_id"`
	Title        string   `json:"title"`
	QuestionType string   `json:"question_type"`
	Options      []string `json:"options"`
}

func toQuestionListResp(questions []entity.FormQuestion) []questionResp {
	out := make([]questionResp, len(questions))
	for i, q := range questions {
		opts := q.Options
		if opts == nil {
			opts = []string{}
		}
		out[i] = questionResp{
			QuestionID:   q.QuestionID,
			Title:        q.Title,
			QuestionType: q.QuestionType,
			Options:      opts,
		}
	}
	return out
}

type memberResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func toMemberListResp(members []entity.Member) []memberResp {
	out := make([]memberResp, len(members))
	for i, m := range members {
		out[i] = memberResp{
			ID:          m.ID.String(),
			Email:       m.Email,
			DisplayName: m.DisplayName,
			Role:        string(m.Role),
		}
	}
	return out
}

type createInviteResp struct {
	InviteID  string `json:"invite_id"`
	ExpiresAt string `json:"expires_at"`
}

type inviteResp struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	InvitedBy string `json:"invited_by"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

func toInviteListResp(invites []entity.Invite) []inviteResp {
	out := make([]inviteResp, len(invites))
	for i, inv := range invites {
		out[i] = inviteResp{
			ID:        inv.ID.String(),
			Email:     inv.Email,
			Role:      string(inv.Role),
			InvitedBy: inv.InvitedBy.String(),
			ExpiresAt: inv.ExpiresAt.UTC().Format(time.RFC3339),
			CreatedAt: inv.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return out
}

type statusResp struct {
	ID           string  `json:"id"`
	FormID       string  `json:"form_id"`
	Name         string  `json:"name"`
	Color        *string `json:"color"`
	DisplayOrder int32   `json:"display_order"`
	IsDefault    bool    `json:"is_default"`
}

func toStatusResp(s entity.FormStatus) statusResp {
	return statusResp{
		ID:           s.ID.String(),
		FormID:       s.FormID.String(),
		Name:         s.Name,
		Color:        s.Color,
		DisplayOrder: s.DisplayOrder,
		IsDefault:    s.IsDefault,
	}
}

func toStatusListResp(statuses []entity.FormStatus) []statusResp {
	out := make([]statusResp, len(statuses))
	for i, s := range statuses {
		out[i] = toStatusResp(s)
	}
	return out
}

type ticketStatusResp struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type ticketAssigneeResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type ticketSummaryResp struct {
	ID              string              `json:"id"`
	FormID          string              `json:"form_id"`
	FormTitle       string              `json:"form_title"`
	ResponseID      string              `json:"response_id"`
	RespondentEmail *string             `json:"respondent_email"`
	Status          ticketStatusResp    `json:"status"`
	Priority        string              `json:"priority"`
	TitleQuestionID *string             `json:"title_question_id"`
	Title           string              `json:"title"`
	Assignee        *ticketAssigneeResp `json:"assignee"`
	SubmittedAt     string              `json:"submitted_at"`
	CreatedAt       string              `json:"created_at"`
}

type ticketAnswerResp struct {
	QuestionID    string   `json:"question_id"`
	QuestionTitle string   `json:"question_title"`
	QuestionType  string   `json:"question_type"`
	Values        []string `json:"values"`
	DisplayValue  string   `json:"display_value"`
}

type ticketNotificationResp struct {
	NotificationType string  `json:"notification_type"`
	LastSentAt       *string `json:"last_sent_at"`
}

type ticketDetailResp struct {
	ticketSummaryResp
	Answers       []ticketAnswerResp       `json:"answers"`
	Notifications []ticketNotificationResp `json:"notifications"`
}

type ticketUpdateResp struct {
	ticketDetailResp
	NotificationResults []notificationResultResp `json:"notification_results"`
}

type notificationResultResp struct {
	NotificationType string `json:"notification_type"`
	Result           string `json:"result"`
}

func toTicketSummaryResp(t usecase.TicketSummary) ticketSummaryResp {
	resp := ticketSummaryResp{
		ID:              t.ID.String(),
		FormID:          t.FormID.String(),
		FormTitle:       t.FormTitle,
		ResponseID:      t.ResponseID,
		RespondentEmail: t.RespondentEmail,
		Status: ticketStatusResp{
			ID:    t.Status.ID.String(),
			Name:  t.Status.Name,
			Color: t.Status.Color,
		},
		Priority:        string(t.Priority),
		TitleQuestionID: t.TitleQuestionID,
		Title:           t.Title,
		SubmittedAt:     t.SubmittedAt.UTC().Format(time.RFC3339),
		CreatedAt:       t.CreatedAt.UTC().Format(time.RFC3339),
	}
	if t.Assignee != nil {
		resp.Assignee = &ticketAssigneeResp{
			ID:          t.Assignee.ID.String(),
			Email:       t.Assignee.Email,
			DisplayName: t.Assignee.DisplayName,
		}
	}
	return resp
}

func toTicketSummaryListResp(tickets []usecase.TicketSummary) []ticketSummaryResp {
	out := make([]ticketSummaryResp, len(tickets))
	for i, t := range tickets {
		out[i] = toTicketSummaryResp(t)
	}
	return out
}

func toTicketDetailResp(d usecase.TicketDetail) ticketDetailResp {
	answers := make([]ticketAnswerResp, len(d.Answers))
	for i, a := range d.Answers {
		answers[i] = ticketAnswerResp{
			QuestionID:    a.QuestionID,
			QuestionTitle: a.QuestionTitle,
			QuestionType:  a.QuestionType,
			Values:        a.Values,
			DisplayValue:  a.DisplayValue,
		}
	}
	notifications := make([]ticketNotificationResp, len(d.Notifications))
	for i, n := range d.Notifications {
		var lastSentAt *string
		if n.LastSentAt != nil {
			formatted := n.LastSentAt.UTC().Format(time.RFC3339)
			lastSentAt = &formatted
		}
		notifications[i] = ticketNotificationResp{
			NotificationType: string(n.NotificationType),
			LastSentAt:       lastSentAt,
		}
	}

	return ticketDetailResp{
		ticketSummaryResp: toTicketSummaryResp(d.TicketSummary),
		Answers:           answers,
		Notifications:     notifications,
	}
}

func toTicketUpdateResp(
	d usecase.TicketDetail,
	results []usecase.NotificationResult,
) ticketUpdateResp {
	out := make([]notificationResultResp, len(results))
	for i, r := range results {
		result := "failed"
		if r.Sent {
			result = "sent"
		}
		out[i] = notificationResultResp{
			NotificationType: string(r.NotificationType),
			Result:           result,
		}
	}
	return ticketUpdateResp{
		ticketDetailResp:    toTicketDetailResp(d),
		NotificationResults: out,
	}
}

func toNotificationSettingsResp(s usecase.NotificationSettings) notificationSettingsResp {
	settings := make([]notificationSettingResp, len(s.Settings))
	for i, v := range s.Settings {
		settings[i] = notificationSettingResp{
			NotificationType: string(v.NotificationType),
			Mode:             string(v.Mode),
			IncludeDetail:    v.IncludeDetail,
		}
	}
	return notificationSettingsResp{
		EmailCollectionType: (*string)(s.EmailCollectionType),
		Settings:            settings,
	}
}

type notificationSettingResp struct {
	NotificationType string `json:"notification_type"`
	Mode             string `json:"mode"`
	IncludeDetail    bool   `json:"include_detail"`
}

type notificationSettingsResp struct {
	EmailCollectionType *string                   `json:"email_collection_type"`
	Settings            []notificationSettingResp `json:"settings"`
}

type sentNotificationResp struct {
	NotificationType string `json:"notification_type"`
	SentAt           string `json:"sent_at"`
}

func toSentNotificationResp(n entity.TicketNotification) sentNotificationResp {
	return sentNotificationResp{
		NotificationType: string(n.NotificationType),
		SentAt:           n.SentAt.UTC().Format(time.RFC3339),
	}
}

type ticketHistoryResp struct {
	ID            string  `json:"id"`
	TicketID      string  `json:"ticket_id"`
	ChangedBy     *string `json:"changed_by"`
	ChangedByName string  `json:"changed_by_name"`
	FieldName     string  `json:"field_name"`
	OldValue      *string `json:"old_value"`
	NewValue      *string `json:"new_value"`
	CreatedAt     string  `json:"created_at"`
}

func toTicketHistoryListResp(histories []entity.TicketHistory) []ticketHistoryResp {
	out := make([]ticketHistoryResp, len(histories))
	for i, h := range histories {
		var changedBy *string
		if h.ChangedBy != nil {
			s := h.ChangedBy.String()
			changedBy = &s
		}
		out[i] = ticketHistoryResp{
			ID:            h.ID.String(),
			TicketID:      h.TicketID.String(),
			ChangedBy:     changedBy,
			ChangedByName: h.ChangedByName,
			FieldName:     string(h.FieldName),
			OldValue:      h.OldValue,
			NewValue:      h.NewValue,
			CreatedAt:     h.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return out
}

type syncResp struct {
	Synced     bool   `json:"synced"`
	NewTickets int    `json:"new_tickets"`
	Last       string `json:"last"`
}

func timePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339Nano)
	return &s
}
