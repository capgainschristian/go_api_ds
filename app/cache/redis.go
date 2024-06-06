package cache

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

type RedisInstance struct {
	Client *redis.Client
}

var RedisClient RedisInstance

func ConnectRedis() {
	redisPassword := os.Getenv("RDB_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: redisPassword,
		DB:       0,
	})

	// Check connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")

	RedisClient = RedisInstance{
		Client: rdb,
	}
}
