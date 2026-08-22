package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var ctx = context.Background()

// InitRedisClientSafe connects to the shared Redis (s-erp-redis) best-effort.
// Unlike InitRedisClient, it NEVER calls log.Fatal: on any failure it logs a
// warning and leaves RedisClient nil, so Redis stays an optional cache rather
// than a hard dependency. Callers must nil-check RedisClient (the internal/cache
// helpers do this for you).
func InitRedisClientSafe() {
	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			db = n
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           db,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		log.Printf("Redis unavailable at %s:%s (%v) — caching disabled",
			os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"), err)
		_ = client.Close()
		RedisClient = nil
		return
	}

	RedisClient = client
	log.Printf("Redis connected at %s:%s — caching enabled",
		os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
}

func InitRedisClient() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: func() int {
			db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
			if err != nil {
				log.Fatalf("Invalid REDIS_DB value: %v", err)
			}
			return db
		}(),
	})

	// Test the Redis connection
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func InitRedisClientTest() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST_TEST"), os.Getenv("REDIS_PORT_TEST")),
		Password: os.Getenv("REDIS_PASSWORD_TEST"),
		DB: func() int {
			db, err := strconv.Atoi(os.Getenv("REDIS_DB_TEST"))
			if err != nil {
				log.Fatalf("Invalid REDIS_DB_TEST value: %v", err)
			}
			return db
		}(),
	})

	// Test the Redis connection
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}
