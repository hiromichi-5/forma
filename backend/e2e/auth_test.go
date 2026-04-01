package e2e

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthScenario(t *testing.T) {
	ctx := context.Background()
	testutil.TruncateAll(t, ctx, testPool)

	var userID string
	var sessionClient *http.Client

	t.Run("signup: 新規ユーザーを登録できる", func(t *testing.T) {
		resp := postJSON(t, http.DefaultClient, "/v1/auth/signup", map[string]string{
			"email":        "test-user@example.com",
			"password":     "password123",
			"display_name": "Test User",
		})
		var body map[string]string
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NotEmpty(t, body["id"])
		userID = body["id"]
	})

	t.Run("signup: 重複したメールアドレスでの登録は409で失敗する。", func(t *testing.T) {
		resp := postJSON(t, http.DefaultClient, "/v1/auth/signup", map[string]string{
			"email":        "test-user@example.com",
			"password":     "password123",
			"display_name": "Test User 2",
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		assert.Equal(t, "CONFLICT", body["code"])
	})

	t.Run("login: メール未認証のユーザは403でログインできない", func(t *testing.T) {
		resp := postJSON(t, http.DefaultClient, "/v1/auth/login", map[string]string{
			"email":    "test-user@example.com",
			"password": "password123",
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "EMAIL_NOT_VERIFIED", body["code"])
	})

	t.Run("verify-email: トークンでメール認証できる", func(t *testing.T) {
		require.NotEmpty(t, userID, "前のテストでuserIDが取得できていない")

		uid, err := uuid.Parse(userID)
		require.NoError(t, err)
		token := testutil.GetEmailVerificationToken(t, ctx, testPool, uid)

		resp := postJSON(t, http.DefaultClient, "/v1/auth/verify-email", map[string]string{
			"token": token,
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("login: 認証済みユーザーがログインできる", func(t *testing.T) {
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		sessionClient = &http.Client{Jar: jar}

		resp := postJSON(t, sessionClient, "/v1/auth/login", map[string]string{
			"email":    "test-user@example.com",
			"password": "password123",
		})
		var body map[string]string
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, body["session_id"])
	})

	t.Run("login: パスワードの誤りは401でログインできない", func(t *testing.T) {
		resp := postJSON(t, http.DefaultClient, "/v1/auth/login", map[string]string{
			"email":    "test-user@example.com",
			"password": "wrongpassword",
		})
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "INVALID_CREDENTIALS", body["code"])
	})

	t.Run("me: プロフィールを取得できる", func(t *testing.T) {
		require.NotNil(t, sessionClient, "ログインが完了していない")

		resp := get(t, sessionClient, "/v1/me")
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "test-user@example.com", body["email"])
		assert.Equal(t, "Test User", body["display_name"])
		assert.NotNil(t, body["verified_at"])
	})

	t.Run("me: セッションなしは401でアクセスできない", func(t *testing.T) {
		resp := get(t, http.DefaultClient, "/v1/me")
		var body map[string]any
		readJSON(t, resp, &body)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "INVALID_SESSION", body["code"])
	})

	t.Run("password-change: パスワードを変更できる", func(t *testing.T) {
		require.NotNil(t, sessionClient)

		resp := patchJSON(t, sessionClient, "/v1/me/password", map[string]string{
			"current_password": "password123",
			"new_password":     "newpassword456",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// 新しいパスワードでログインできる
		resp2 := postJSON(t, http.DefaultClient, "/v1/auth/login", map[string]string{
			"email":    "test-user@example.com",
			"password": "newpassword456",
		})
		var body map[string]string
		readJSON(t, resp2, &body)
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
	})

	t.Run("password-reset: パスワードをリセットできる", func(t *testing.T) {
		resp := postJSON(t, http.DefaultClient, "/v1/auth/password-reset", map[string]string{
			"email": "test-user@example.com",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		uid, err := uuid.Parse(userID)
		require.NoError(t, err)
		token := testutil.GetPasswordResetToken(t, ctx, testPool, uid)

		resp2 := postJSON(
			t,
			http.DefaultClient,
			"/v1/auth/password-reset/confirm",
			map[string]string{
				"token":        token,
				"new_password": "resetpassword789",
			},
		)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp2.StatusCode)

		// リセット後のパスワードでログインできる
		resp3 := postJSON(t, http.DefaultClient, "/v1/auth/login", map[string]string{
			"email":    "test-user@example.com",
			"password": "resetpassword789",
		})
		var body map[string]string
		readJSON(t, resp3, &body)
		assert.Equal(t, http.StatusOK, resp3.StatusCode)
	})

	t.Run("logout: ログアウトしたセッションを使うと401でアクセスできない", func(t *testing.T) {
		require.NotNil(t, sessionClient)

		resp := postJSON(t, sessionClient, "/v1/auth/login", map[string]string{
			"email":    "test-user@example.com",
			"password": "resetpassword789",
		})
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp2 := postJSON(t, sessionClient, "/v1/auth/logout", nil)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp2.StatusCode)

		// ログアウト後にプロフィール取得を試みる
		resp3 := get(t, sessionClient, "/v1/me")
		var body map[string]any
		readJSON(t, resp3, &body)
		assert.Equal(t, http.StatusUnauthorized, resp3.StatusCode)
	})
}
