package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/app"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	testPool        *pgxpool.Pool
	testServer      *httptest.Server
	testRouter      *gin.Engine
	mockFetcher     *MockFormFetcher
	mockEmailSender *MockEmailSender
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, cleanup := testutil.SetupPostgres(ctx)
	testPool = pool

	mockFetcher = &MockFormFetcher{}
	mockEmailSender = &MockEmailSender{}
	testRouter = app.NewRouter(
		app.Deps{
			Pool:            pool,
			Fetcher:         mockFetcher,
			EmailSender:     mockEmailSender,
			FrontendBaseURL: "http://localhost:5173",
		},
		app.Option{CookieSecure: false},
	)
	testServer = httptest.NewServer(testRouter)

	code := m.Run()

	testServer.Close()
	cleanup()
	os.Exit(code)
}

type MockEmailSender struct {
	SendEmailFunc func(ctx context.Context, input repository.SendEmailInput) error
}

func (m *MockEmailSender) SendEmail(ctx context.Context, input repository.SendEmailInput) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(ctx, input)
	}
	return nil
}

type MockFormFetcher struct {
	GetFormFunc       func(ctx context.Context, formID string) (*repository.GoogleForm, error)
	ListResponsesFunc func(ctx context.Context, formID, filter, pageToken string) (*repository.GoogleFormResponsePage, error)
}

func (m *MockFormFetcher) GetForm(
	ctx context.Context,
	formID string,
) (*repository.GoogleForm, error) {
	if m.GetFormFunc != nil {
		return m.GetFormFunc(ctx, formID)
	}
	return nil, repository.ErrNotFound
}

func (m *MockFormFetcher) ListResponses(
	ctx context.Context,
	formID, filter, pageToken string,
) (*repository.GoogleFormResponsePage, error) {
	if m.ListResponsesFunc != nil {
		return m.ListResponsesFunc(ctx, formID, filter, pageToken)
	}
	return &repository.GoogleFormResponsePage{}, nil
}

func loginUser(t *testing.T, email, password, displayName string) *http.Client {
	t.Helper()
	ctx := context.Background()

	testutil.CreateVerifiedUser(t, ctx, testPool, email, password, displayName)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}

	resp := postJSON(t, client, "/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	return client
}

func loginUserExisting(t *testing.T, email, password string) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}

	resp := postJSON(t, client, "/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	return client
}

func postJSON(t *testing.T, client *http.Client, path string, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, testServer.URL+path, strings.NewReader(string(b)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func patchJSON(t *testing.T, client *http.Client, path string, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPatch, testServer.URL+path, strings.NewReader(string(b)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func get(t *testing.T, client *http.Client, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, testServer.URL+path, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func putJSON(t *testing.T, client *http.Client, path string, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPut, testServer.URL+path, strings.NewReader(string(b)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func del(t *testing.T, client *http.Client, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, testServer.URL+path, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func readJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, v))
}
