package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	RedisClient *redis.Client
)

func InitClients() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}
	MongoClient = client
	log.Println("✅ Connected to MongoDB")

	// Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("✅ Connected to Redis")

	return nil
}
