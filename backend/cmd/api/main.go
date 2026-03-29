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
	"github.com/hiromichi-5/forma/backend/internal/api"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	gforms "github.com/hiromichi-5/forma/backend/internal/google"
	"github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
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

	gf, err := gforms.NewRealFormsClient(ctx, saPath)
	if err != nil {
		return fmt.Errorf("forms client: %w", err)
	}

	q := db.New(pool)

	svc := service.NewService(q, gf)

	r := NewRouter()
	cookieCfg := api.AuthCookieConfig{
		Name:     "forma_token",
		Path:     "/",
		Secure:   appEnv == "production",
		SameSite: http.SameSiteLaxMode,
	}

	ah := &api.AuthHandler{Svc: service.NewAuthService(q), Cookie: cookieCfg}
	r.POST("/v1/auth/login", ah.PostV1AuthLogin)
	r.POST("/v1/auth/signup", ah.PostV1AuthSignup)
	r.POST("/v1/auth/logout", ah.PostV1AuthLogout)
	r.POST("/v1/auth/verify-email", ah.PostV1AuthVerifyEmail)
	r.POST("/v1/auth/verify-email/resend", ah.PostV1AuthVerifyEmailResend)
	r.POST("/v1/auth/password-reset", ah.PostV1AuthPasswordReset)
	r.POST("/v1/auth/password-reset/confirm", ah.PostV1AuthPasswordResetConfirm)

	authz := r.Group("/v1")
	authz.Use(auth.SessionMiddleware(q, cookieCfg.Name))

	ph := &api.ProfileHandler{Svc: service.NewProfileService(q)}
	authz.GET("/me", ph.GetV1Me)
	authz.PATCH("/me", ph.PatchV1Me)
	authz.DELETE("/me", ph.DeleteV1Me)
	authz.PATCH("/me/password", ph.PatchV1MePassword)

	authz.GET("/whoami", func(c *gin.Context) {
		if uid, ok := auth.UserID(c); ok {
			c.JSON(http.StatusOK, gin.H{"user_id": uid})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
	})

	fh := &api.FormsHandler{S: svc}
	mh := &api.MembersHandler{Svc: svc}
	ih := &api.InvitesHandler{Svc: svc}
	sh := &api.StatusesHandler{Svc: svc}
	th := &api.TicketHistoriesHandler{Svc: svc}
	authz.POST("/forms", fh.PostV1Forms)
	authz.GET("/forms", fh.GetV1Forms)
	authz.GET("/forms/:form_id", func(c *gin.Context) {
		fh.GetV1FormsId(c, c.Param("form_id"))
	})
	authz.PATCH("/forms/:form_id", func(c *gin.Context) {
		fh.PatchV1FormsId(c, c.Param("form_id"))
	})
	authz.POST("/forms/:form_id/sync", func(c *gin.Context) {
		fh.PostV1FormsFormIdSync(c, c.Param("form_id"))
	})
	authz.GET("/forms/:form_id/members", mh.GetV1FormsFormIdMembers)
	authz.POST("/forms/:form_id/members", mh.PostV1FormsFormIdMembers)
	authz.PUT("/forms/:form_id/members/:user_id", mh.PutV1FormsFormIdMembersUserId)
	authz.DELETE("/forms/:form_id/members/:user_id", mh.DeleteV1FormsFormIdMembersUserId)
	authz.GET("/forms/:form_id/invites", ih.GetV1FormsFormIdInvites)
	authz.POST("/forms/:form_id/invites", ih.PostV1FormsFormIdInvites)
	authz.DELETE("/forms/:form_id/invites/:invite_id", ih.DeleteV1FormsFormIdInvitesInviteId)
	authz.GET("/forms/:form_id/statuses", func(c *gin.Context) {
		sh.GetV1FormsIdStatuses(c, c.Param("form_id"))
	})
	authz.POST("/forms/:form_id/statuses", func(c *gin.Context) {
		sh.PostV1FormsIdStatuses(c, c.Param("form_id"))
	})
	authz.PATCH("/forms/:form_id/statuses/:status_id", func(c *gin.Context) {
		sh.PatchV1FormsIdStatusesStatusId(c, c.Param("form_id"), c.Param("status_id"))
	})
	authz.POST("/forms/:form_id/statuses/:status_id/default", func(c *gin.Context) {
		sh.PostV1FormsIdStatusesStatusIdDefault(c, c.Param("form_id"), c.Param("status_id"))
	})
	authz.DELETE("/forms/:form_id/statuses/:status_id", func(c *gin.Context) {
		sh.DeleteV1FormsIdStatusesStatusId(c, c.Param("form_id"), c.Param("status_id"))
	})
	authz.GET("/forms/:form_id/questions", func(c *gin.Context) {
		fh.GetV1FormsFormIdQuestions(c, c.Param("form_id"))
	})
	authz.POST("/invites/:invite_id/accept", func(c *gin.Context) {
		ih.PostV1InvitesInviteIdAccept(c, c.Param("invite_id"))
	})

	authz.GET("/tickets", fh.GetV1Tickets)
	authz.GET("/tickets/:ticket_id", func(c *gin.Context) {
		fh.GetV1TicketsTicketId(c, c.Param("ticket_id"))
	})
	authz.PATCH("/tickets/:ticket_id", func(c *gin.Context) {
		fh.PatchV1TicketsTicketId(c, c.Param("ticket_id"))
	})
	authz.GET("/tickets/:ticket_id/histories", func(c *gin.Context) {
		th.GetV1TicketsTicketIdHistories(c, c.Param("ticket_id"))
	})

	return r.Run(addr)
}
