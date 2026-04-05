package ses

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

//go:embed templates
var templateFS embed.FS

var _ repository.EmailSender = (*EmailSender)(nil)

type EmailSender struct {
	client  *sesv2.Client
	from    string
	replyTo []string
}

func NewEmailSender(cfg aws.Config, from string, replyTo []string) *EmailSender {
	return &EmailSender{
		client:  sesv2.NewFromConfig(cfg),
		from:    from,
		replyTo: replyTo,
	}
}

func (s *EmailSender) SendEmail(ctx context.Context, input repository.SendEmailInput) error {
	subject, html, text, err := renderTemplate(input.TemplateName, input.TemplateData)
	if err != nil {
		return err
	}

	// SES v2 の SendEmail が返すエラーはすべてインフラ側の問題であり、
	// usecase が分岐してビジネスエラーに変換するケースがないため、
	// repository エラーへの変換は行わずそのまま返す。
	_, err = s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		ReplyToAddresses: s.replyTo,
		Destination: &types.Destination{
			ToAddresses: input.To,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data:    aws.String(html),
						Charset: aws.String("UTF-8"),
					},
					Text: &types.Content{
						Data:    aws.String(text),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	})
	return err
}

// SESのtemplateを使用すると管理の手間が増えるため、ローカルのテンプレートファイルを使用してメール内容を生成する
func renderTemplate(name string, data map[string]string) (subject, html, text string, err error) {
	dir := "templates/" + name

	subjectBytes, err := templateFS.ReadFile(dir + "/subject.txt")
	if err != nil {
		return "", "", "", fmt.Errorf("template %s: subject.txt not found", name)
	}
	htmlBytes, err := templateFS.ReadFile(dir + "/body.html")
	if err != nil {
		return "", "", "", fmt.Errorf("template %s: body.html not found", name)
	}
	textBytes, err := templateFS.ReadFile(dir + "/body.txt")
	if err != nil {
		return "", "", "", fmt.Errorf("template %s: body.txt not found", name)
	}

	subject = replaceVars(string(subjectBytes), data)
	html = replaceVars(string(htmlBytes), data)
	text = replaceVars(string(textBytes), data)
	return subject, html, text, nil
}

func replaceVars(tmpl string, data map[string]string) string {
	for k, v := range data {
		tmpl = strings.ReplaceAll(tmpl, "{{"+k+"}}", v)
	}
	return tmpl
}
