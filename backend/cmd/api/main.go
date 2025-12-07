package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hiromichi-5/forma/backend/internal/api"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/db"
	gforms "github.com/hiromichi-5/forma/backend/internal/google"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
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
	viper.AutomaticEnv()
	appEnv := viper.GetString("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}
	addr := viper.GetString("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	secret := viper.GetString("APP_SECRET")
	if secret == "" {
		log.Fatal("APP_SECRET required")
	}
	pgDSN := viper.GetString("PG_DSN")
	if pgDSN == "" {
		log.Fatal("PG_DSN required")
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
		log.Fatalf("pgxpool: %v", err)
	}
	defer pool.Close()
	q := db.New(pool)

	signer := auth.Signer{Secret: []byte(secret), TTL: time.Hour}
	if signer.TTL <= 0 {
		log.Fatal("APP_JWT_TTL must be positive duration")
	}

	gf, err := gforms.NewRealFormsClient(ctx, saPath)
	if err != nil {
		log.Fatalf("forms client: %v", err)
	}

	svc := service.NewService(q, gf)

	r := NewRouter()
	cookieCfg := api.AuthCookieConfig{
		Name:     "forma_token",
		Path:     "/",
		Secure:   appEnv == "production",
		SameSite: http.SameSiteLaxMode,
	}

	ah := &api.AuthHandler{Svc: service.NewAuthService(q), JWT: signer, Cookie: cookieCfg}
	r.POST("/v1/auth/login", ah.PostV1AuthLogin)
	r.POST("/v1/auth/signup", ah.PostV1AuthSignup)
	r.POST("/v1/auth/logout", ah.PostV1AuthLogout)

	authz := r.Group("/v1")
	authz.Use(auth.BearerMiddleware(signer, cookieCfg.Name))

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
	authz.POST("/forms", fh.PostV1Forms)
	authz.GET("/forms", fh.GetV1Forms)
	authz.GET("/forms/:form_id/health", func(c *gin.Context) {
		fh.GetV1FormsFormIdHealth(c, c.Param("form_id"))
	})
	authz.POST("/forms/:form_id/sync", func(c *gin.Context) {
		fh.PostV1FormsFormIdSync(c, c.Param("form_id"))
	})
	authz.GET("/forms/:form_id/members", fh.GetV1FormsFormIdMembers)
	authz.POST("/forms/:form_id/members", fh.PostV1FormsFormIdMembers)
	authz.PUT("/forms/:form_id/members/:user_id", fh.PutV1FormsFormIdMembersUserId)
	authz.DELETE("/forms/:form_id/members/:user_id", fh.DeleteV1FormsFormIdMembersUserId)
	authz.GET("/forms/:form_id/invites", func(c *gin.Context) {
		fh.GetV1FormsFormIdInvites(c, c.Param("form_id"))
	})
	authz.POST("/forms/:form_id/invites", func(c *gin.Context) {
		fh.PostV1FormsFormIdInvites(c, c.Param("form_id"))
	})
	authz.DELETE("/forms/:form_id/invites/:code", func(c *gin.Context) {
		fh.DeleteV1FormsFormIdInvitesCode(c, c.Param("form_id"), c.Param("code"))
	})
	authz.GET("/forms/:form_id/questions", func(c *gin.Context) {
		fh.GetV1FormsFormIdQuestions(c, c.Param("form_id"))
	})
	authz.PATCH("/forms/:form_id/title-question", func(c *gin.Context) {
		fh.PatchV1FormsFormIdTitleQuestion(c, c.Param("form_id"))
	})
	authz.POST("/invites/accept", fh.PostV1InvitesAccept)

	authz.GET("/responses", fh.GetV1Responses)

	authz.GET("/tickets", fh.GetV1Tickets)
	authz.GET("/tickets/:ticket_id", func(c *gin.Context) {
		fh.GetV1TicketsTicketId(c, c.Param("ticket_id"))
	})
	authz.PATCH("/tickets/:ticket_id", func(c *gin.Context) {
		fh.PatchV1TicketsTicketId(c, c.Param("ticket_id"))
	})

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
