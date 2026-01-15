package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

// InitRedis initializes the Redis client
func InitRedis(redisURL string) error {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}

	rdb = redis.NewClient(opt)

	// Test connection
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}

	log.Println("Redis connected successfully")
	return nil
}

// Get retrieves a value from cache and unmarshals it into the provided interface
func Get(key string, dest interface{}) error {
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key doesn't exist
		return nil
	} else if err != nil {
		return err
	}

	// Unmarshal JSON into destination
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with the given TTL
func Set(key string, value interface{}, ttl time.Duration) error {
	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key from cache
func Delete(key string) error {
	return rdb.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache
func Exists(key string) (bool, error) {
	result, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetClient returns the Redis client instance
func GetClient() *redis.Client {
	return rdb
}

// Close closes the Redis connection
func Close() error {
	if rdb != nil {
		return rdb.Close()
	}
	return nil
}
