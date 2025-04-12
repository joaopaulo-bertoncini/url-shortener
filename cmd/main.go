package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	handler "github.com/joaopaulo-bertoncini/url-shortener/internal/handler"
	logger "github.com/joaopaulo-bertoncini/url-shortener/internal/logger"
	metrics "github.com/joaopaulo-bertoncini/url-shortener/internal/metrics"
	middleware "github.com/joaopaulo-bertoncini/url-shortener/internal/middleware"
	repo "github.com/joaopaulo-bertoncini/url-shortener/internal/repository"
	telemetry "github.com/joaopaulo-bertoncini/url-shortener/internal/telemetry"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or could not load it")
	}
}

func main() {
	ctx := context.Background()
	cleanup := telemetry.InitTracer(ctx)
	defer cleanup()

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
	r.Use(middleware.MetricsMiddleware())

	metrics.InitCustomMetrics()

	r.GET("/:shortID", handler.HandleRedirect)
	r.GET("/stats/:shortID", handler.HandleStats)
	r.GET("/metrics", handler.HandleMetrics)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/shorten", handler.HandleShorten)
	protected.DELETE("/short/:shortID", handler.HandleDelete)

	logger.Log.Infof("🚀 Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		logger.Log.Fatalf("could not run server: %v", err)
	}
}
