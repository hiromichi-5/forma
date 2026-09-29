package resend

import (
	"testing"

	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	t.Run("HTML 本文に差し込む値だけがエスケープされること", func(t *testing.T) {
		t.Parallel()

		title := `<script>alert("x")</script> & Co.`
		subject, html, text, err := renderTemplate(
			repository.TemplateTicketStatusChanged,
			map[string]string{"form_title": title, "status_name": "完了"},
		)
		require.NoError(t, err)

		assert.Contains(t, html, "&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt; &amp; Co.")
		assert.NotContains(t, html, "<script>")
		assert.Contains(t, subject, title)
		assert.Contains(t, text, title)
	})
}
