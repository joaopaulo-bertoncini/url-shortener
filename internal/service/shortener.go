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
	"github.com/joaopaulo-bertoncini/url-shortener/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
	shortID := generateShortID(longURL)
	shortURL := urlPrefix + shortID

	// Redis
	if err := repository.RedisClient.Set(ctx, shortID, longURL, ttl).Err(); err != nil {
		logger.Log.Errorf("Redis SET error: %v", err)
		return "", errors.New("could not store in cache")
	}

	// MongoDB
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	doc := URLMapping{ShortID: shortID, LongURL: longURL, Created: time.Now(), AccessCount: 0}
	if _, err := collection.InsertOne(ctx, doc); err != nil {
		logger.Log.Errorf("Mongo Insert error: %v", err)
		return "", errors.New("could not store in database")
	}

	return shortURL, nil
}

func ResolveShortID(ctx context.Context, shortID string) (string, error) {
	// Redis (rápido)
	longURL, err := repository.RedisClient.Get(ctx, shortID).Result()
	if err == nil {
		incrementAccessCount(ctx, shortID)
		return longURL, nil
	}
	if err != redis.Nil {
		logger.Log.Warnf("Redis error: %v", err)
	}

	// Mongo (fallback)
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	var result URLMapping
	err = collection.FindOneAndUpdate(
		ctx,
		bson.M{"short_id": shortID},
		bson.M{"$inc": bson.M{"access_count": 1}},
	).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return "", errors.New("short URL not found")
	} else if err != nil {
		logger.Log.Errorf("Mongo error: %v", err)
		return "", errors.New("internal error")
	}

	// Reescrever cache
	_ = repository.RedisClient.Set(ctx, shortID, result.LongURL, ttl).Err()

	return result.LongURL, nil
}

func GetURLStats(ctx context.Context, shortID string) (*URLMapping, error) {
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	var result URLMapping
	err := collection.FindOne(ctx, bson.M{"short_id": shortID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("short URL not found")
	} else if err != nil {
		logger.Log.Errorf("Stats error: %v", err)
		return nil, errors.New("internal error")
	}
	return &result, nil
}

func incrementAccessCount(ctx context.Context, shortID string) {
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	_, _ = collection.UpdateOne(
		ctx,
		bson.M{"short_id": shortID},
		bson.M{"$inc": bson.M{"access_count": 1}},
	)
}

func DeleteShortID(ctx context.Context, shortID string) error {
	// Remover do Redis
	if err := repository.RedisClient.Del(ctx, shortID).Err(); err != nil && err != redis.Nil {
		logger.Log.Warnf("Redis DEL error: %v", err)
	}

	// Remover do MongoDB
	collection := repository.MongoClient.Database("shortener").Collection("urls")
	res, err := collection.DeleteOne(ctx, bson.M{"short_id": shortID})
	if err != nil {
		logger.Log.Errorf("Mongo Delete error: %v", err)
		return errors.New("failed to delete from database")
	}
	if res.DeletedCount == 0 {
		return errors.New("short URL not found")
	}

	return nil
}
