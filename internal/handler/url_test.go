package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/url-shortener/internal/middleware"
	"github.com/joaopaulo-bertoncini/url-shortener/internal/repository"
	"github.com/joaopaulo-bertoncini/url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestShortenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("AUTH_TOKEN", "testtoken123")

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.POST("/shorten", HandleShorten)

	body := map[string]string{"url": "https://example.com"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer testtoken123")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["short_url"], "http://localhost:8080/")
}

func TestRedirectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/:id", HandleRedirect)

	// Pré-popula o Redis com shortID
	shortID := "abc12345"
	longURL := "https://example.com"
	_ = repository.RedisClient.Set(context.Background(), shortID, longURL, time.Hour).Err()

	req, _ := http.NewRequest(http.MethodGet, "/"+shortID, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusFound, resp.Code)
	assert.Equal(t, longURL, resp.Header().Get("Location"))
}

func TestStatsHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("AUTH_TOKEN", "testtoken123")

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/stats/:id", HandleStats)

	// Insere no Mongo direto
	shortID := "abc12345"
	doc := service.URLMapping{
		ShortID:     shortID,
		LongURL:     "https://example.com",
		Created:     time.Now(),
		AccessCount: 42,
	}
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	_, _ = collection.InsertOne(context.Background(), doc)

	req, _ := http.NewRequest(http.MethodGet, "/stats/"+shortID, nil)
	req.Header.Set("Authorization", "Bearer supersecrettoken")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response service.URLMapping
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, doc.ShortID, response.ShortID)
	assert.Equal(t, doc.LongURL, response.LongURL)
	assert.Equal(t, doc.AccessCount, response.AccessCount)
}
