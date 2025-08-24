package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hiromichi-5/forma/backend/internal/api"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

func NewRouter() *gin.Engine {
	appEnv := viper.GetString("APP_ENV")
	println("APP_ENV:", viper.GetString("APP_ENV"))

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

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

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer pool.Close()
	q := db.New(pool)

	authSvc := service.NewAuthService(q)
	signer := auth.Signer{Secret: []byte(secret), TTL: time.Hour}

	r := NewRouter()

	ah := &api.AuthHandler{Svc: authSvc, JWT: signer}
	r.POST("/v1/auth/login", ah.PostV1AuthLogin)

	authz := r.Group("/v1")
	authz.Use(auth.BearerMiddleware(signer))
	authz.GET("/whoami", func(c *gin.Context) {
		if uid, ok := auth.UserID(c); ok {
			c.JSON(http.StatusOK, gin.H{"user_id": uid})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
	})

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
