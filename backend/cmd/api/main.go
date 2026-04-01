package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/app"
	"github.com/hiromichi-5/forma/backend/internal/infra/google"
	"github.com/jackc/pgx/v5/pgxpool"
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

	allowedOrigins := viper.GetString("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173"
	}

	r := app.NewRouter(
		app.Deps{Pool: pool, Fetcher: fetcher},
		app.Option{
			CookieSecure:   appEnv == "production",
			AllowedOrigins: strings.Split(allowedOrigins, ","),
		},
	)

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if appEnv != "production" {
		r.StaticFile("/openapi.yaml", "openapi/openapi.yaml")
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/openapi.yaml"),
			ginSwagger.DocExpansion("none"),
		))
	}

	return r.Run(addr)
}
