// Package cache provides small, best-effort JSON helpers over the shared Redis
// (s-erp-redis). Every function is a no-op / miss when Redis is disabled or
// erroring, so callers can use it without turning Redis into a hard dependency.
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nibroos/s-erp-api/service/internal/config"
)

// GetJSON loads key into dest. Returns true only on a genuine hit; a miss, a
// disabled cache, or any Redis/unmarshal error all return false (treated as a
// miss so the caller falls back to the source of truth).
func GetJSON(ctx context.Context, key string, dest interface{}) bool {
	if config.RedisClient == nil {
		return false
	}
	data, err := config.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return false // redis.Nil (miss) or a transient error
	}
	if json.Unmarshal(data, dest) != nil {
		return false
	}
	return true
}

// SetJSON caches val under key with a TTL. Best-effort: errors are swallowed.
func SetJSON(ctx context.Context, key string, val interface{}, ttl time.Duration) {
	if config.RedisClient == nil {
		return
	}
	data, err := json.Marshal(val)
	if err != nil {
		return
	}
	_ = config.RedisClient.Set(ctx, key, data, ttl).Err()
}

// Del removes keys (e.g. to invalidate on write). Best-effort.
func Del(ctx context.Context, keys ...string) {
	if config.RedisClient == nil || len(keys) == 0 {
		return
	}
	_ = config.RedisClient.Del(ctx, keys...).Err()
}
