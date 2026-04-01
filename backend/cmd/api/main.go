package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/infra/google"
	"github.com/hiromichi-5/forma/backend/internal/infra/postgres"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/handler"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	config := cors.DefaultConfig()
	allowedOrigins := viper.GetString("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173"
	}
	config.AllowOrigins = strings.Split(allowedOrigins, ",")
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	appEnv := viper.GetString("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}
	if appEnv != "production" {
		r.StaticFile("/openapi.yaml", "openapi/openapi.yaml")
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/openapi.yaml"),
			ginSwagger.DocExpansion("none"),
		))
	}

	r.GET("/healthz", healthz)
	return r
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	viper.AutomaticEnv()
	appEnv := viper.GetString("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}
	addr := viper.GetString("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	pgDSN := viper.GetString("PG_DSN")
	if pgDSN == "" {
		return fmt.Errorf("PG_DSN required")
	}
	saPath := os.Getenv("GOOGLE_SERVICE_ACCOUNT_PATH")
	if saPath == "" {
		saPath = "/run/secrets/google_sa.json"

		if _, err := os.Stat(saPath); os.IsNotExist(err) {
			saPath = "./secrets/google_sa.json"
		}
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		return fmt.Errorf("pgxpool: %w", err)
	}
	defer pool.Close()

	fetcher, err := google.NewFormClient(ctx, saPath)
	if err != nil {
		return fmt.Errorf("forms client: %w", err)
	}

	userRepo := postgres.NewUserRepository(pool)
	formRepo := postgres.NewFormRepository(pool)
	memberRepo := postgres.NewMemberRepository(pool)
	statusRepo := postgres.NewStatusRepository(pool)
	ticketRepo := postgres.NewTicketRepository(pool)
	inviteRepo := postgres.NewInviteRepository(pool)

	authUC := usecase.NewAuthUseCase(userRepo, postgres.NewAuthUoW(pool))
	profileUC := usecase.NewProfileUseCase(userRepo)
	formUC := usecase.NewFormUseCase(
		formRepo,
		memberRepo,
		statusRepo,
		fetcher,
		postgres.NewFormUoW(pool),
	)
	memberUC := usecase.NewMemberUseCase(memberRepo, userRepo)
	inviteUC := usecase.NewInviteUseCase(
		inviteRepo,
		memberRepo,
		userRepo,
		postgres.NewInviteUoW(pool),
	)
	statusUC := usecase.NewStatusUseCase(
		statusRepo,
		memberRepo,
		ticketRepo,
		postgres.NewStatusUoW(pool),
	)
	ticketUC := usecase.NewTicketUseCase(
		ticketRepo,
		formRepo,
		statusRepo,
		memberRepo,
		userRepo,
		postgres.NewTicketUoW(pool),
	)
	syncUC := usecase.NewSyncUseCase(formRepo, ticketRepo, statusRepo, memberRepo, fetcher)

	cookieCfg := handler.CookieConfig{
		Name:     "forma_token",
		Path:     "/",
		Secure:   appEnv == "production",
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
	syh := handler.NewSyncHandler(syncUC)

	r := NewRouter()

	r.POST("/v1/auth/login", ah.PostV1AuthLogin)
	r.POST("/v1/auth/signup", ah.PostV1AuthSignup)
	r.POST("/v1/auth/logout", ah.PostV1AuthLogout)
	r.POST("/v1/auth/verify-email", ah.PostV1AuthVerifyEmail)
	r.POST("/v1/auth/verify-email/resend", ah.PostV1AuthVerifyEmailResend)
	r.POST("/v1/auth/password-reset", ah.PostV1AuthPasswordReset)
	r.POST("/v1/auth/password-reset/confirm", ah.PostV1AuthPasswordResetConfirm)

	authz := r.Group("/v1")
	authz.Use(middleware.SessionMiddleware(userRepo, cookieCfg.Name))

	authz.GET("/me", ph.GetV1Me)
	authz.PATCH("/me", ph.PatchV1Me)
	authz.DELETE("/me", ph.DeleteV1Me)
	authz.PATCH("/me/password", ph.PatchV1MePassword)

	authz.GET("/whoami", func(c *gin.Context) {
		if uid, ok := middleware.UserID(c); ok {
			c.JSON(http.StatusOK, gin.H{"user_id": uid})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
	})

	authz.POST("/forms", fh.PostV1Forms)
	authz.GET("/forms", fh.GetV1Forms)
	authz.GET("/forms/:form_id", fh.GetV1FormsId)
	authz.PATCH("/forms/:form_id", fh.PatchV1FormsId)
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
	authz.POST("/invites/:invite_id/accept", ih.PostV1InvitesInviteIdAccept)

	authz.GET("/tickets", tkh.GetV1Tickets)
	authz.GET("/tickets/:ticket_id", tkh.GetV1TicketsTicketId)
	authz.PATCH("/tickets/:ticket_id", tkh.PatchV1TicketsTicketId)
	authz.GET("/tickets/:ticket_id/histories", thh.GetV1TicketsTicketIdHistories)

	return r.Run(addr)
}
