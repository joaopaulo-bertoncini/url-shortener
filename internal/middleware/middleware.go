package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	metrics "github.com/joaopaulo-bertoncini/url-shortener/internal/metrics"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			metrics.InvalidTokens.Inc()
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		providedToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		expectedToken := os.Getenv("AUTH_TOKEN")

		if providedToken != expectedToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Next()
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Processa a requisição
		c.Next()

		duration := time.Since(start).Seconds()
		path := c.FullPath()
		status := strconv.Itoa(c.Writer.Status())
		size := float64(c.Writer.Size())

		if path == "" {
			path = c.Request.URL.Path // fallback
		}

		// Métricas
		metrics.UrlRequests.WithLabelValues(path).Observe(duration)
		metrics.HttpErrors.WithLabelValues(path, status).Add(1)
		metrics.ResponseSize.WithLabelValues(path).Observe(size)
	}
}
