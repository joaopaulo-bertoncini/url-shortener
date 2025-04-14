package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/url-shortener/internal/metrics"
	"github.com/joaopaulo-bertoncini/url-shortener/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type reqBody struct {
	URL string `json:"url" binding:"required,url"`
}

var tracer = otel.Tracer("url-shortener/handler")

func HandleShorten(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "HandleShorten")
	defer span.End()

	var body reqBody
	if err := c.ShouldBindJSON(&body); err != nil {
		span.SetStatus(codes.Error, "invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	shortURL, err := service.ShortenURL(ctx, body.URL)
	if err != nil {
		span.SetStatus(codes.Error, "failed to shorten URL")
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.ShortenCounter.Inc()
	span.SetAttributes(attribute.String("original_url", body.URL), attribute.String("short_url", shortURL))
	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
}

func HandleRedirect(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "HandleRedirect")
	defer span.End()

	shortID := c.Param("shortID")
	longURL, err := service.ResolveShortID(ctx, shortID)
	if err != nil {
		span.SetStatus(codes.Error, "short ID not found")
		span.RecordError(err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	metrics.RedirectCounter.Inc()
	span.SetAttributes(attribute.String("short_id", shortID), attribute.String("redirect_url", longURL))
	//c.JSON(http.StatusOK, gin.H{"target": longURL})
	c.Redirect(http.StatusMovedPermanently, longURL)
}

func HandleDelete(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "HandleDelete")
	defer span.End()

	shortID := c.Param("shortID")

	err := service.DeleteShortID(ctx, shortID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to delete short ID")
		span.RecordError(err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("short_id", shortID))
	c.JSON(http.StatusOK, gin.H{"message": "short URL deleted successfully"})
}

func HandleStats(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "HandleStats")
	defer span.End()

	shortID := c.Param("shortID")
	stats, err := service.GetURLStats(ctx, shortID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get stats")
		span.RecordError(err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("short_id", shortID))
	c.JSON(http.StatusOK, stats)
}

func HandleMetrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
