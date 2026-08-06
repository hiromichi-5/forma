package resend

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/hiromichi-5/forma/backend/internal/repository"
	resendsdk "github.com/resend/resend-go/v3"
)

//go:embed templates
var templateFS embed.FS

var _ repository.EmailSender = (*EmailSender)(nil)

type EmailSender struct {
	client  *resendsdk.Client
	from    string
	replyTo string
}

func NewEmailSender(apiKey, from, replyTo string) *EmailSender {
	return &EmailSender{
		client:  resendsdk.NewClient(apiKey),
		from:    from,
		replyTo: replyTo,
	}
}

func (s *EmailSender) SendEmail(ctx context.Context, input repository.SendEmailInput) error {
	subject, html, text, err := renderTemplate(input.TemplateName, input.TemplateData)
	if err != nil {
		return err
	}

	// Resend の Send が返すエラーはすべてインフラ側の問題であり、
	// usecase が分岐してビジネスエラーに変換するケースがないため、
	// repository エラーへの変換は行わずそのまま返す。
	_, err = s.client.Emails.SendWithContext(ctx, &resendsdk.SendEmailRequest{
		From:    s.from,
		To:      input.To,
		ReplyTo: s.replyTo,
		Subject: subject,
		Html:    html,
		Text:    text,
	})
	return err
}

// ローカルのテンプレートファイルを使用してメール内容を生成する。AmazonSESを使っていた際にテンプレート機能を利用する手間のほうが大きかったため。Resendでも当面はこの方法で運用する。
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

	subject = strings.TrimSpace(replaceVars(string(subjectBytes), data))
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
