package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormScenario(t *testing.T) {
	ctx := context.Background()
	testutil.TruncateAll(t, ctx, testPool)

	mockFetcher.GetFormFunc = func(_ context.Context, formID string) (*repository.GoogleForm, error) {
		if formID == "test-form-id-123456789" {
			return &repository.GoogleForm{
				FormID:      formID,
				Title:       "テストフォーム",
				Description: "テスト用の説明",
				Items: []repository.GoogleFormItem{
					{
						Title: "お名前",
						Questions: []repository.GoogleFormQuestion{
							{QuestionID: "q1", QuestionType: "TEXT", Choices: nil},
						},
					},
					{
						Title: "ご意見",
						Questions: []repository.GoogleFormQuestion{
							{QuestionID: "q2", QuestionType: "PARAGRAPH_TEXT", Choices: nil},
						},
					},
				},
			}, nil
		}
		return nil, repository.ErrNotFound
	}
	defer func() { mockFetcher.GetFormFunc = nil }()

	client := loginUser(t, "form-admin@example.com", "password123", "FormAdmin")

	var formUUID string

	t.Run("フォーム登録: Google Form IDでフォームを登録できる", func(t *testing.T) {
		resp := postJSON(t, client, "/v1/forms", map[string]string{
			"url": "test-form-id-123456789",
		})
		var body map[string]string
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NotEmpty(t, body["id"])
		formUUID = body["id"]
	})

	t.Run("フォーム登録: 同じフォームは409で重複登録できない", func(t *testing.T) {
		resp := postJSON(t, client, "/v1/forms", map[string]string{
			"url": "test-form-id-123456789",
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		assert.Equal(t, "FORM_ALREADY_REGISTERED", body["code"])
	})

	t.Run("フォーム一覧: 自分が所属するフォームを取得できる", func(t *testing.T) {
		resp := get(t, client, "/v1/forms")
		var body struct {
			Forms []map[string]any `json:"forms"`
		}
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, body.Forms, 1)
		assert.Equal(t, "テストフォーム", body.Forms[0]["title"])
	})

	t.Run("フォーム詳細: フォームの詳細を取得できる", func(t *testing.T) {
		require.NotEmpty(t, formUUID)

		resp := get(t, client, "/v1/forms/"+formUUID)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "テストフォーム", body["title"])
		assert.Equal(t, "テスト用の説明", body["description"])
	})

	t.Run("ステータス一覧: 初期ステータス3件が作成されている", func(t *testing.T) {
		resp := get(t, client, fmt.Sprintf("/v1/forms/%s/statuses", formUUID))
		var body struct {
			Statuses []map[string]any `json:"statuses"`
		}
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, body.Statuses, 3)
		assert.Equal(t, "未対応", body.Statuses[0]["name"])
		assert.Equal(t, "対応中", body.Statuses[1]["name"])
		assert.Equal(t, "対応完了", body.Statuses[2]["name"])
	})

	var newStatusID string

	t.Run("ステータス作成: 新しいステータスを追加できる", func(t *testing.T) {
		color := "#9C27B0"
		resp := postJSON(t, client, fmt.Sprintf("/v1/forms/%s/statuses", formUUID), map[string]any{
			"name":          "保留",
			"color":         color,
			"display_order": 4,
			"is_default":    false,
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "保留", body["name"])
		assert.Equal(t, "#9C27B0", body["color"])
		newStatusID = body["id"].(string)
	})

	t.Run("ステータス更新: ステータス名を変更できる", func(t *testing.T) {
		require.NotEmpty(t, newStatusID)

		resp := patchJSON(
			t,
			client,
			fmt.Sprintf("/v1/forms/%s/statuses/%s", formUUID, newStatusID),
			map[string]any{
				"name": "一時保留",
			},
		)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "一時保留", body["name"])
	})

	t.Run("ステータス削除: ステータスを削除できる", func(t *testing.T) {
		require.NotEmpty(t, newStatusID)

		resp := del(t, client, fmt.Sprintf("/v1/forms/%s/statuses/%s", formUUID, newStatusID))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// 削除後は3件に戻る
		resp2 := get(t, client, fmt.Sprintf("/v1/forms/%s/statuses", formUUID))
		var body struct {
			Statuses []map[string]any `json:"statuses"`
		}
		readJSON(t, resp2, &body)
		assert.Len(t, body.Statuses, 3)
	})

	t.Run("質問一覧: フォームの質問を取得できる", func(t *testing.T) {
		resp := postJSON(t, client, fmt.Sprintf("/v1/forms/%s/sync", formUUID), nil)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp2 := get(t, client, fmt.Sprintf("/v1/forms/%s/questions", formUUID))
		var body struct {
			Questions []map[string]any `json:"questions"`
		}
		readJSON(t, resp2, &body)

		assert.Equal(t, http.StatusOK, resp2.StatusCode)
		require.Len(t, body.Questions, 2)
		assert.Equal(t, "お名前", body.Questions[0]["title"])
		assert.Equal(t, "ご意見", body.Questions[1]["title"])
	})

	t.Run("フォーム設定更新: title_question_id を設定できる", func(t *testing.T) {
		resp := patchJSON(t, client, "/v1/forms/"+formUUID, map[string]any{
			"title_question_id": "q1",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("フォーム設定更新: title_question_id を null でクリアできる", func(t *testing.T) {
		resp := patchJSON(t, client, "/v1/forms/"+formUUID, map[string]any{
			"title_question_id": nil,
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("フォーム登録解除: editor は削除できず403で失敗する", func(t *testing.T) {
		editorID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"form-editor@example.com",
			"password123",
			"FormEditor",
		)
		testutil.AddMember(t, ctx, testPool, editorID, uuid.MustParse(formUUID), "editor")
		editorClient := loginUserExisting(t, "form-editor@example.com", "password123")

		resp := del(t, editorClient, "/v1/forms/"+formUUID)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "FORBIDDEN", body["code"])
	})

	t.Run("フォーム登録解除: admin がフォームを削除できる", func(t *testing.T) {
		resp := del(t, client, "/v1/forms/"+formUUID)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp2 := get(t, client, "/v1/forms/"+formUUID)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp2.StatusCode)

		resp3 := get(t, client, "/v1/forms")
		var body struct {
			Forms []map[string]any `json:"forms"`
		}
		readJSON(t, resp3, &body)
		assert.Equal(t, http.StatusOK, resp3.StatusCode)
		assert.Empty(t, body.Forms)
	})
}
