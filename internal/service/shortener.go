package service

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joaopaulo-bertoncini/url-shortener/internal/logger"
	"github.com/joaopaulo-bertoncini/url-shortener/internal/metrics"
	"github.com/joaopaulo-bertoncini/url-shortener/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var tracer = otel.Tracer("url-shortener/service")

const (
	shortIDLength = 8
	ttl           = 24 * time.Hour
)

var urlPrefix = getURLPrefix()

func getURLPrefix() string {
	if v := os.Getenv("URL_PREFIX"); v != "" {
		return v
	}
	return "http://localhost:8080/"
}

type URLMapping struct {
	ShortID     string    `bson:"short_id" json:"short_id"`
	LongURL     string    `bson:"long_url" json:"long_url"`
	Created     time.Time `bson:"created_at" json:"created_at"`
	AccessCount int       `bson:"access_count" json:"access_count"`
}

func generateShortID(longURL string) string {
	hash := sha1.Sum([]byte(longURL + fmt.Sprint(time.Now().UnixNano())))
	return base64.URLEncoding.EncodeToString(hash[:])[:shortIDLength]
}

func ShortenURL(ctx context.Context, longURL string) (string, error) {
	ctx, span := tracer.Start(ctx, "ShortenURL")
	defer span.End()

	shortID := generateShortID(longURL)
	shortURL := urlPrefix + shortID
	// Redis SET
	start := time.Now()
	err := repository.RedisClient.Set(ctx, shortID, longURL, ttl).Err()
	metrics.RedisOpDuration.WithLabelValues("SET").Observe(time.Since(start).Seconds())
	if err != nil {
		span.SetStatus(codes.Error, "failed to save mapping")
		span.RecordError(err)
		logger.Log.Errorf("Redis SET error: %v", err)
		return "", errors.New("could not store in cache")
	}

	// Mongo INSERT
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	doc := URLMapping{ShortID: shortID, LongURL: longURL, Created: time.Now(), AccessCount: 0}
	start = time.Now()
	_, err = collection.InsertOne(ctx, doc)
	metrics.MongoOpDuration.WithLabelValues("InsertOne").Observe(time.Since(start).Seconds())
	if err != nil {
		span.SetStatus(codes.Error, "failed to save mapping")
		span.RecordError(err)
		logger.Log.Errorf("Mongo Insert error: %v", err)
		return "", errors.New("could not store in database")
	}

	return shortURL, nil
}

func ResolveShortID(ctx context.Context, shortID string) (string, error) {
	ctx, span := tracer.Start(ctx, "ResolveShortID")
	defer span.End()
	// Redis GET
	start := time.Now()
	longURL, err := repository.RedisClient.Get(ctx, shortID).Result()
	metrics.RedisOpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())

	if err == nil {
		metrics.RedisCacheHits.Inc()
		incrementAccessCount(ctx, shortID)
		metrics.RedirectCounter.Inc()
		return longURL, nil
	}
	if err != redis.Nil {
		span.SetStatus(codes.Error, "failed to resolve short ID")
		span.RecordError(err)
		logger.Log.Warnf("Redis error: %v", err)
	}
	metrics.RedisCacheMisses.Inc()

	// Mongo fallback
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	var result URLMapping
	start = time.Now()
	err = collection.FindOneAndUpdate(
		ctx,
		bson.M{"short_id": shortID},
		bson.M{"$inc": bson.M{"access_count": 1}},
	).Decode(&result)
	metrics.MongoOpDuration.WithLabelValues("FindOneAndUpdate").Observe(time.Since(start).Seconds())

	if err == mongo.ErrNoDocuments {
		span.SetStatus(codes.Error, "short URL not found")
		span.RecordError(err)
		return "", errors.New("short URL not found")
	} else if err != nil {
		span.SetStatus(codes.Error, "failed to resolve short ID")
		span.RecordError(err)
		logger.Log.Errorf("Mongo error: %v", err)
		return "", errors.New("internal error")
	}

	// Reescreve no cache
	start = time.Now()
	_ = repository.RedisClient.Set(ctx, shortID, result.LongURL, ttl).Err()
	metrics.RedisOpDuration.WithLabelValues("SET").Observe(time.Since(start).Seconds())

	return result.LongURL, nil
}

func GetURLStats(ctx context.Context, shortID string) (*URLMapping, error) {
	ctx, span := tracer.Start(ctx, "GetURLStats")
	defer span.End()
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	var result URLMapping
	err := collection.FindOne(ctx, bson.M{"short_id": shortID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		span.SetStatus(codes.Error, "short URL not found")
		span.RecordError(err)
		return nil, errors.New("short URL not found")
	} else if err != nil {
		span.SetStatus(codes.Error, "failed to get URL stats")
		span.RecordError(err)
		logger.Log.Errorf("Stats error: %v", err)
		return nil, errors.New("internal error")
	}
	return &result, nil
}

func incrementAccessCount(ctx context.Context, shortID string) {
	ctx, span := tracer.Start(ctx, "incrementAccessCount")
	defer span.End()
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	_, _ = collection.UpdateOne(
		ctx,
		bson.M{"short_id": shortID},
		bson.M{"$inc": bson.M{"access_count": 1}},
	)
}

func DeleteShortID(ctx context.Context, shortID string) error {
	ctx, span := tracer.Start(ctx, "DeleteShortID")
	defer span.End()
	// Redis DEL
	start := time.Now()
	err := repository.RedisClient.Del(ctx, shortID).Err()
	metrics.RedisOpDuration.WithLabelValues("DEL").Observe(time.Since(start).Seconds())
	if err != nil && err != redis.Nil {
		span.SetStatus(codes.Error, "failed to delete from cache")
		span.RecordError(err)
		logger.Log.Warnf("Redis DEL error: %v", err)
	}

	// Mongo DELETE
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	start = time.Now()
	res, err := collection.DeleteOne(ctx, bson.M{"short_id": shortID})
	metrics.MongoOpDuration.WithLabelValues("DeleteOne").Observe(time.Since(start).Seconds())
	if err != nil {
		span.SetStatus(codes.Error, "failed to delete from database")
		span.RecordError(err)
		logger.Log.Errorf("Mongo Delete error: %v", err)
		return errors.New("failed to delete from database")
	}
	if res.DeletedCount == 0 {
		span.SetStatus(codes.Error, "short URL not found")
		span.RecordError(errors.New("short URL not found"))
		return errors.New("short URL not found")
	}

	return nil
}
