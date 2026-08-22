package app

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/infra/postgres"
	"github.com/hiromichi-5/forma/backend/internal/infra/pubsub"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/handler"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool            *pgxpool.Pool
	Fetcher         repository.FormFetcher
	EmailSender     repository.EmailSender
	FrontendBaseURL string
}

type Option struct {
	CookieSecure   bool
	CookieDomain   string
	AllowedOrigins []string
}

func NewRouter(deps Deps, opt Option) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	config := cors.DefaultConfig()
	if len(opt.AllowedOrigins) > 0 {
		config.AllowOrigins = opt.AllowedOrigins
	} else {
		config.AllowOrigins = []string{"http://localhost:5173"}
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	userRepo := postgres.NewUserRepository(deps.Pool)
	sessionRepo := postgres.NewSessionRepository(deps.Pool)
	emailTokenRepo := postgres.NewEmailVerificationTokenRepository(deps.Pool)
	resetTokenRepo := postgres.NewPasswordResetTokenRepository(deps.Pool)
	formRepo := postgres.NewFormRepository(deps.Pool)
	memberRepo := postgres.NewMemberRepository(deps.Pool)
	statusRepo := postgres.NewStatusRepository(deps.Pool)
	ticketRepo := postgres.NewTicketRepository(deps.Pool)
	inviteRepo := postgres.NewInviteRepository(deps.Pool)
	notificationRepo := postgres.NewNotificationRepository(deps.Pool)

	authUC := usecase.NewAuthUseCase(
		userRepo,
		sessionRepo,
		emailTokenRepo,
		resetTokenRepo,
		postgres.NewAuthUoW(deps.Pool),
		deps.EmailSender,
		deps.FrontendBaseURL,
	)
	profileUC := usecase.NewProfileUseCase(userRepo)
	syncUC := usecase.NewSyncUseCase(formRepo, ticketRepo, statusRepo, memberRepo, deps.Fetcher)
	formUC := usecase.NewFormUseCase(
		formRepo, memberRepo, statusRepo, deps.Fetcher,
		postgres.NewFormUoW(deps.Pool),
		syncUC,
	)
	memberUC := usecase.NewMemberUseCase(memberRepo, userRepo)
	inviteUC := usecase.NewInviteUseCase(
		inviteRepo, memberRepo, userRepo,
		postgres.NewInviteUoW(deps.Pool),
		deps.EmailSender, deps.FrontendBaseURL,
	)
	statusUC := usecase.NewStatusUseCase(
		statusRepo, memberRepo, ticketRepo,
		postgres.NewStatusUoW(deps.Pool),
	)
	notificationUC := usecase.NewNotificationUseCase(
		notificationRepo, ticketRepo, formRepo, statusRepo, memberRepo, userRepo,
		postgres.NewNotificationUoW(deps.Pool),
		deps.EmailSender,
	)
	hub := pubsub.NewMemoryHub()
	ticketUC := usecase.NewTicketUseCase(
		ticketRepo, formRepo, statusRepo, memberRepo, userRepo,
		postgres.NewTicketUoW(deps.Pool),
		hub,
		notificationUC,
	)

	cookieCfg := handler.CookieConfig{
		Name:     "forma_token",
		Path:     "/",
		Domain:   opt.CookieDomain,
		Secure:   opt.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	ah := handler.NewAuthHandler(authUC, cookieCfg)
	ph := handler.NewProfileHandler(profileUC)
	fh := handler.NewFormHandler(formUC)
	mh := handler.NewMemberHandler(memberUC)
	ih := handler.NewInviteHandler(inviteUC)
	sh := handler.NewStatusHandler(statusUC)
	tkh := handler.NewTicketHandler(ticketUC)
	thh := handler.NewTicketHistoryHandler(ticketUC)
	sseh := handler.NewStreamHandler(ticketUC, hub)
	syh := handler.NewSyncHandler(syncUC)
	nh := handler.NewNotificationHandler(notificationUC)

	r.POST("/v1/auth/login", ah.PostV1AuthLogin)
	r.POST("/v1/auth/signup", ah.PostV1AuthSignup)
	r.POST("/v1/auth/logout", ah.PostV1AuthLogout)
	r.POST("/v1/auth/verify-email", ah.PostV1AuthVerifyEmail)
	r.POST("/v1/auth/verify-email/resend", ah.PostV1AuthVerifyEmailResend)
	r.POST("/v1/auth/password-reset", ah.PostV1AuthPasswordReset)
	r.POST("/v1/auth/password-reset/confirm", ah.PostV1AuthPasswordResetConfirm)

	authz := r.Group("/v1")
	authz.Use(middleware.SessionMiddleware(sessionRepo, cookieCfg.Name))

	authz.GET("/me", ph.GetV1Me)
	authz.PATCH("/me", ph.PatchV1Me)
	authz.DELETE("/me", ph.DeleteV1Me)
	authz.PATCH("/me/password", ph.PatchV1MePassword)

	authz.POST("/forms", fh.PostV1Forms)
	authz.GET("/forms", fh.GetV1Forms)
	authz.GET("/forms/:form_id", fh.GetV1FormsId)
	authz.PATCH("/forms/:form_id", fh.PatchV1FormsId)
	authz.DELETE("/forms/:form_id", fh.DeleteV1FormsId)
	authz.POST("/forms/:form_id/sync", syh.PostV1FormsFormIdSync)
	authz.GET("/forms/:form_id/members", mh.GetV1FormsFormIdMembers)
	authz.POST("/forms/:form_id/members", mh.PostV1FormsFormIdMembers)
	authz.PUT("/forms/:form_id/members/:user_id", mh.PutV1FormsFormIdMembersUserId)
	authz.DELETE("/forms/:form_id/members/:user_id", mh.DeleteV1FormsFormIdMembersUserId)
	authz.GET("/forms/:form_id/invites", ih.GetV1FormsFormIdInvites)
	authz.POST("/forms/:form_id/invites", ih.PostV1FormsFormIdInvites)
	authz.DELETE("/forms/:form_id/invites/:invite_id", ih.DeleteV1FormsFormIdInvitesInviteId)
	authz.GET("/forms/:form_id/statuses", sh.GetV1FormsIdStatuses)
	authz.POST("/forms/:form_id/statuses", sh.PostV1FormsIdStatuses)
	authz.PATCH("/forms/:form_id/statuses/:status_id", sh.PatchV1FormsIdStatusesStatusId)
	authz.DELETE("/forms/:form_id/statuses/:status_id", sh.DeleteV1FormsIdStatusesStatusId)
	authz.GET("/forms/:form_id/questions", fh.GetV1FormsFormIdQuestions)
	authz.GET(
		"/forms/:form_id/notification-settings",
		nh.GetV1FormsFormIdNotificationSettings,
	)
	authz.PATCH(
		"/forms/:form_id/notification-settings",
		nh.PatchV1FormsFormIdNotificationSettings,
	)
	authz.GET("/forms/:form_id/stream", sseh.GetV1FormsFormIdStream)
	authz.POST("/invites/:invite_id/accept", ih.PostV1InvitesInviteIdAccept)

	authz.GET("/tickets", tkh.GetV1Tickets)
	authz.GET("/tickets/:ticket_id", tkh.GetV1TicketsTicketId)
	authz.PATCH("/tickets/:ticket_id", tkh.PatchV1TicketsTicketId)
	authz.GET("/tickets/:ticket_id/histories", thh.GetV1TicketsTicketIdHistories)
	authz.POST("/tickets/:ticket_id/notifications", nh.PostV1TicketsTicketIdNotifications)

	return r
}
