package cache

import (
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{Client: client}
}

// func (r *RedisCache) FetchAndCacheSubscribes(ctx context.Context, subscribes []dtos.SubscribeListDTO) error {
// 	// Marshal the data to JSON
// 	data, err := json.Marshal(subscribes)
// 	if err != nil {
// 		return err
// 	}

// 	// Store the data in Redis
// 	err = r.Client.Set(ctx, "subscribes", data, 0).Err()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
