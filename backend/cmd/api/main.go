package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/app"
	"github.com/hiromichi-5/forma/backend/internal/infra/google"
	"github.com/hiromichi-5/forma/backend/internal/infra/ses"
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

	sesFrom := viper.GetString("SES_FROM_ADDRESS")
	if sesFrom == "" {
		return fmt.Errorf("SES_FROM_ADDRESS required")
	}
	sesReplyTo := viper.GetString("SES_REPLY_TO_ADDRESS")

	awsRegion := viper.GetString("AWS_REGION")
	var awsOpts []func(*awsconfig.LoadOptions) error
	if awsRegion != "" {
		awsOpts = append(awsOpts, awsconfig.WithRegion(awsRegion))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsOpts...)
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	var replyTo []string
	if sesReplyTo != "" {
		replyTo = []string{sesReplyTo}
	}
	emailSender := ses.NewEmailSender(awsCfg, sesFrom, replyTo)

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

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	swaggerUser := viper.GetString("SWAGGER_USER")
	swaggerPassword := viper.GetString("SWAGGER_PASSWORD")
	if appEnv != "production" || (swaggerUser != "" && swaggerPassword != "") {
		docs := r.Group("/")
		if swaggerUser != "" && swaggerPassword != "" {
			docs.Use(gin.BasicAuth(gin.Accounts{swaggerUser: swaggerPassword}))
		}
		docs.GET("/openapi.yaml", func(c *gin.Context) {
			c.File("openapi/openapi.yaml")
		})
		docs.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/openapi.yaml"),
			ginSwagger.DocExpansion("list"),
		))
	}

	slog.Info("server listening", "addr", addr)
	return r.Run(addr)
}
