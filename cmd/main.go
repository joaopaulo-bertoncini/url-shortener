package main

import (
	"os"

	"github.com/gin-gonic/gin"

	handler "github.com/joaopaulo-bertoncini/url-shortener/internal/handler"
	logger "github.com/joaopaulo-bertoncini/url-shortener/internal/logger"
	middleware "github.com/joaopaulo-bertoncini/url-shortener/internal/middleware"
	repo "github.com/joaopaulo-bertoncini/url-shortener/internal/repository"
)

func main() {
	if err := logger.InitLogger(); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := repo.InitClients(); err != nil {
		logger.Log.Fatalf("failed to init clients: %v", err)
	}

	r := gin.Default()

	r.GET("/:shortID", handler.HandleRedirect)
	r.GET("/stats/:shortID", handler.HandleStats)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/shorten", handler.HandleShorten)
	protected.DELETE("/short/:shortID", handler.HandleDelete)

	logger.Log.Infof("🚀 Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		logger.Log.Fatalf("could not run server: %v", err)
	}
}
