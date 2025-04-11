// internal/handler/url.go
package handler

import (
	"net/http"

	"github.com/joaopaulo-bertoncini/url-shortener/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
)

type reqBody struct {
	URL string `json:"url" binding:"required,url"`
}

func HandleShorten(c *gin.Context) {
	var body reqBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	shortURL, err := service.ShortenURL(c.Request.Context(), body.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ShortenCounter.Inc()
	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
}

func HandleRedirect(c *gin.Context) {
	shortID := c.Param("shortID")
	longURL, err := service.ResolveShortID(c.Request.Context(), shortID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	RedirectCounter.Inc()
	c.Redirect(http.StatusMovedPermanently, longURL)
}

func HandleDelete(c *gin.Context) {
	shortID := c.Param("shortID")

	err := service.DeleteShortID(c.Request.Context(), shortID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "short URL deleted successfully"})
}

func HandleStats(c *gin.Context) {
	shortID := c.Param("shortID")
	stats, err := service.GetURLStats(c.Request.Context(), shortID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func HandleMetrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
