package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	openapispec "github.com/hiromichi-5/forma/backend/internal/api"
	"github.com/hiromichi-5/forma/backend/internal/app"
	"github.com/hiromichi-5/forma/backend/internal/infra/google"
	"github.com/hiromichi-5/forma/backend/internal/infra/resend"
	"github.com/hiromichi-5/forma/backend/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}
	if appEnv != "production" {
		_ = godotenv.Load(".env.local")
	}
	viper.AutomaticEnv()
	logLevel := viper.GetString("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	logger.Setup(appEnv, logLevel)
	slog.Info("application startup initiated", "env", appEnv, "log_level", logLevel) //nolint:gosec

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
	slog.Info("database connection established")

	fetcher, err := google.NewFormClient(ctx, saPath)
	if err != nil {
		return fmt.Errorf("forms client: %w", err)
	}

	frontendBaseURL := viper.GetString("FRONTEND_BASE_URL")
	if frontendBaseURL == "" {
		frontendBaseURL = "http://localhost:5173"
	}

	resendAPIKey := viper.GetString("RESEND_API_KEY")
	if resendAPIKey == "" {
		return fmt.Errorf("RESEND_API_KEY required")
	}
	resendFrom := viper.GetString("RESEND_FROM_ADDRESS")
	if resendFrom == "" {
		return fmt.Errorf("RESEND_FROM_ADDRESS required")
	}
	resendReplyTo := viper.GetString("RESEND_REPLY_TO_ADDRESS")

	emailSender := resend.NewEmailSender(resendAPIKey, resendFrom, resendReplyTo)

	allowedOrigins := viper.GetString("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173"
	}

	cookieDomain := viper.GetString("COOKIE_DOMAIN")

	r := app.NewRouter(
		app.Deps{
			Pool:            pool,
			Fetcher:         fetcher,
			EmailSender:     emailSender,
			FrontendBaseURL: frontendBaseURL,
		},
		app.Option{
			CookieSecure:   appEnv == "production",
			CookieDomain:   cookieDomain,
			AllowedOrigins: strings.Split(allowedOrigins, ","),
		},
	)

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	swaggerUser := viper.GetString("SWAGGER_USER")
	swaggerPassword := viper.GetString("SWAGGER_PASSWORD")
	if appEnv != "production" || (swaggerUser != "" && swaggerPassword != "") {
		swaggerAuth := func(c *gin.Context) {
			if swaggerUser != "" && swaggerPassword != "" {
				gin.BasicAuth(gin.Accounts{swaggerUser: swaggerPassword})(c)
			}
		}
		r.GET("/openapi.yaml", swaggerAuth, func(c *gin.Context) {
			swagger, err := openapispec.GetSwagger()
			if err != nil {
				slog.Error("failed to load embedded openapi spec", "error", err)
				c.Status(http.StatusInternalServerError)
				return
			}

			data, err := json.Marshal(swagger)
			if err != nil {
				slog.Error("failed to marshal embedded openapi spec", "error", err)
				c.Status(http.StatusInternalServerError)
				return
			}

			c.Data(http.StatusOK, "application/json; charset=utf-8", data)
		})
		r.GET("/swagger/*any", swaggerAuth, ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/openapi.yaml"),
			ginSwagger.DocExpansion("list"),
		))
	}

	slog.Info("server listening", "addr", addr)
	return r.Run(addr)
}
