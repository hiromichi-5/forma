package e2e

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type accessLevel int

const (
	levelEditor accessLevel = iota
	levelAdmin
	// levelNoFormScope はフォーム単位の認可を持たないルート
	levelNoFormScope
)

type routeSpec struct {
	level accessLevel
	// body は POST / PATCH / PUT で bind を通すために必要な最小の入力。
	// バリデーションは認可より先に実行されるため、妥当な値でないと 400 になる。
	body  any
	query string
}

// 全てのルートについて、必要な権限を宣言する。宣言されていないルートはテストで検出される。
var routePermissions = map[string]routeSpec{
	"POST /v1/auth/signup":                 {level: levelNoFormScope},
	"POST /v1/auth/login":                  {level: levelNoFormScope},
	"POST /v1/auth/logout":                 {level: levelNoFormScope},
	"POST /v1/auth/verify-email":           {level: levelNoFormScope},
	"POST /v1/auth/verify-email/resend":    {level: levelNoFormScope},
	"POST /v1/auth/password-reset":         {level: levelNoFormScope},
	"POST /v1/auth/password-reset/confirm": {level: levelNoFormScope},

	"GET /v1/me":            {level: levelNoFormScope},
	"PATCH /v1/me":          {level: levelNoFormScope},
	"DELETE /v1/me":         {level: levelNoFormScope},
	"PATCH /v1/me/password": {level: levelNoFormScope},

	"POST /v1/forms": {level: levelNoFormScope},
	// ListAccessibleForms でクエリ自体が認可を内包するため、フォーム単位の認可は不要。
	"GET /v1/forms": {level: levelNoFormScope},
	// 有効な招待トークンであるかを検証するため、フォーム単位の認可は不要。
	"POST /v1/invites/:invite_id/accept": {level: levelNoFormScope},

	"GET /v1/forms/:form_id": {level: levelEditor},
	// title_question_id が未指定だとハンドラが 204 で返し UseCase に到達しないため、null を送る。
	"PATCH /v1/forms/:form_id": {
		level: levelEditor,
		body:  map[string]any{"title_question_id": nil},
	},
	"GET /v1/forms/:form_id/members":   {level: levelEditor},
	"GET /v1/forms/:form_id/questions": {level: levelEditor},
	"GET /v1/forms/:form_id/statuses":  {level: levelEditor},
	"POST /v1/forms/:form_id/statuses": {
		level: levelEditor,
		body:  map[string]any{"name": "matrix", "display_order": 99},
	},
	"PATCH /v1/forms/:form_id/statuses/:status_id": {
		level: levelEditor,
		body:  map[string]any{"name": "matrix-updated"},
	},
	"DELETE /v1/forms/:form_id/statuses/:status_id": {level: levelEditor},
	"GET /v1/forms/:form_id/notification-settings":  {level: levelEditor},
	"GET /v1/forms/:form_id/stream":                 {level: levelEditor},
	"POST /v1/forms/:form_id/sync":                  {level: levelEditor},
	"GET /v1/tickets": {
		level: levelEditor,
		query: "?form_id=:form_id",
	},
	"GET /v1/tickets/:ticket_id":           {level: levelEditor},
	"PATCH /v1/tickets/:ticket_id":         {level: levelEditor, body: map[string]any{}},
	"GET /v1/tickets/:ticket_id/histories": {level: levelEditor},
	"POST /v1/tickets/:ticket_id/notifications": {
		level: levelEditor,
		body:  map[string]any{"notification_type": "status_change"},
	},

	"DELETE /v1/forms/:form_id": {level: levelAdmin},
	"POST /v1/forms/:form_id/members": {
		level: levelAdmin,
		body:  map[string]any{"email": "matrix-target@example.com", "role": "editor"},
	},
	"PUT /v1/forms/:form_id/members/:user_id": {
		level: levelAdmin,
		body:  map[string]any{"role": "editor"},
	},
	"DELETE /v1/forms/:form_id/members/:user_id": {level: levelAdmin},
	"GET /v1/forms/:form_id/invites":             {level: levelAdmin},
	"POST /v1/forms/:form_id/invites": {
		level: levelAdmin,
		body:  map[string]any{"email": "matrix-invitee@example.com", "role": "editor"},
	},
	"DELETE /v1/forms/:form_id/invites/:invite_id": {level: levelAdmin},
	"PATCH /v1/forms/:form_id/notification-settings": {
		level: levelAdmin,
		body: map[string]any{"settings": []map[string]any{
			{"notification_type": "status_change", "mode": "off", "include_detail": false},
		}},
	},
}

type permissionFixture struct {
	formID       string
	statusID     string
	ticketID     string
	memberUserID string
	inviteID     string
}

func newPermissionFixture(
	t *testing.T,
	adminUserID, editorUserID uuid.UUID,
) permissionFixture {
	t.Helper()
	ctx := context.Background()

	// forms.form_id はグローバルに一意。他はフォーム単位の制約なので固定値でよい。
	formID, statusID := testutil.CreateForm(
		t, ctx, testPool, uuid.NewString(), "権限マトリクス用フォーム", adminUserID,
	)
	testutil.AddMember(t, ctx, testPool, editorUserID, formID, entity.RoleEditor)
	ticketID := testutil.CreateTicket(t, ctx, testPool, formID, statusID, "matrix-response")
	inviteID := testutil.CreateInvite(
		t, ctx, testPool, formID, adminUserID, "matrix-invite@example.com", entity.RoleEditor,
	)

	return permissionFixture{
		formID:       formID.String(),
		statusID:     statusID.String(),
		ticketID:     ticketID.String(),
		memberUserID: editorUserID.String(),
		inviteID:     inviteID.String(),
	}
}

func (f permissionFixture) resolve(pattern string) string {
	r := strings.NewReplacer(
		":form_id", f.formID,
		":status_id", f.statusID,
		":ticket_id", f.ticketID,
		":user_id", f.memberUserID,
		":invite_id", f.inviteID,
	)
	return r.Replace(pattern)
}

func requestRoute(
	t *testing.T,
	client *http.Client,
	routeKey string,
	spec routeSpec,
	fx permissionFixture,
) *http.Response {
	t.Helper()

	method, pattern, ok := strings.Cut(routeKey, " ")
	require.True(t, ok, "ルートキーの形式が不正: %s", routeKey)
	path := fx.resolve(pattern) + fx.resolve(spec.query)

	switch method {
	case http.MethodGet:
		return get(t, client, path)
	case http.MethodPost:
		return postJSON(t, client, path, spec.body)
	case http.MethodPatch:
		return patchJSON(t, client, path, spec.body)
	case http.MethodPut:
		return putJSON(t, client, path, spec.body)
	case http.MethodDelete:
		return del(t, client, path)
	default:
		t.Fatalf("未対応のメソッド: %s", method)
		return nil
	}
}

func routeKeysAt(levels ...accessLevel) []string {
	var keys []string
	for key, spec := range routePermissions {
		if slices.Contains(levels, spec.level) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func TestPermissionMatrix(t *testing.T) {
	ctx := context.Background()
	testutil.TruncateAll(t, ctx, testPool)

	adminUserID := testutil.CreateVerifiedUser(
		t, ctx, testPool, "matrix-admin@example.com", "password123", "Admin",
	)
	editorUserID := testutil.CreateVerifiedUser(
		t, ctx, testPool, "matrix-editor@example.com", "password123", "Editor",
	)

	editorClient := loginUserExisting(t, "matrix-editor@example.com", "password123")
	nonMemberClient := loginUser(t, "matrix-outsider@example.com", "password123", "Outsider")

	// SSE は認可が通ると接続が維持されるため、タイムアウトがないとテストが停止する。
	editorClient.Timeout = 10 * time.Second
	nonMemberClient.Timeout = 10 * time.Second

	t.Run("ルータに登録された全ルートの権限がテストに宣言されている", func(t *testing.T) {
		registered := map[string]bool{}
		for _, r := range testRouter.Routes() {
			key := fmt.Sprintf("%s %s", r.Method, r.Path)
			registered[key] = true
			assert.Contains(
				t, routePermissions, key,
				"ルート %s が routePermissions にない。必要な権限を宣言すること", key,
			)
		}
		for key := range routePermissions {
			assert.True(
				t, registered[key],
				"routePermissions の %s はルータに登録されていない。消したルートなら表からも消すこと", key,
			)
		}
	})

	t.Run("非メンバーはフォーム単位のルートに到達できない", func(t *testing.T) {
		for _, key := range routeKeysAt(levelEditor, levelAdmin) {
			t.Run(key, func(t *testing.T) {
				fx := newPermissionFixture(t, adminUserID, editorUserID)
				resp := requestRoute(t, nonMemberClient, key, routePermissions[key], fx)

				var body map[string]any
				readJSON(t, resp, &body)
				assert.Equal(t, http.StatusNotFound, resp.StatusCode)
				assert.Equal(t, "RESOURCE_HIDDEN", body["code"])
			})
		}
	})

	t.Run("editorはadmin専用ルートを実行できない", func(t *testing.T) {
		for _, key := range routeKeysAt(levelAdmin) {
			t.Run(key, func(t *testing.T) {
				fx := newPermissionFixture(t, adminUserID, editorUserID)
				resp := requestRoute(t, editorClient, key, routePermissions[key], fx)

				var body map[string]any
				readJSON(t, resp, &body)
				assert.Equal(t, http.StatusForbidden, resp.StatusCode)
				assert.Equal(t, "FORBIDDEN", body["code"])
			})
		}
	})
}
