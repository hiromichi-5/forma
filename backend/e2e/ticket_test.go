package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketScenario(t *testing.T) {
	ctx := context.Background()
	testutil.TruncateAll(t, ctx, testPool)

	answersJSON, _ := json.Marshal(map[string]any{
		"q1": map[string]any{
			"questionId": "q1",
			"textAnswers": map[string]any{
				"answers": []map[string]any{
					{"value": "テスト太郎"},
				},
			},
		},
	})

	mockFetcher.GetFormFunc = func(_ context.Context, formID string) (*repository.GoogleForm, error) {
		if formID == "ticket-test-form-00001" {
			return &repository.GoogleForm{
				FormID:      formID,
				Title:       "チケットテスト用フォーム",
				Description: "",
				Items: []repository.GoogleFormItem{
					{
						Title: "お名前",
						Questions: []repository.GoogleFormQuestion{
							{QuestionID: "q1", QuestionType: "TEXT"},
						},
					},
				},
			}, nil
		}
		return nil, repository.ErrNotFound
	}

	submittedAt := time.Now().Add(-time.Hour)
	callCount := 0
	mockFetcher.ListResponsesFunc = func(_ context.Context, formID, _, _ string) (*repository.GoogleFormResponsePage, error) {
		callCount++
		if callCount == 1 {
			// 最初の sync: 2件のレスポンスを返す
			return &repository.GoogleFormResponsePage{
				Responses: []repository.GoogleFormResponse{
					{
						ResponseID:      "resp-001",
						RespondentEmail: "respondent1@example.com",
						SubmittedAt:     submittedAt,
						AnswersJSON:     answersJSON,
					},
					{
						ResponseID:      "resp-002",
						RespondentEmail: "respondent2@example.com",
						SubmittedAt:     submittedAt.Add(time.Minute),
						AnswersJSON:     answersJSON,
					},
				},
			}, nil
		}
		// 2回目以降: 新規レスポンスなし
		return &repository.GoogleFormResponsePage{}, nil
	}
	defer func() {
		mockFetcher.GetFormFunc = nil
		mockFetcher.ListResponsesFunc = nil
	}()

	client := loginUser(t, "ticket-admin@example.com", "password123", "TicketAdmin")

	// フォーム登録
	resp := postJSON(t, client, "/v1/forms", map[string]string{"url": "ticket-test-form-00001"})
	var formBody map[string]string
	readJSON(t, resp, &formBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	formID := formBody["id"]

	var ticketID string
	var statusIDs []string

	t.Run("sync: Google Formsのレスポンスを同期してチケットを作成できる", func(t *testing.T) {
		resp := postJSON(t, client, fmt.Sprintf("/v1/forms/%s/sync", formID), nil)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, true, body["synced"])
		assert.Equal(t, float64(2), body["new_tickets"])
	})

	t.Run("sync: 再同期で新規チケットが増えないこと", func(t *testing.T) {
		resp := postJSON(t, client, fmt.Sprintf("/v1/forms/%s/sync", formID), nil)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, float64(0), body["new_tickets"])
	})

	t.Run("チケット一覧: form_idでチケットを取得できる", func(t *testing.T) {
		resp := get(t, client, fmt.Sprintf("/v1/tickets?form_id=%s", formID))
		var body struct {
			Tickets []map[string]any `json:"tickets"`
		}
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, body.Tickets, 2)
		ticketID = body.Tickets[0]["id"].(string)
	})

	t.Run("チケット詳細: チケットの詳細と回答を取得できる", func(t *testing.T) {
		require.NotEmpty(t, ticketID)

		resp := get(t, client, fmt.Sprintf("/v1/tickets/%s", ticketID))
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "medium", body["priority"])
		assert.NotNil(t, body["status"])
	})

	t.Run("ステータスIDを取得", func(t *testing.T) {
		resp := get(t, client, fmt.Sprintf("/v1/forms/%s/statuses", formID))
		var body struct {
			Statuses []map[string]any `json:"statuses"`
		}
		readJSON(t, resp, &body)
		require.Len(t, body.Statuses, 3)
		for _, s := range body.Statuses {
			statusIDs = append(statusIDs, s["id"].(string))
		}
	})

	t.Run("チケット更新: ステータスを変更できる", func(t *testing.T) {
		require.NotEmpty(t, ticketID)
		require.True(t, len(statusIDs) >= 2)

		// "対応中" (2番目のステータス) に変更
		resp := patchJSON(t, client, fmt.Sprintf("/v1/tickets/%s", ticketID), map[string]any{
			"status_id": statusIDs[1],
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		status := body["status"].(map[string]any)
		assert.Equal(t, "対応中", status["name"])
	})

	t.Run("チケット更新: 担当者を設定できる", func(t *testing.T) {
		// 自分自身を担当者に
		meResp := get(t, client, "/v1/me")
		var me map[string]any
		readJSON(t, meResp, &me)
		myID := me["id"].(string)

		resp := patchJSON(t, client, fmt.Sprintf("/v1/tickets/%s", ticketID), map[string]any{
			"assignee_id": myID,
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, body["assignee"])
		assignee := body["assignee"].(map[string]any)
		assert.Equal(t, myID, assignee["id"])
	})

	t.Run("チケット更新: 優先度を変更できる", func(t *testing.T) {
		resp := patchJSON(t, client, fmt.Sprintf("/v1/tickets/%s", ticketID), map[string]any{
			"priority": "high",
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "high", body["priority"])
	})

	t.Run("チケット更新: 担当者をクリアできる", func(t *testing.T) {
		resp := patchJSON(t, client, fmt.Sprintf("/v1/tickets/%s", ticketID), map[string]any{
			"assignee_id": nil,
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Nil(t, body["assignee"])
	})

	t.Run("履歴確認: チケットの変更履歴を取得できる", func(t *testing.T) {
		resp := get(t, client, fmt.Sprintf("/v1/tickets/%s/histories", ticketID))
		var body struct {
			Histories []map[string]any `json:"histories"`
		}
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// status変更, assignee設定, priority変更, assigneeクリア = 4件
		require.Len(t, body.Histories, 4)

		// フィールド名が記録されている
		fields := make([]string, len(body.Histories))
		for i, h := range body.Histories {
			fields[i] = h["field_name"].(string)
		}
		assert.Contains(t, fields, "status")
		assert.Contains(t, fields, "assignee")
		assert.Contains(t, fields, "priority")
	})
}
