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

func TestMemberInviteScenario(t *testing.T) {
	ctx := context.Background()
	testutil.TruncateAll(t, ctx, testPool)

	mockFetcher.GetFormFunc = func(_ context.Context, formID string) (*repository.GoogleForm, error) {
		if formID == "member-test-form-00001" {
			return &repository.GoogleForm{
				FormID: formID,
				Title:  "メンバーテスト用フォーム",
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

	// admin ユーザーでフォームを登録
	adminClient := loginUser(t, "admin@example.com", "password123", "Admin")
	resp := postJSON(
		t,
		adminClient,
		"/v1/forms",
		map[string]string{"url": "member-test-form-00001"},
	)
	var formBody map[string]string
	readJSON(t, resp, &formBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	formID := formBody["id"]

	// editor ユーザーを作成（まだフォームに未参加）
	editorUserID := testutil.CreateVerifiedUser(
		t,
		ctx,
		testPool,
		"editor@example.com",
		"password123",
		"Editor",
	)

	var inviteID string

	t.Run("招待作成: admin がメンバーを招待できる", func(t *testing.T) {
		resp := postJSON(
			t,
			adminClient,
			fmt.Sprintf("/v1/forms/%s/invites", formID),
			map[string]string{
				"email": "editor@example.com",
				"role":  "editor",
			},
		)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NotEmpty(t, body["invite_id"])
		inviteID = body["invite_id"].(string)
	})

	t.Run("招待作成: 同じメールへの重複招待は409で失敗する", func(t *testing.T) {
		resp := postJSON(
			t,
			adminClient,
			fmt.Sprintf("/v1/forms/%s/invites", formID),
			map[string]string{
				"email": "editor@example.com",
				"role":  "editor",
			},
		)
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		assert.Equal(t, "ACTIVE_INVITE_ALREADY_EXISTS", body["code"])
	})

	t.Run("招待一覧: 未承諾の招待を一覧取得できる", func(t *testing.T) {
		resp := get(t, adminClient, fmt.Sprintf("/v1/forms/%s/invites", formID))
		var body struct {
			Invites []map[string]any `json:"invites"`
		}
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, body.Invites, 1)
		assert.Equal(t, "editor@example.com", body.Invites[0]["email"])
	})

	t.Run("招待受諾: 招待されたユーザーが受諾してメンバーになる", func(t *testing.T) {
		require.NotEmpty(t, inviteID)

		// editor@example.com のユーザーでログイン
		editorClient := loginUserExisting(t, "editor@example.com", "password123")

		resp := postJSON(t, editorClient, fmt.Sprintf("/v1/invites/%s/accept", inviteID), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("メンバー一覧: admin と editor が表示される", func(t *testing.T) {
		resp := get(t, adminClient, fmt.Sprintf("/v1/forms/%s/members", formID))
		var body struct {
			Members []map[string]any `json:"members"`
		}
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, body.Members, 2)
	})

	t.Run("ロール変更: editor を admin に変更できる", func(t *testing.T) {
		resp := putJSON(
			t,
			adminClient,
			fmt.Sprintf("/v1/forms/%s/members/%s", formID, editorUserID),
			map[string]string{
				"role": "admin",
			},
		)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("メンバー削除: admin がメンバーを削除できる", func(t *testing.T) {
		resp := del(t, adminClient, fmt.Sprintf("/v1/forms/%s/members/%s", formID, editorUserID))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// メンバーは1人に戻る
		resp2 := get(t, adminClient, fmt.Sprintf("/v1/forms/%s/members", formID))
		var body struct {
			Members []map[string]any `json:"members"`
		}
		readJSON(t, resp2, &body)
		assert.Len(t, body.Members, 1)
	})

	t.Run("最後のadmin削除: 最後の管理者は409で削除できない", func(t *testing.T) {
		// admin@example.com のユーザーIDを取得
		resp := get(t, adminClient, "/v1/me")
		var me map[string]any
		readJSON(t, resp, &me)
		adminUserID := me["id"].(string)

		resp2 := del(t, adminClient, fmt.Sprintf("/v1/forms/%s/members/%s", formID, adminUserID))
		var body map[string]any
		readJSON(t, resp2, &body)

		assert.Equal(t, http.StatusConflict, resp2.StatusCode)
		assert.Equal(t, "LAST_ADMIN", body["code"])
	})
}
