package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", healthz)
	return r
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func main() {
	r := NewRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
