package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	openapispec "github.com/hiromichi-5/forma/backend/internal/api"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/openapi.yaml", func(c *gin.Context) {
		swagger, err := openapispec.GetSwagger()
		if err != nil {
			log.Printf("failed to load embedded openapi spec: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(swagger)
		if err != nil {
			log.Printf("failed to marshal embedded openapi spec: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Data(http.StatusOK, "application/json; charset=utf-8", data)
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/openapi.yaml"),
		ginSwagger.DocExpansion("list"),
	))

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
