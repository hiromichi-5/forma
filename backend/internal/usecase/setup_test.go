package usecase_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/infra/postgres"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	pool, cleanup := testutil.SetupPostgres(ctx)
	testPool = pool

	code := m.Run()

	cleanup()
	os.Exit(code)
}

func newUserRepo() repository.UserRepository     { return postgres.NewUserRepository(testPool) }
func newFormRepo() repository.FormRepository     { return postgres.NewFormRepository(testPool) }
func newMemberRepo() repository.MemberRepository { return postgres.NewMemberRepository(testPool) }
func newStatusRepo() repository.StatusRepository { return postgres.NewStatusRepository(testPool) }
func newTicketRepo() repository.TicketRepository { return postgres.NewTicketRepository(testPool) }
func newInviteRepo() repository.InviteRepository { return postgres.NewInviteRepository(testPool) }

func newAuthUseCase() *usecase.AuthUseCase {
	return usecase.NewAuthUseCase(
		newUserRepo(), postgres.NewAuthUoW(testPool),
		&mockEmailSender{}, "http://localhost:5173",
	)
}

func newProfileUseCase() *usecase.ProfileUseCase {
	return usecase.NewProfileUseCase(newUserRepo())
}

func newFormUseCase(fetcher repository.FormFetcher) *usecase.FormUseCase {
	return newFormUseCaseWithSyncer(fetcher, newSyncUseCase(fetcher))
}

func newFormUseCaseWithSyncer(
	fetcher repository.FormFetcher,
	syncer usecase.FormSyncer,
) *usecase.FormUseCase {
	return usecase.NewFormUseCase(
		newFormRepo(), newMemberRepo(), newStatusRepo(), fetcher,
		postgres.NewFormUoW(testPool),
		syncer,
	)
}

func newMemberUseCase() *usecase.MemberUseCase {
	return usecase.NewMemberUseCase(newMemberRepo(), newUserRepo())
}

func newInviteUseCase() *usecase.InviteUseCase {
	return usecase.NewInviteUseCase(
		newInviteRepo(), newMemberRepo(), newUserRepo(),
		postgres.NewInviteUoW(testPool),
		&mockEmailSender{}, "http://localhost:5173",
	)
}

func newStatusUseCase() *usecase.StatusUseCase {
	return usecase.NewStatusUseCase(
		newStatusRepo(), newMemberRepo(), newTicketRepo(),
		postgres.NewStatusUoW(testPool),
	)
}

func newTicketUseCase() *usecase.TicketUseCase {
	return newTicketUseCaseWithPublisher(&noopEventPublisher{})
}

func newTicketUseCaseWithPublisher(publisher usecase.EventPublisher) *usecase.TicketUseCase {
	return usecase.NewTicketUseCase(
		newTicketRepo(), newFormRepo(), newStatusRepo(), newMemberRepo(), newUserRepo(),
		postgres.NewTicketUoW(testPool),
		publisher,
	)
}

func newSyncUseCase(fetcher repository.FormFetcher) *usecase.SyncUseCase {
	return usecase.NewSyncUseCase(
		newFormRepo(), newTicketRepo(), newStatusRepo(), newMemberRepo(), fetcher,
	)
}

type mockEmailSender struct {
	sendEmailFunc func(ctx context.Context, input repository.SendEmailInput) error
}

func (m *mockEmailSender) SendEmail(ctx context.Context, input repository.SendEmailInput) error {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(ctx, input)
	}
	return nil
}

type mockFormFetcher struct {
	getFormFunc       func(ctx context.Context, formID string) (*repository.GoogleForm, error)
	listResponsesFunc func(ctx context.Context, formID, filter, pageToken string) (*repository.GoogleFormResponsePage, error)
}

func (m *mockFormFetcher) GetForm(
	ctx context.Context,
	formID string,
) (*repository.GoogleForm, error) {
	if m.getFormFunc != nil {
		return m.getFormFunc(ctx, formID)
	}
	return nil, repository.ErrNotFound
}

func (m *mockFormFetcher) ListResponses(
	ctx context.Context,
	formID, filter, pageToken string,
) (*repository.GoogleFormResponsePage, error) {
	if m.listResponsesFunc != nil {
		return m.listResponsesFunc(ctx, formID, filter, pageToken)
	}
	return &repository.GoogleFormResponsePage{}, nil
}

type mockFormSyncer struct {
	syncFormOnceFunc func(ctx context.Context, formID, userID uuid.UUID) (int, time.Time, error)
}

func (m *mockFormSyncer) SyncFormOnce(
	ctx context.Context,
	formID, userID uuid.UUID,
) (int, time.Time, error) {
	if m.syncFormOnceFunc != nil {
		return m.syncFormOnceFunc(ctx, formID, userID)
	}
	return 0, time.Time{}, nil
}

func truncate(t *testing.T) {
	t.Helper()
	testutil.TruncateAll(t, context.Background(), testPool)
}

type noopEventPublisher struct{}

func (n *noopEventPublisher) PublishTicketUpdated(_ context.Context, _ usecase.TicketEvent) error {
	return nil
}
