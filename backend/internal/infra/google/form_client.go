package google

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/hiromichi-5/forma/backend/internal/repository"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var _ repository.FormFetcher = (*FormClient)(nil)

type FormClient struct {
	svc *forms.Service
}

func NewFormClient(ctx context.Context, saJSONPath string) (*FormClient, error) {
	svc, err := forms.NewService(
		ctx,
		option.WithAuthCredentialsFile(option.ServiceAccount, saJSONPath),
	)
	if err != nil {
		return nil, err
	}
	return &FormClient{svc: svc}, nil
}

func (c *FormClient) GetForm(ctx context.Context, formID string) (*repository.GoogleForm, error) {
	form, err := c.svc.Forms.Get(formID).Context(ctx).Do()
	if err != nil {
		return nil, classifyGoogleAPIError(err)
	}

	gf := &repository.GoogleForm{
		FormID: formID,
	}
	if form.Info != nil {
		gf.Title = form.Info.Title
		gf.Description = form.Info.Description
	}

	for _, item := range form.Items {
		if item == nil {
			continue
		}
		questions := extractQuestions(item)
		if len(questions) == 0 {
			continue
		}

		gItem := repository.GoogleFormItem{
			Title: item.Title,
		}
		for _, q := range questions {
			if q == nil || q.QuestionId == "" {
				continue
			}
			gItem.Questions = append(gItem.Questions, repository.GoogleFormQuestion{
				QuestionID:   q.QuestionId,
				QuestionType: detectQuestionType(q),
				Choices:      extractChoices(q),
			})
		}
		if len(gItem.Questions) > 0 {
			gf.Items = append(gf.Items, gItem)
		}
	}

	return gf, nil
}

func (c *FormClient) ListResponses(
	ctx context.Context,
	formID, filter, pageToken string,
) (*repository.GoogleFormResponsePage, error) {
	call := c.svc.Forms.Responses.List(formID).Context(ctx)
	if filter != "" {
		call = call.Filter(filter)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, classifyGoogleAPIError(err)
	}

	page := &repository.GoogleFormResponsePage{
		NextPageToken: resp.NextPageToken,
	}

	for _, r := range resp.Responses {
		if r == nil {
			continue
		}

		ts := r.LastSubmittedTime
		if ts == "" {
			ts = r.CreateTime
		}
		if ts == "" {
			continue
		}

		submittedAt, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}

		answersJSON := marshalAnswers(r.Answers)

		page.Responses = append(page.Responses, repository.GoogleFormResponse{
			ResponseID:      r.ResponseId,
			RespondentEmail: r.RespondentEmail,
			SubmittedAt:     submittedAt,
			AnswersJSON:     answersJSON,
		})
	}

	return page, nil
}

func extractQuestions(item *forms.Item) []*forms.Question {
	if item == nil {
		return nil
	}
	if item.QuestionItem != nil && item.QuestionItem.Question != nil {
		return []*forms.Question{item.QuestionItem.Question}
	}
	if item.QuestionGroupItem != nil && len(item.QuestionGroupItem.Questions) > 0 {
		return item.QuestionGroupItem.Questions
	}
	return nil
}

func detectQuestionType(q *forms.Question) string {
	if q == nil {
		return "unknown"
	}
	if q.TextQuestion != nil {
		if q.TextQuestion.Paragraph {
			return "paragraph"
		}
		return "text"
	}
	if q.ChoiceQuestion != nil {
		switch strings.ToUpper(q.ChoiceQuestion.Type) {
		case "RADIO":
			return "radio"
		case "CHECKBOX":
			return "checkbox"
		case "DROP_DOWN":
			return "drop_down"
		default:
			return "choice"
		}
	}
	if q.ScaleQuestion != nil {
		return "scale"
	}
	if q.DateQuestion != nil {
		return "date"
	}
	if q.TimeQuestion != nil {
		return "time"
	}
	if q.FileUploadQuestion != nil {
		return "file"
	}
	if q.RowQuestion != nil {
		return "row"
	}
	if q.RatingQuestion != nil {
		return "rating"
	}
	return "unknown"
}

func extractChoices(q *forms.Question) []string {
	if q == nil || q.ChoiceQuestion == nil {
		return nil
	}
	var choices []string
	for _, opt := range q.ChoiceQuestion.Options {
		if opt == nil {
			continue
		}
		if v := strings.TrimSpace(opt.Value); v != "" {
			choices = append(choices, opt.Value)
		}
	}
	return choices
}

func marshalAnswers(answers map[string]forms.Answer) []byte {
	if len(answers) == 0 {
		return nil
	}
	b, err := json.Marshal(answers)
	if err != nil {
		return nil
	}
	return b
}

func classifyGoogleAPIError(err error) error {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	switch apiErr.Code {
	case http.StatusForbidden:
		return repository.ErrForbidden
	case http.StatusNotFound:
		return repository.ErrNotFound
	default:
		return err
	}
}
