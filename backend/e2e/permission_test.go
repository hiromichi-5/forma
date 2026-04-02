package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionScenario(t *testing.T) {
	ctx := context.Background()
	testutil.TruncateAll(t, ctx, testPool)

	mockFetcher.GetFormFunc = func(_ context.Context, formID string) (*repository.GoogleForm, error) {
		if formID == "perm-test-form-000001" {
			return &repository.GoogleForm{
				FormID: formID,
				Title:  "権限テスト用フォーム",
				Items: []repository.GoogleFormItem{
					{
						Title: "質問1",
						Questions: []repository.GoogleFormQuestion{
							{QuestionID: "pq1", QuestionType: "TEXT"},
						},
					},
				},
			}, nil
		}
		return nil, repository.ErrNotFound
	}
	mockFetcher.ListResponsesFunc = func(_ context.Context, _, _, _ string) (*repository.GoogleFormResponsePage, error) {
		return &repository.GoogleFormResponsePage{}, nil
	}
	defer func() {
		mockFetcher.GetFormFunc = nil
		mockFetcher.ListResponsesFunc = nil
	}()

	// admin: フォームを登録
	adminClient := loginUser(t, "admin@example.com", "password123", "Admin")

	resp := postJSON(t, adminClient, "/v1/forms", map[string]string{"url": "perm-test-form-000001"})
	var formBody map[string]string
	readJSON(t, resp, &formBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	formID := formBody["id"]

	// editor ユーザーを作成してメンバーに追加
	editorUserID := testutil.CreateVerifiedUser(
		t,
		ctx,
		testPool,
		"editor@example.com",
		"password123",
		"Editor",
	)
	resp2 := postJSON(
		t,
		adminClient,
		fmt.Sprintf("/v1/forms/%s/members", formID),
		map[string]string{
			"email": "editor@example.com",
			"role":  "editor",
		},
	)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusCreated, resp2.StatusCode)
	_ = editorUserID

	editorClient := loginUserExisting(t, "editor@example.com", "password123")

	// 非メンバーユーザー
	nonMemberClient := loginUser(t, "outsider@example.com", "password123", "Outsider")

	t.Run("admin: フォーム詳細を取得できる", func(t *testing.T) {
		resp := get(t, adminClient, "/v1/forms/"+formID)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("admin: メンバー一覧を取得できる", func(t *testing.T) {
		resp := get(t, adminClient, fmt.Sprintf("/v1/forms/%s/members", formID))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("admin: 招待一覧を取得できる", func(t *testing.T) {
		resp := get(t, adminClient, fmt.Sprintf("/v1/forms/%s/invites", formID))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("admin: ステータスを追加できる", func(t *testing.T) {
		resp := postJSON(
			t,
			adminClient,
			fmt.Sprintf("/v1/forms/%s/statuses", formID),
			map[string]any{
				"name":          "admin追加ステータス",
				"display_order": 10,
			},
		)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("editor: フォーム詳細を取得できる", func(t *testing.T) {
		resp := get(t, editorClient, "/v1/forms/"+formID)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("editor: ステータスを追加できる", func(t *testing.T) {
		resp := postJSON(
			t,
			editorClient,
			fmt.Sprintf("/v1/forms/%s/statuses", formID),
			map[string]any{
				"name":          "editor追加ステータス",
				"display_order": 11,
			},
		)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("editor: 招待の作成は403で失敗する", func(t *testing.T) {
		resp := postJSON(
			t,
			editorClient,
			fmt.Sprintf("/v1/forms/%s/invites", formID),
			map[string]string{
				"email": "someone@example.com",
				"role":  "editor",
			},
		)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "FORBIDDEN", body["code"])
	})

	t.Run("editor: 招待一覧の取得は403で失敗する", func(t *testing.T) {
		resp := get(t, editorClient, fmt.Sprintf("/v1/forms/%s/invites", formID))
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "FORBIDDEN", body["code"])
	})

	t.Run("editor: メンバー追加は403で失敗する", func(t *testing.T) {
		resp := postJSON(
			t,
			editorClient,
			fmt.Sprintf("/v1/forms/%s/members", formID),
			map[string]string{
				"email": "someone@example.com",
				"role":  "editor",
			},
		)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "FORBIDDEN", body["code"])
	})

	t.Run("非メンバー: フォームに404でアクセスできない", func(t *testing.T) {
		resp := get(t, nonMemberClient, "/v1/forms/"+formID)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, "RESOURCE_HIDDEN", body["code"])
	})

	t.Run("非メンバー: メンバー一覧に404でアクセスできない", func(t *testing.T) {
		resp := get(t, nonMemberClient, fmt.Sprintf("/v1/forms/%s/members", formID))
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("非メンバー: ステータス一覧に404でアクセスできない", func(t *testing.T) {
		resp := get(t, nonMemberClient, fmt.Sprintf("/v1/forms/%s/statuses", formID))
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("非メンバー: チケット一覧に404でアクセスできない", func(t *testing.T) {
		resp := get(t, nonMemberClient, fmt.Sprintf("/v1/tickets?form_id=%s", formID))
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("未認証: フォーム詳細に401でアクセスできない", func(t *testing.T) {
		resp := get(t, http.DefaultClient, "/v1/forms/"+formID)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "INVALID_SESSION", body["code"])
	})
}
